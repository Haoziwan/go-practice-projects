package task

import (
	"time"
)

// Task represents a single todo item
type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	CompletedAt *time.Time // nil if not completed, timestamp if completed
}

// IsComplete returns whether the task has been completed
func (t *Task) IsComplete() bool {
	return t.CompletedAt != nil
}

// Complete marks the task as completed with current timestamp
func (t *Task) Complete() {
	now := time.Now()
	t.CompletedAt = &now
}
