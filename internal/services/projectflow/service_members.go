package projectflow

import (
	"context"
	"errors"
	"strings"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/strutil"

	"github.com/google/uuid"
)

func (s *Service) UpdateProject(ctx context.Context, userID, projectID uuid.UUID, title, description *string) (domain.Project, error) {
	if title == nil && description == nil {
		return domain.Project{}, ErrInvalidInput
	}
	if err := s.requireProjectEditAccess(ctx, userID, projectID); err != nil {
		return domain.Project{}, err
	}

	titleVal := ""
	titleSet := false
	if title != nil {
		titleSet = true
		titleVal = strings.TrimSpace(*title)
		if titleVal == "" {
			return domain.Project{}, ErrInvalidInput
		}
	}
	descVal := ""
	descSet := false
	if description != nil {
		descSet = true
		descVal = strings.TrimSpace(*description)
	}

	if err := s.projectsRepo.UpdateProject(ctx, projectID, titleSet, titleVal, descSet, descVal); err != nil {
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) SetStacks(ctx context.Context, userID, projectID uuid.UUID, stacks []string) ([]Stack, error) {
	if err := s.requireProjectEditAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	norm := normalizeStackCodes(stacks)

	if err := s.stacksRepo.ReplaceProjectStacks(ctx, projectID, norm); err != nil {
		if errors.Is(err, ErrSchemaMissing) {
			return stacksFromCodes(norm), nil
		}
		return nil, err
	}

	return s.ListStacks(ctx, projectID)
}

func (s *Service) ListStacks(ctx context.Context, projectID uuid.UUID) ([]Stack, error) {
	codes, err := s.stacksRepo.ListProjectStackCodes(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return stacksFromCodes(codes), nil
}

func (s *Service) OpenRecruitment(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err != nil {
		return domain.Project{}, err
	}
	if err := s.projectsRepo.OpenProjectRecruitment(ctx, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Project{}, ErrInvalidInput
		}
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) CreatePosition(ctx context.Context, userID, projectID uuid.UUID, code, name string, capacity int) (Position, error) {
	if err := s.requireProjectPermission(ctx, userID, "position.create", projectID); err != nil {
		return Position{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Position{}, ErrInvalidInput
	}
	if capacity <= 0 {
		capacity = 1
	}
	code = normalizePositionCode(code, name)
	if code == "" {
		return Position{}, ErrInvalidInput
	}

	return s.positionsRepo.CreateProjectPosition(ctx, projectID, code, name, capacity)
}

func (s *Service) ListPositions(ctx context.Context, projectID uuid.UUID) ([]Position, error) {
	return s.positionsRepo.ListProjectPositions(ctx, projectID)
}

func (s *Service) ListStudentCandidates(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]StudentCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 100)
	return s.projectsRepo.ListStudentCandidates(ctx, p.FacultyID, projectID, userID, p.CreatedBy, term, limit)
}

func (s *Service) InviteStudent(ctx context.Context, userID, projectID, studentID uuid.UUID, comment string) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}
	if studentID == userID || studentID == p.CreatedBy {
		return Member{}, ErrInvalidInput
	}

	studentOK, err := s.membersRepo.IsActiveStudentInFaculty(ctx, studentID, p.FacultyID)
	if err != nil {
		return Member{}, err
	}
	if !studentOK {
		return Member{}, ErrInvalidInput
	}

	comment = strings.TrimSpace(comment)
	comment = strutil.TruncateUTF8(comment, 500)

	m, err := s.membersRepo.UpsertInvitedMember(ctx, projectID, studentID, userID, comment)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, ErrInvalidInput
		}
		return Member{}, err
	}
	if err := s.ensureProjectRole(ctx, studentID, "INVITED_MEMBER", projectID); err != nil {
		return Member{}, err
	}
	return m, nil
}

func (s *Service) ApplyMember(ctx context.Context, userID, projectID uuid.UUID, comment string) (Member, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if err := s.requireFacultyPermission(ctx, userID, "member.apply", p.FacultyID); err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}

	comment = strings.TrimSpace(comment)
	comment = strutil.TruncateUTF8(comment, 500)

	m, err := s.membersRepo.UpsertAppliedMember(ctx, projectID, userID, comment)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, ErrInvalidInput
		}
		return Member{}, err
	}
	return m, nil
}

func (s *Service) ListMembers(ctx context.Context, projectID uuid.UUID) ([]Member, error) {
	return s.membersRepo.ListProjectMembers(ctx, projectID)
}

func (s *Service) ApproveMember(ctx context.Context, userID, projectID, memberUserID uuid.UUID, positionID *uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}
	if positionID != nil {
		if err := s.ensurePositionCapacity(ctx, projectID, *positionID, &memberUserID); err != nil {
			return Member{}, err
		}
	}

	m, err := s.membersRepo.ApproveProjectMember(ctx, projectID, memberUserID, positionID)
	if err != nil {
		return Member{}, err
	}
	if err := s.ensureProjectRole(ctx, memberUserID, "MEMBER", projectID); err != nil {
		return Member{}, err
	}
	return m, nil
}

func (s *Service) RejectMemberApplication(ctx context.Context, userID, projectID, memberUserID uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}

	m, err := s.membersRepo.RejectProjectMemberApplication(ctx, projectID, memberUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, ErrInviteNotFound
		}
		return Member{}, err
	}
	if err := s.revokeProjectRole(ctx, memberUserID, "INVITED_MEMBER", projectID); err != nil {
		return Member{}, err
	}
	if err := s.revokeProjectRole(ctx, memberUserID, "MEMBER", projectID); err != nil {
		return Member{}, err
	}
	return m, nil
}

func (s *Service) RemoveMember(ctx context.Context, userID, projectID, memberUserID uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if memberUserID == p.CreatedBy {
		return Member{}, ErrInvalidInput
	}

	m, err := s.membersRepo.RemoveProjectMember(ctx, projectID, memberUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, ErrNotFound
		}
		return Member{}, err
	}

	for _, roleCode := range []string{"INVITED_MEMBER", "MEMBER", "CO_LEAD", "RECRUITER", "TASK_MANAGER"} {
		if err := s.revokeProjectRole(ctx, memberUserID, roleCode, projectID); err != nil {
			return Member{}, err
		}
	}

	return m, nil
}

func (s *Service) SetMemberPosition(ctx context.Context, userID, projectID, memberUserID, positionID uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	if err := s.ensurePositionCapacity(ctx, projectID, positionID, &memberUserID); err != nil {
		return Member{}, err
	}
	return s.membersRepo.SetActiveMemberPosition(ctx, projectID, memberUserID, positionID)
}

func (s *Service) RespondMemberInvite(ctx context.Context, userID, projectID uuid.UUID, accept bool) (Member, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if err := s.requireFacultyPermission(ctx, userID, "member.apply", p.FacultyID); err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}

	if accept {
		invitePositionID, err := s.membersRepo.GetInvitedMemberPosition(ctx, projectID, userID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return Member{}, ErrInviteNotFound
			}
			return Member{}, err
		}
		if invitePositionID != nil {
			if err := s.ensurePositionCapacity(ctx, projectID, *invitePositionID, &userID); err != nil {
				return Member{}, err
			}
		}
	}

	m, err := s.membersRepo.RespondMemberInvite(ctx, projectID, userID, accept)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, ErrInviteNotFound
		}
		return Member{}, err
	}
	if accept {
		if err := s.ensureProjectRole(ctx, userID, "MEMBER", projectID); err != nil {
			return Member{}, err
		}
		if err := s.revokeProjectRole(ctx, userID, "INVITED_MEMBER", projectID); err != nil {
			return Member{}, err
		}
	} else {
		if err := s.revokeProjectRole(ctx, userID, "INVITED_MEMBER", projectID); err != nil {
			return Member{}, err
		}
	}
	return m, nil
}
