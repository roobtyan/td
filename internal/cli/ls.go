package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
	"td/internal/domain"
)

func newLsCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls [today]",
		Short: "List tasks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			var tasks []domain.Task
			if len(args) == 0 {
				uc := usecase.ListTaskUseCase{Repo: repo}
				tasks, err = uc.Execute(cmd.Context(), usecase.ListTaskInput{})
				if err != nil {
					return err
				}
				tasks = filterOutDeleted(tasks)
			} else {
				view := strings.ToLower(strings.TrimSpace(args[0]))
				switch view {
				case string(domain.ViewToday):
					queryUC := usecase.NewNavQueryUseCase(repo)
					tasks, err = queryUC.ListByView(cmd.Context(), domain.ViewToday, time.Now().Local(), "", false)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("unsupported ls view %q, only today is supported", args[0])
				}
			}
			if len(args) == 0 {
				sortTasksForLS(tasks)
			}
			for _, task := range tasks {
				cmd.Println(formatTaskLine(task.ID, string(task.Status), task.Title, task.Project, task.DueAt, task.Priority))
			}
			return nil
		},
	}
	return cmd
}

const (
	lsIDWidth       = 5
	lsStatusWidth   = 8
	lsTitleWidth    = 34
	lsProjectWidth  = 16
	lsDueWidth      = 16
	lsPriorityWidth = 3
)

func formatTaskLine(id int64, status, title, project string, dueAt *time.Time, priority string) string {
	parts := []string{
		padLSField(strconv.FormatInt(id, 10), lsIDWidth),
		padLSField("["+status+"]", lsStatusWidth),
		padLSField(title, lsTitleWidth),
		padLSField(formatProject(project), lsProjectWidth),
		padLSField(formatDue(dueAt), lsDueWidth),
		padLSField(formatPriority(priority), lsPriorityWidth),
	}
	return strings.TrimRight(strings.Join(parts, " "), " ")
}

func padLSField(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = flattenLSField(text)
	tail := ""
	if width > 3 {
		tail = "..."
	}
	text = ansi.Truncate(text, width, tail)
	w := ansi.StringWidth(text)
	if w >= width {
		return text
	}
	return text + strings.Repeat(" ", width-w)
}

func flattenLSField(text string) string {
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return text
}

func formatProject(project string) string {
	project = strings.TrimSpace(project)
	if project == "" {
		return "-"
	}
	return project
}

func formatDue(dueAt *time.Time) string {
	if dueAt == nil {
		return "-"
	}
	return dueAt.In(time.Local).Format("2006-01-02 15:04")
}

func formatPriority(priority string) string {
	priority = domain.NormalizePriority(priority)
	if !domain.IsValidPriority(priority) {
		return domain.DefaultPriority
	}
	return priority
}

func filterOutDeleted(tasks []domain.Task) []domain.Task {
	out := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == domain.StatusDeleted {
			continue
		}
		out = append(out, task)
	}
	return out
}

func sortTasksForLS(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		left := tasks[i]
		right := tasks[j]

		lp := strings.TrimSpace(left.Project)
		rp := strings.TrimSpace(right.Project)
		if lp != rp {
			if lp == "" {
				return false
			}
			if rp == "" {
				return true
			}
			return lp < rp
		}

		ls := lsStatusRank(left.Status)
		rs := lsStatusRank(right.Status)
		if ls != rs {
			return ls < rs
		}
		return left.ID < right.ID
	})
}

func lsStatusRank(status domain.Status) int {
	switch status {
	case domain.StatusDoing:
		return 0
	case domain.StatusTodo:
		return 1
	case domain.StatusDone:
		return 2
	case domain.StatusInbox:
		return 3
	case domain.StatusDeleted:
		return 4
	default:
		return 5
	}
}
