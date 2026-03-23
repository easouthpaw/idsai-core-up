package projectflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

func (s *Service) CreateCriterion(ctx context.Context, userID, projectID uuid.UUID, title, description string, weight int) (Criterion, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.set_criteria", projectID); err != nil {
		return Criterion{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return Criterion{}, ErrInvalidInput
	}
	if weight <= 0 {
		weight = 1
	}
	if weight > 100 {
		weight = 100
	}
	currentWeight, err := s.criteriaRepo.GetProjectCriteriaWeightSum(ctx, projectID)
	if err != nil {
		return Criterion{}, err
	}
	if currentWeight+weight > 100 {
		return Criterion{}, fmt.Errorf("%w: total criteria weight exceeds 100", ErrInvalidInput)
	}

	return s.criteriaRepo.CreateProjectCriterion(ctx, projectID, userID, title, description, weight)
}

func (s *Service) ListCriteria(ctx context.Context, projectID uuid.UUID) ([]Criterion, error) {
	return s.criteriaRepo.ListProjectCriteria(ctx, projectID)
}

func (s *Service) GetGrading(ctx context.Context, userID, projectID uuid.UUID) ([]CriterionGrade, error) {
	if err := s.requireProjectPermission(ctx, userID, "grading.view", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	professorID := userID
	if p.ProfessorID != nil {
		professorID = *p.ProfessorID
	}
	items, err := s.criteriaRepo.ListProjectCriterionGrades(ctx, projectID, professorID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) UpsertGrading(ctx context.Context, userID, projectID uuid.UUID, items []CriterionGrade) ([]CriterionGrade, error) {
	if err := s.requireProjectPermission(ctx, userID, "grading.mark_criteria", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.ProjectReview && p.Status != domain.ProjectGrading {
		return nil, ErrInvalidInput
	}

	uniq := make(map[uuid.UUID]CriterionGradeUpsert, len(items))
	for _, item := range items {
		cid, err := uuid.Parse(strings.TrimSpace(item.CriterionID))
		if err != nil {
			return nil, ErrInvalidInput
		}
		comment := strings.TrimSpace(item.Comment)
		if len(comment) > 3000 {
			comment = comment[:3000]
		}
		uniq[cid] = CriterionGradeUpsert{
			CriterionID: cid,
			IsMet:       item.IsMet,
			Comment:     comment,
		}
	}

	sanitized := make([]CriterionGradeUpsert, 0, len(uniq))
	for _, item := range uniq {
		sanitized = append(sanitized, item)
	}
	if err := s.criteriaRepo.UpsertProjectCriterionGrades(ctx, projectID, userID, sanitized); err != nil {
		if errors.Is(err, ErrSchemaMissing) {
			return nil, ErrInvalidInput
		}
		if errors.Is(err, ErrInvalidInput) {
			return nil, ErrInvalidInput
		}
		return nil, err
	}
	return s.GetGrading(ctx, userID, projectID)
}

func (s *Service) Readiness(ctx context.Context, projectID uuid.UUID) (Readiness, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Readiness{}, err
	}

	members, err := s.membersRepo.ListProjectMembers(ctx, projectID)
	if err != nil {
		return Readiness{}, err
	}
	requiredMembers := 0
	activeMembers := 0
	for _, m := range members {
		if strings.ToUpper(strings.TrimSpace(m.Status)) != "ACTIVE" {
			continue
		}
		if strings.TrimSpace(m.UserID) == p.CreatedBy.String() {
			continue
		}
		requiredMembers++
		if m.PositionID != nil && strings.TrimSpace(*m.PositionID) != "" {
			activeMembers++
		}
	}

	criteriaCount, err := s.criteriaRepo.CountProjectCriteria(ctx, projectID)
	if err != nil {
		return Readiness{}, err
	}

	hasProfessor := p.ProfessorID != nil
	professorStatus := strings.ToUpper(strings.TrimSpace(p.ProfessorReviewStatus))
	if professorStatus == "" {
		professorStatus = "NONE"
	}
	professorAccepted := professorStatus == "ACCEPTED"
	canActivate := requiredMembers > 0 && activeMembers >= requiredMembers && professorAccepted && criteriaCount > 0

	return Readiness{
		ProjectID:       projectID.String(),
		Status:          string(p.Status),
		RequiredMembers: requiredMembers,
		ActiveMembers:   activeMembers,
		HasProfessor:    hasProfessor,
		ProfessorStatus: professorStatus,
		CriteriaCount:   criteriaCount,
		CanActivate:     canActivate,
	}, nil
}

func (s *Service) ApproveProject(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, Readiness, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.approve", projectID); err != nil {
		return domain.Project{}, Readiness{}, err
	}
	ready, err := s.Readiness(ctx, projectID)
	if err != nil {
		return domain.Project{}, Readiness{}, err
	}
	if !ready.CanActivate {
		return domain.Project{}, ready, fmt.Errorf("%w: members=%d/%d professor=%s criteria=%d", ErrProjectNotReady, ready.ActiveMembers, ready.RequiredMembers, ready.ProfessorStatus, ready.CriteriaCount)
	}

	if err := s.lifecycleRepo.ActivateProject(ctx, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Project{}, Readiness{}, ErrInvalidInput
		}
		return domain.Project{}, Readiness{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	return p, ready, err
}

func (s *Service) SubmitProjectForGrading(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectSubmitAccess(ctx, userID, projectID); err != nil {
		return domain.Project{}, err
	}

	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.Status != domain.ProjectActive {
		return domain.Project{}, fmt.Errorf("%w: project status must be ACTIVE", ErrInvalidInput)
	}
	if p.ProfessorID == nil {
		return domain.Project{}, fmt.Errorf("%w: professor is not assigned", ErrInvalidInput)
	}
	if strings.ToUpper(strings.TrimSpace(p.ProfessorReviewStatus)) != "ACCEPTED" {
		return domain.Project{}, fmt.Errorf("%w: professor invitation is not accepted", ErrInvalidInput)
	}

	tasksTotal, tasksDone, err := s.lifecycleRepo.CountProjectTasksSummary(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if tasksTotal == 0 {
		return domain.Project{}, fmt.Errorf("%w: at least one task is required", ErrInvalidInput)
	}
	if tasksDone < tasksTotal {
		return domain.Project{}, fmt.Errorf("%w: all tasks must be DONE before grading", ErrInvalidInput)
	}

	if err := s.lifecycleRepo.MoveProjectToGrading(ctx, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Project{}, ErrInvalidInput
		}
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) PublishGrading(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "grading.publish", projectID); err != nil {
		return domain.Project{}, err
	}

	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.Status != domain.ProjectGrading {
		return domain.Project{}, fmt.Errorf("%w: project status must be GRADING", ErrInvalidInput)
	}
	if p.ProfessorID == nil || *p.ProfessorID != userID {
		return domain.Project{}, domain.ErrForbidden
	}

	criteriaTotal, err := s.criteriaRepo.CountProjectCriteria(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if criteriaTotal == 0 {
		return domain.Project{}, fmt.Errorf("%w: criteria are not configured", ErrInvalidInput)
	}

	gradedTotal, err := s.criteriaRepo.CountProjectGradedCriteria(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrSchemaMissing) {
			return domain.Project{}, fmt.Errorf("%w: grading table is missing", ErrInvalidInput)
		}
		return domain.Project{}, err
	}
	if gradedTotal < criteriaTotal {
		return domain.Project{}, fmt.Errorf("%w: grading is incomplete (%d/%d)", ErrInvalidInput, gradedTotal, criteriaTotal)
	}

	if err := s.lifecycleRepo.MoveProjectToArchive(ctx, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Project{}, ErrInvalidInput
		}
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	return s.lifecycleRepo.DeleteOwnedProject(ctx, projectID, userID)
}
