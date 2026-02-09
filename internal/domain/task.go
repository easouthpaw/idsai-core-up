package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskOpen       TaskStatus = "OPEN"
	TaskInProgress TaskStatus = "IN_PROGRESS"
	TaskDone       TaskStatus = "DONE"
)

type Task struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Title          string
	Description    string
	PositionID     uuid.UUID
	AssigneeUserID *uuid.UUID
	Status         TaskStatus
	CreatedBy      uuid.UUID
	DueAt          *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
