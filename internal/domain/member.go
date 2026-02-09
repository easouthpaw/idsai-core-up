package domain

import (
	"time"

	"github.com/google/uuid"
)

type MemberStatus string

const (
	MemberApplied MemberStatus = "APPLIED"
	MemberActive  MemberStatus = "ACTIVE"
	MemberRemoved MemberStatus = "REMOVED"
)

type ProjectMember struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	UserID     uuid.UUID
	PositionID *uuid.UUID
	Status     MemberStatus
	JoinedAt   *time.Time
	CreatedAt  time.Time
}
