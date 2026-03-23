package projects

import (
	"time"

	"idsai-core-up/internal/domain"
)

type ViewerAccess struct {
	CanViewWorkspace  bool
	CanApply          bool
	CanViewFinalGrade bool
}

type ReviewSummary struct {
	Score       string
	PassPercent int
	Met         int
	Total       int
	ReviewedAt  *time.Time
	Reviewer    string
}

type ProjectView struct {
	Project       domain.Project
	Access        ViewerAccess
	ReviewSummary *ReviewSummary
}
