package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectPosition struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Code      string
	Name      string
	Capacity  int
	CreatedAt time.Time
}
