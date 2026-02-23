package domain

type Status string

const (
	StatusInbox   Status = "inbox"
	StatusTodo    Status = "todo"
	StatusDoing   Status = "doing"
	StatusDone    Status = "done"
	StatusDeleted Status = "deleted"
)

func ParseStatus(raw string) (Status, error) {
	s := Status(raw)
	if !IsValidStatus(s) {
		return "", ErrInvalidStatus
	}
	return s, nil
}

func IsValidStatus(s Status) bool {
	switch s {
	case StatusInbox, StatusTodo, StatusDoing, StatusDone, StatusDeleted:
		return true
	default:
		return false
	}
}

func CanTransit(from, to Status) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusInbox:
		return to == StatusTodo || to == StatusDoing || to == StatusDone || to == StatusDeleted
	case StatusTodo:
		return to == StatusDoing || to == StatusDone || to == StatusDeleted
	case StatusDoing:
		return to == StatusTodo || to == StatusDone || to == StatusDeleted
	case StatusDone:
		return to == StatusTodo || to == StatusDeleted
	case StatusDeleted:
		return to == StatusTodo
	default:
		return false
	}
}
