package projectflow

import (
	"context"
	"errors"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

func (s *Service) SearchProfessors(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]ProfessorCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.invite_professor", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 50)
	return s.professorsRepo.ListProfessorCandidates(ctx, p.FacultyID, term, limit, userID, p.CreatedBy)
}

func (s *Service) AssignProfessor(ctx context.Context, userID, projectID, professorID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.invite_professor", projectID); err != nil {
		return domain.Project{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if professorID == userID || professorID == p.CreatedBy {
		return domain.Project{}, ErrInvalidInput
	}

	professorOK, err := s.professorsRepo.IsActiveProfessorInFaculty(ctx, professorID, p.FacultyID)
	if err != nil {
		return domain.Project{}, err
	}
	if !professorOK {
		return domain.Project{}, ErrInvalidInput
	}

	if err := s.professorsRepo.AssignProjectProfessor(ctx, projectID, professorID); err != nil {
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) GetAssignedProfessor(ctx context.Context, userID, projectID uuid.UUID) (*ProfessorCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.view", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.ProfessorID == nil {
		return nil, nil
	}

	item, err := s.professorsRepo.GetProfessorCandidateByID(ctx, *p.ProfessorID, p.FacultyID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *Service) RespondProfessorInvite(ctx context.Context, professorID, projectID uuid.UUID, accept bool) (domain.Project, error) {
	p, err := s.professorsRepo.RespondProfessorInvite(ctx, projectID, professorID, accept)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Project{}, ErrInviteNotFound
		}
		return domain.Project{}, err
	}

	if accept {
		if err := s.ensureProjectRole(ctx, professorID, "PROJECT_PROFESSOR", projectID); err != nil {
			return domain.Project{}, err
		}
	}
	return p, nil
}

func (s *Service) ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, query string, limit int) ([]domain.Project, error) {
	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 100)
	return s.professorsRepo.ListProfessorReviewInvites(ctx, professorID, term, limit)
}

func (s *Service) ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]IncomingInvite, error) {
	limit = clampLimit(limit, 100)
	return s.membersRepo.ListIncomingInvites(ctx, userID, limit)
}

func (s *Service) ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]OutgoingApplication, error) {
	limit = clampLimit(limit, 100)
	return s.membersRepo.ListOutgoingApplications(ctx, userID, limit)
}
