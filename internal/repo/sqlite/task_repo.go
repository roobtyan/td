package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"td/internal/domain"
	"td/internal/repo"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

var _ repo.TaskRepository = (*TaskRepository)(nil)

func (r *TaskRepository) Create(ctx context.Context, task domain.Task) (int64, error) {
	status := task.Status
	if status == "" {
		status = domain.StatusInbox
	}
	if !domain.IsValidStatus(status) {
		return 0, domain.ErrInvalidStatus
	}
	priority := task.Priority
	if priority == "" {
		priority = "P2"
	}

	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tasks(title, notes, status, project, priority)
		 VALUES(?, ?, ?, ?, ?)`,
		task.Title, task.Notes, string(status), task.Project, priority,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	var (
		task      domain.Task
		rawStatus string
		doneAt    sql.NullTime
	)
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, title, notes, status, project, priority, done_at, created_at, updated_at
		   FROM tasks
		  WHERE id = ?`,
		id,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Notes,
		&rawStatus,
		&task.Project,
		&task.Priority,
		&doneAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}

	status, err := domain.ParseStatus(rawStatus)
	if err != nil {
		return domain.Task{}, err
	}
	task.Status = status
	if doneAt.Valid {
		t := doneAt.Time.UTC()
		task.DoneAt = &t
	}
	return task, nil
}

func (r *TaskRepository) MarkDone(ctx context.Context, ids []int64) error {
	return r.transit(ctx, ids, domain.StatusDone)
}

func (r *TaskRepository) Reopen(ctx context.Context, ids []int64) error {
	return r.transit(ctx, ids, domain.StatusTodo)
}

func (r *TaskRepository) transit(ctx context.Context, ids []int64, to domain.Status) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		from, err := r.statusByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if !domain.CanTransit(from, to) {
			return domain.NewInvalidTransitionError(from, to)
		}
		var doneAt any
		if to == domain.StatusDone {
			doneAt = time.Now().UTC()
		} else {
			doneAt = nil
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE tasks
			    SET status = ?, done_at = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`,
			string(to), doneAt, id,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TaskRepository) statusByID(ctx context.Context, tx *sql.Tx, id int64) (domain.Status, error) {
	var rawStatus string
	err := tx.QueryRowContext(
		ctx,
		`SELECT status FROM tasks WHERE id = ?`,
		id,
	).Scan(&rawStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrTaskNotFound
	}
	if err != nil {
		return "", err
	}
	return domain.ParseStatus(rawStatus)
}
