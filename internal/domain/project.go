package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectDraft       ProjectStatus = "DRAFT"
	ProjectReview      ProjectStatus = "REVIEW"
	ProjectRecruitment ProjectStatus = "RECRUITMENT"
	ProjectActive      ProjectStatus = "ACTIVE"
	ProjectGrading     ProjectStatus = "GRADING"
	ProjectArchive     ProjectStatus = "ARCHIVE"
)

type Project struct {
	ID                    uuid.UUID
	Title                 string
	Description           string
	Status                ProjectStatus
	IsPublic              bool
	CreatedBy             uuid.UUID
	CreatedByName         string
	CreatedByEmail        string
	ProfessorID           *uuid.UUID
	ProfessorReviewStatus string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	FacultyID             uuid.UUID
	Visibility            string
	GroupID               *uuid.UUID
	ImageKey              string
	ImageURL              string
	DefaultCoverVariant   int
	ImageUpdatedAt        *time.Time
}
