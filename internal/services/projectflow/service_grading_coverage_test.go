package projectflow

import (
	"context"
	"testing"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListCriteria(t *testing.T) {
	projectID := uuid.New()
	criteria := []Criterion{
		{ID: uuid.NewString(), Title: "Code Quality", Weight: 60},
		{ID: uuid.NewString(), Title: "Documentation", Weight: 40},
	}
	repo := &projectFlowRepos{criteriaList: criteria}
	svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

	got, err := svc.ListCriteria(context.Background(), projectID)
	require.NoError(t, err)
	require.Equal(t, criteria, got)
}

func TestDeleteProject(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := &projectFlowRepos{}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})
		require.NoError(t, svc.DeleteProject(context.Background(), ownerID, projectID))
	})

	t.Run("repo error propagates", func(t *testing.T) {
		repo := &projectFlowRepos{deletedOwnedErr: ErrNotFound}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})
		require.ErrorIs(t, svc.DeleteProject(context.Background(), ownerID, projectID), ErrNotFound)
	})
}

func TestApproveProject_NotReady(t *testing.T) {
	projectID := uuid.New()
	creatorID := uuid.New()
	professorID := uuid.New()

	t.Run("returns ErrProjectNotReady when no members", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectRecruitment,
				CreatedBy:             creatorID,
				ProfessorID:           &professorID,
				ProfessorReviewStatus: "ACCEPTED",
			},
			criteriaCount: 1,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		_, ready, err := svc.ApproveProject(context.Background(), creatorID, projectID)
		require.ErrorIs(t, err, ErrProjectNotReady)
		require.False(t, ready.CanActivate)
	})

	t.Run("returns ErrProjectNotReady when no professor", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		posID := uuid.NewString()
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectRecruitment,
				CreatedBy:             creatorID,
				ProfessorReviewStatus: "NONE",
			},
			projectMembers: []Member{
				{UserID: creatorID.String(), Status: "ACTIVE"},
				{UserID: uuid.NewString(), Status: "ACTIVE", PositionID: &posID},
			},
			criteriaCount: 1,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		_, ready, err := svc.ApproveProject(context.Background(), creatorID, projectID)
		require.ErrorIs(t, err, ErrProjectNotReady)
		require.False(t, ready.CanActivate)
	})
}

func TestSubmitProjectForGrading_ErrorPaths(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	professorID := uuid.New()

	t.Run("wrong status returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectDraft,
				ProfessorID:           &professorID,
				ProfessorReviewStatus: "ACCEPTED",
			},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("no professor returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("professor not accepted returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectActive,
				ProfessorID:           &professorID,
				ProfessorReviewStatus: "PENDING",
			},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("no tasks returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectActive,
				ProfessorID:           &professorID,
				ProfessorReviewStatus: "ACCEPTED",
			},
			tasksTotal: 0,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("tasks not done returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:                    projectID,
				Status:                domain.ProjectActive,
				ProfessorID:           &professorID,
				ProfessorReviewStatus: "ACCEPTED",
			},
			tasksTotal: 3,
			tasksDone:  1,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestPublishGrading_SuccessAndErrors(t *testing.T) {
	projectID := uuid.New()
	professorID := uuid.New()
	project := domain.Project{
		ID:          projectID,
		Status:      domain.ProjectGrading,
		ProfessorID: &professorID,
	}

	t.Run("success publishes and moves to completed", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       project,
			criteriaCount: 2,
			gradedCount:   2,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		got, err := svc.PublishGrading(context.Background(), professorID, projectID)
		require.NoError(t, err)
		require.Equal(t, project, got)
	})

	t.Run("wrong status returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:          projectID,
				Status:      domain.ProjectActive,
				ProfessorID: &professorID,
			},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.PublishGrading(context.Background(), professorID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("professor mismatch returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		otherID := uuid.New()
		repo := &projectFlowRepos{project: project}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.PublishGrading(context.Background(), otherID, projectID)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("no criteria returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: project, criteriaCount: 0}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.PublishGrading(context.Background(), professorID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestReturnProjectForRetake(t *testing.T) {
	projectID := uuid.New()
	professorID := uuid.New()
	project := domain.Project{
		ID:          projectID,
		Status:      domain.ProjectGrading,
		ProfessorID: &professorID,
	}

	t.Run("success returns project to active", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: project}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		got, err := svc.ReturnProjectForRetake(context.Background(), professorID, projectID)
		require.NoError(t, err)
		require.Equal(t, project, got)
	})

	t.Run("wrong status returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{
				ID:          projectID,
				Status:      domain.ProjectActive,
				ProfessorID: &professorID,
			},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ReturnProjectForRetake(context.Background(), professorID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("professor mismatch returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		otherID := uuid.New()
		repo := &projectFlowRepos{project: project}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ReturnProjectForRetake(context.Background(), otherID, projectID)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("no professor assigned returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectGrading},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ReturnProjectForRetake(context.Background(), professorID, projectID)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})
}

func TestUpsertGrading_WrongStatus(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	criterionID := uuid.New()
	isMet := true

	authz := &projectFlowAuthz{canResult: true}
	repo := &projectFlowRepos{
		project: domain.Project{ID: projectID, Status: domain.ProjectActive},
	}
	svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

	_, err := svc.UpsertGrading(context.Background(), userID, projectID, []CriterionGrade{
		{CriterionID: criterionID.String(), IsMet: &isMet},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestRequireProjectSubmitAccess(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	t.Run("active member passes when no direct permission", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: false}
		repo := &projectFlowRepos{activeMember: true}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		require.NoError(t, svc.requireProjectSubmitAccess(context.Background(), userID, projectID))
	})

	t.Run("non-member without permission returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: false}
		repo := &projectFlowRepos{activeMember: false}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		require.ErrorIs(t, svc.requireProjectSubmitAccess(context.Background(), userID, projectID), domain.ErrForbidden)
	})
}
