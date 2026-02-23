package domain

import (
	"errors"
	"fmt"
)

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidStatus = errors.New("invalid status")
)

type InvalidTransitionError struct {
	From Status
	To   Status
}

func (e InvalidTransitionError) Error() string {
	return fmt.Sprintf("invalid status transition: %s -> %s", e.From, e.To)
}

func NewInvalidTransitionError(from, to Status) error {
	return InvalidTransitionError{From: from, To: to}
}
