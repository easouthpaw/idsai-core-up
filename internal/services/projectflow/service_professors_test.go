package projectflow

import (
	"context"
	"testing"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type professorFlowRepos struct {
	projectFlowRepos

	professorCandidates []ProfessorCandidate
	professorSearchErr  error
	professorFacultyID  uuid.UUID
	professorTerm       string
	professorLimit      int
	professorRequester  uuid.UUID
	professorOwner      uuid.UUID

	professorOK     bool
	professorOKErr  error
	assignErr       error
	assignProject   uuid.UUID
	assignProfessor uuid.UUID

	professorByID    ProfessorCandidate
	professorByIDErr error
	gotProfessorID   uuid.UUID
	gotFacultyID     uuid.UUID

	respondProject   domain.Project
	respondErr       error
	respondAccept    bool
	respondProfessor uuid.UUID
	respondProjectID uuid.UUID

	reviewInvites      []domain.Project
	reviewInvitesErr   error
	reviewInvitesTerm  string
	reviewInvitesLimit int

	incomingInvites []IncomingInvite
	outgoingApps    []OutgoingApplication
	incomingLimit   int
	outgoingLimit   int
}

func (r *professorFlowRepos) ListProfessorCandidates(ctx context.Context, facultyID uuid.UUID, term string, limit int, requesterUserID, projectOwnerID uuid.UUID) ([]ProfessorCandidate, error) {
	r.professorFacultyID = facultyID
	r.professorTerm = term
	r.professorLimit = limit
	r.professorRequester = requesterUserID
	r.professorOwner = projectOwnerID
	return append([]ProfessorCandidate(nil), r.professorCandidates...), r.professorSearchErr
}

func (r *professorFlowRepos) IsActiveProfessorInFaculty(ctx context.Context, professorID, facultyID uuid.UUID) (bool, error) {
	r.gotProfessorID = professorID
	r.gotFacultyID = facultyID
	return r.professorOK, r.professorOKErr
}

func (r *professorFlowRepos) AssignProjectProfessor(ctx context.Context, projectID, professorID uuid.UUID) error {
	r.assignProject = projectID
	r.assignProfessor = professorID
	return r.assignErr
}

func (r *professorFlowRepos) GetProfessorCandidateByID(ctx context.Context, professorID, facultyID uuid.UUID) (ProfessorCandidate, error) {
	r.gotProfessorID = professorID
	r.gotFacultyID = facultyID
	return r.professorByID, r.professorByIDErr
}

func (r *professorFlowRepos) RespondProfessorInvite(ctx context.Context, projectID, professorID uuid.UUID, accept bool) (domain.Project, error) {
	r.respondProjectID = projectID
	r.respondProfessor = professorID
	r.respondAccept = accept
	return r.respondProject, r.respondErr
}

func (r *professorFlowRepos) ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, term string, limit int) ([]domain.Project, error) {
	r.reviewInvitesTerm = term
	r.reviewInvitesLimit = limit
	return append([]domain.Project(nil), r.reviewInvites...), r.reviewInvitesErr
}

func (r *professorFlowRepos) ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]IncomingInvite, error) {
	r.incomingLimit = limit
	return append([]IncomingInvite(nil), r.incomingInvites...), nil
}

func (r *professorFlowRepos) ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]OutgoingApplication, error) {
	r.outgoingLimit = limit
	return append([]OutgoingApplication(nil), r.outgoingApps...), nil
}

func TestSearchProfessors_NormalizesQueryAndClampsLimit(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	ownerID := uuid.New()
	repo := &professorFlowRepos{
		projectFlowRepos: projectFlowRepos{
			project: domain.Project{ID: projectID, FacultyID: facultyID, CreatedBy: ownerID},
		},
		professorCandidates: []ProfessorCandidate{{UserID: "prof-1", FullName: "Professor Example"}},
	}
	authz := &projectFlowAuthz{canResult: true}
	svc := newProjectFlowService(&repo.projectFlowRepos, authz, &projectFlowGrantor{})
	svc.professorsRepo = repo
	svc.membersRepo = repo

	items, err := svc.SearchProfessors(context.Background(), userID, projectID, " Prof ", 999)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "project.invite_professor", authz.lastPerm)
	require.Equal(t, facultyID, repo.professorFacultyID)
	require.Equal(t, "prof", repo.professorTerm)
	require.Equal(t, 50, repo.professorLimit)
	require.Equal(t, userID, repo.professorRequester)
	require.Equal(t, ownerID, repo.professorOwner)
}

func TestAssignProfessor_ValidatesAndAssigns(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	ownerID := uuid.New()
	professorID := uuid.New()
	authz := &projectFlowAuthz{canResult: true}

	repo := &professorFlowRepos{
		projectFlowRepos: projectFlowRepos{
			project: domain.Project{ID: projectID, FacultyID: facultyID, CreatedBy: ownerID, Title: "AI Platform"},
		},
		professorOK: true,
	}
	svc := newProjectFlowService(&repo.projectFlowRepos, authz, &projectFlowGrantor{})
	svc.professorsRepo = repo

	_, err := svc.AssignProfessor(context.Background(), userID, projectID, userID)
	require.ErrorIs(t, err, ErrInvalidInput)

	project, err := svc.AssignProfessor(context.Background(), userID, projectID, professorID)
	require.NoError(t, err)
	require.Equal(t, projectID, repo.assignProject)
	require.Equal(t, professorID, repo.assignProfessor)
	require.Equal(t, facultyID, repo.gotFacultyID)
	require.Equal(t, "AI Platform", project.Title)
}

func TestGetAssignedProfessorAndRespondInvite(t *testing.T) {
	projectID := uuid.New()
	facultyID := uuid.New()
	professorID := uuid.New()
	authz := &projectFlowAuthz{canResult: true}
	grantor := &projectFlowGrantor{}
	repo := &professorFlowRepos{
		projectFlowRepos: projectFlowRepos{
			project: domain.Project{
				ID:          projectID,
				FacultyID:   facultyID,
				ProfessorID: &professorID,
			},
		},
		professorByID:  ProfessorCandidate{UserID: professorID.String(), FullName: "Professor Example"},
		respondProject: domain.Project{ID: projectID, Title: "AI Platform"},
	}
	svc := newProjectFlowService(&repo.projectFlowRepos, authz, grantor)
	svc.professorsRepo = repo

	assigned, err := svc.GetAssignedProfessor(context.Background(), uuid.New(), projectID)
	require.NoError(t, err)
	require.NotNil(t, assigned)
	require.Equal(t, professorID.String(), assigned.UserID)

	project, err := svc.RespondProfessorInvite(context.Background(), professorID, projectID, true)
	require.NoError(t, err)
	require.Equal(t, "AI Platform", project.Title)
	require.True(t, repo.respondAccept)
	require.True(t, grantor.called)
	require.Equal(t, "PROJECT_PROFESSOR", grantor.roleCode)

	repo.professorByIDErr = ErrNotFound
	assigned, err = svc.GetAssignedProfessor(context.Background(), uuid.New(), projectID)
	require.NoError(t, err)
	require.Nil(t, assigned)

	repo.respondErr = ErrNotFound
	_, err = svc.RespondProfessorInvite(context.Background(), professorID, projectID, false)
	require.ErrorIs(t, err, ErrInviteNotFound)
}

func TestListProfessorAndMemberInviteCollections(t *testing.T) {
	repo := &professorFlowRepos{
		reviewInvites:   []domain.Project{{ID: uuid.New(), Title: "Review 1"}},
		incomingInvites: []IncomingInvite{{ProjectID: "project-1"}},
		outgoingApps:    []OutgoingApplication{{ProjectID: "project-2"}},
	}
	svc := newProjectFlowService(&repo.projectFlowRepos, &projectFlowAuthz{}, &projectFlowGrantor{})
	svc.professorsRepo = repo
	svc.membersRepo = repo

	reviewProjects, err := svc.ListProfessorReviewInvites(context.Background(), uuid.New(), " AI ", 999)
	require.NoError(t, err)
	require.Len(t, reviewProjects, 1)
	require.Equal(t, "ai", repo.reviewInvitesTerm)
	require.Equal(t, 100, repo.reviewInvitesLimit)

	incoming, err := svc.ListIncomingInvites(context.Background(), uuid.New(), 999)
	require.NoError(t, err)
	require.Len(t, incoming, 1)
	require.Equal(t, 100, repo.incomingLimit)

	outgoing, err := svc.ListOutgoingApplications(context.Background(), uuid.New(), 999)
	require.NoError(t, err)
	require.Len(t, outgoing, 1)
	require.Equal(t, 100, repo.outgoingLimit)
}
