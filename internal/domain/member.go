package domain

import (
	"time"

	"github.com/google/uuid"
)

type MemberStatus string

const (
	MemberApplied  MemberStatus = "APPLIED"
	MemberInvited  MemberStatus = "INVITED"
	MemberActive   MemberStatus = "ACTIVE"
	MemberRejected MemberStatus = "REJECTED"
	MemberRemoved  MemberStatus = "REMOVED"
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
