package domain

import (
	"fmt"
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
	ProjectCompleted   ProjectStatus = "COMPLETED"
	ProjectArchive     ProjectStatus = "ARCHIVE"
)

const (
	ProjectRetakePenaltyPerAttempt = 5
	ProjectRetakePenaltyCap        = 25
)

func RetakePenaltyPercent(retakeCount int) int {
	if retakeCount <= 0 {
		return 0
	}
	penalty := retakeCount * ProjectRetakePenaltyPerAttempt
	if penalty > ProjectRetakePenaltyCap {
		return ProjectRetakePenaltyCap
	}
	return penalty
}

func ReviewScoreFromPercent(passPercent int) string {
	if passPercent < 0 {
		passPercent = 0
	}
	if passPercent > 100 {
		passPercent = 100
	}
	return fmt.Sprintf("%.1f", float64(passPercent)*5/100)
}

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
	RetakeCount           int
	ImageUpdatedAt        *time.Time
}
