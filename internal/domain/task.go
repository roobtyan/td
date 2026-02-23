package domain

import "time"

type Task struct {
	ID        int64
	Title     string
	Notes     string
	Status    Status
	Project   string
	Priority  string
	DueAt     *time.Time
	DoneAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
