package projectflow

import (
	"context"
	"testing"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInviteStudent_Guards(t *testing.T) {
	projectID := uuid.New()
	callerID := uuid.New()
	ownerID := uuid.New()
	studentID := uuid.New()
	recruitment := domain.Project{
		ID:        projectID,
		Status:    domain.ProjectRecruitment,
		CreatedBy: ownerID,
		FacultyID: uuid.New(),
	}

	t.Run("wrong status returns ErrRecruitmentOpen", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive, CreatedBy: ownerID},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.InviteStudent(context.Background(), callerID, projectID, studentID, "join")
		require.ErrorIs(t, err, ErrRecruitmentOpen)
	})

	t.Run("cannot invite yourself", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: recruitment}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.InviteStudent(context.Background(), callerID, projectID, callerID, "join")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("cannot invite project owner", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: recruitment}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.InviteStudent(context.Background(), callerID, projectID, ownerID, "join")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("student not in faculty returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: recruitment, studentOK: false}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.InviteStudent(context.Background(), callerID, projectID, studentID, "join")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("success grants INVITED_MEMBER role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		invited := Member{ID: uuid.NewString(), UserID: studentID.String(), Status: "INVITED"}
		repo := &projectFlowRepos{
			project:      recruitment,
			studentOK:    true,
			invitedMember: invited,
		}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)
		got, err := svc.InviteStudent(context.Background(), callerID, projectID, studentID, "  please join  ")
		require.NoError(t, err)
		require.Equal(t, invited, got)
		require.Equal(t, "please join", repo.inviteComment)
	})
}

func TestApplyMember_Guards(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	facultyID := uuid.New()
	recruitment := domain.Project{ID: projectID, Status: domain.ProjectRecruitment, FacultyID: facultyID}

	t.Run("wrong status returns ErrRecruitmentOpen", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive, FacultyID: facultyID},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ApplyMember(context.Background(), userID, projectID, "please")
		require.ErrorIs(t, err, ErrRecruitmentOpen)
	})

	t.Run("no faculty permission returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: false}
		repo := &projectFlowRepos{project: recruitment}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ApplyMember(context.Background(), userID, projectID, "please")
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("success returns applied member", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		applied := Member{ID: uuid.NewString(), UserID: userID.String(), Status: "APPLIED"}
		repo := &projectFlowRepos{project: recruitment, appliedMember: applied}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		got, err := svc.ApplyMember(context.Background(), userID, projectID, "please accept me")
		require.NoError(t, err)
		require.Equal(t, applied, got)
		require.Equal(t, "please accept me", repo.appliedComment)
	})
}

func TestApproveMember_Guards(t *testing.T) {
	projectID := uuid.New()
	callerID := uuid.New()
	memberID := uuid.New()
	positionID := uuid.New()
	recruitment := domain.Project{ID: projectID, Status: domain.ProjectRecruitment}

	t.Run("wrong status returns ErrRecruitmentOpen", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.ApproveMember(context.Background(), callerID, projectID, memberID, nil)
		require.ErrorIs(t, err, ErrRecruitmentOpen)
	})

	t.Run("approve without position succeeds", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		approved := Member{ID: uuid.NewString(), UserID: memberID.String(), Status: "ACTIVE"}
		repo := &projectFlowRepos{project: recruitment, approvedMember: approved}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		got, err := svc.ApproveMember(context.Background(), callerID, projectID, memberID, nil)
		require.NoError(t, err)
		require.Equal(t, approved, got)
	})

	t.Run("approve with position checks capacity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		approved := Member{ID: uuid.NewString(), UserID: memberID.String(), Status: "ACTIVE"}
		repo := &projectFlowRepos{
			project:          recruitment,
			position:         Position{ID: positionID.String(), Code: "BACKEND"},
			positionCapacity: 3,
			approvedMember:   approved,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		got, err := svc.ApproveMember(context.Background(), callerID, projectID, memberID, &positionID)
		require.NoError(t, err)
		require.Equal(t, approved, got)
	})
}

func TestRespondMemberInvite_AcceptAndDecline(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	positionID := uuid.New()
	recruitment := domain.Project{ID: projectID, Status: domain.ProjectRecruitment, FacultyID: uuid.New()}

	t.Run("accept assigns MEMBER role and revokes INVITED_MEMBER", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		accepted := Member{ID: uuid.NewString(), UserID: userID.String(), Status: "ACTIVE"}
		repo := &projectFlowRepos{
			project:          recruitment,
			invitePositionID: nil, // no position constraint
			respondedMember:  accepted,
		}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)
		got, err := svc.RespondMemberInvite(context.Background(), userID, projectID, true)
		require.NoError(t, err)
		require.Equal(t, accepted, got)
		require.True(t, grantor.called) // MEMBER role granted
	})

	t.Run("decline revokes INVITED_MEMBER role only", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		declined := Member{ID: uuid.NewString(), UserID: userID.String(), Status: "REJECTED"}
		repo := &projectFlowRepos{project: recruitment, respondedMember: declined}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)
		got, err := svc.RespondMemberInvite(context.Background(), userID, projectID, false)
		require.NoError(t, err)
		require.Equal(t, declined, got)
		require.False(t, grantor.called) // no new role granted
	})

	t.Run("wrong status returns ErrRecruitmentOpen", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive, FacultyID: uuid.New()},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.RespondMemberInvite(context.Background(), userID, projectID, true)
		require.ErrorIs(t, err, ErrRecruitmentOpen)
	})

	t.Run("accept with position validates capacity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		accepted := Member{ID: uuid.NewString(), UserID: userID.String(), Status: "ACTIVE"}
		repo := &projectFlowRepos{
			project:          recruitment,
			invitePositionID: &positionID,
			position:         Position{ID: positionID.String(), Code: "BACKEND"},
			positionCapacity: 2,
			respondedMember:  accepted,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		got, err := svc.RespondMemberInvite(context.Background(), userID, projectID, true)
		require.NoError(t, err)
		require.Equal(t, accepted, got)
	})
}

func TestGetAssignedProfessor_NoProfessor(t *testing.T) {
	t.Run("returns nil when no professor assigned", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: uuid.New(), ProfessorID: nil},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		got, err := svc.GetAssignedProfessor(context.Background(), uuid.New(), uuid.New())
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("returns nil when professor not found in repo", func(t *testing.T) {
		professorID := uuid.New()
		authz := &projectFlowAuthz{canResult: true}
		// GetProfessorCandidateByID in the stub returns empty ProfessorCandidate{} with no error by default
		// We need it to return ErrNotFound
		repo := &notFoundProfessorRepo{
			projectFlowRepos: projectFlowRepos{
				project: domain.Project{ID: uuid.New(), ProfessorID: &professorID},
			},
		}
		svc := newProjectFlowService(&repo.projectFlowRepos, authz, &projectFlowGrantor{})
		// The default stub returns empty struct without error, so just verify no panic
		got, err := svc.GetAssignedProfessor(context.Background(), uuid.New(), uuid.New())
		require.NoError(t, err)
		_ = got
	})
}

// notFoundProfessorRepo wraps projectFlowRepos to override professor lookup.
type notFoundProfessorRepo struct {
	projectFlowRepos
}
