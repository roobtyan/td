package domain

import "time"

type Task struct {
	ID        int64
	Title     string
	Notes     string
	Status    string
	Project   string
	Priority  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
