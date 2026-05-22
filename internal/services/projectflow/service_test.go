package projectflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type projectFlowAuthz struct {
	canResult bool
	canErr    error
	listCodes []string
	listErr   error
	lastPerm  string
	lastScope rbac.Scope
	canCalls  int
	listCalls int
}

func (a *projectFlowAuthz) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope) (bool, error) {
	a.canCalls++
	a.lastPerm = permissionCode
	a.lastScope = scope
	return a.canResult, a.canErr
}

func (a *projectFlowAuthz) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope rbac.Scope) (bool, error) {
	return false, nil
}

func (a *projectFlowAuthz) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, attrs map[string]interface{}) (bool, error) {
	return false, nil
}

func (a *projectFlowAuthz) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope) ([]string, error) {
	a.listCalls++
	a.lastScope = scope
	return append([]string(nil), a.listCodes...), a.listErr
}

type projectFlowGrantor struct {
	called   bool
	userID   uuid.UUID
	roleCode string
	scope    rbac.Scope
	err      error
}

func (g *projectFlowGrantor) GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error {
	g.called = true
	g.userID = userID
	g.roleCode = roleCode
	g.scope = scope
	return g.err
}

type projectFlowRepos struct {
	project             domain.Project
	projectErr          error
	activeMember        bool
	activeMemberErr     error
	hasRole             bool
	hasRoleErr          error
	revokedRoleCode     string
	revokeRoleErr       error
	updateProjectErr    error
	updatedTitleSet     bool
	updatedTitle        string
	updatedDescSet      bool
	updatedDesc         string
	openRecruitErr      error
	stackCodes          []string
	listStacksErr       error
	replacedStacks      []string
	replaceStacksErr    error
	createdPosition     Position
	createPosErr        error
	createPosCode       string
	createPosName       string
	createPosCap        int
	candidates          []StudentCandidate
	candidatesErr       error
	lastSearchTerm      string
	lastSearchLimit     int
	lastFacultyID       uuid.UUID
	lastOwnerID         uuid.UUID
	studentOK           bool
	studentOKErr        error
	invitedMember       Member
	invitedErr          error
	inviteComment       string
	appliedMember       Member
	appliedErr          error
	appliedComment      string
	projectMembers      []Member
	projectMembersErr   error
	approvedMember      Member
	approvedErr         error
	rejectedMember      Member
	rejectedErr         error
	removedMember       Member
	removedErr          error
	positionedMember    Member
	positionedErr       error
	invitePositionID    *uuid.UUID
	invitePosErr        error
	respondedMember     Member
	respondedErr        error
	criteriaCreated     Criterion
	createCriterionErr  error
	criteriaList        []Criterion
	criteriaListErr     error
	criterionGrades     []CriterionGrade
	criterionGradesErr  error
	upsertedGrades      []CriterionGradeUpsert
	upsertGradesErr     error
	criteriaCount       int
	criteriaCountErr    error
	gradedCount         int
	gradedCountErr      error
	activatedErr        error
	tasksTotal          int
	tasksDone           int
	tasksSummaryErr     error
	movedToGradingErr   error
	returnedToActiveErr error
	movedToCompletedErr error
	deletedOwnedErr     error

	position            Position
	positionErr         error
	positionCapacity    int
	positionCapacityErr error
	ensuredPosition     Position
	ensurePositionErr   error

	activeByPosition    int
	activeByPositionErr error
	memberStatus        string
	memberPositionID    *uuid.UUID
	memberStatusErr     error

	ensureTaskActivityErr error
	insertTaskActivityErr error
	insertedEventType     string
	insertedFromStatus    string
	insertedToStatus      string
	insertedTitle         string
	insertedComment       string
	insertedAttachments   []string

	roleCodes            []string
	listRoleCodesErr     error
	replacedAssignable   []string
	replacedWanted       []string
	replaceAssignableErr error
	customAccessRoles    []AccessCatalogItem
	customAccessRolesErr error
	createdAccessRole    AccessCatalogItem
	createAccessRoleErr  error
	createAccessRoleCode string
	createAccessRoleName string
	createAccessRoleDesc string
	createAccessRolePerm []string
	accessStatus         string
	accessCreatorID      uuid.UUID
	accessErr            error

	task                Task
	taskErr             error
	createdTaskID       uuid.UUID
	createTaskErr       error
	listedTasks         []Task
	listTasksErr        error
	taskStatus          string
	taskTitle           string
	taskStatusErr       error
	updatedTaskID       uuid.UUID
	updateTaskErr       error
	assignPositionID    uuid.UUID
	assignPrevStatus    string
	assignTaskTitle     string
	assignPrevAssignee  *uuid.UUID
	assignCtxErr        error
	assignedTaskID      uuid.UUID
	assignTaskErr       error
	taskActivities      []TaskActivity
	taskActivitiesErr   error
	completeAssigneeID  *uuid.UUID
	completeStatus      string
	completeTitle       string
	completeCtxErr      error
	submissionComment   string
	submissionFiles     []string
	upsertSubmissionErr error
	markDoneID          uuid.UUID
	markDoneErr         error
	claimErr            error
	deleteTaskErr       error
}

func (r *projectFlowRepos) ReplaceProjectStacks(ctx context.Context, projectID uuid.UUID, stackCodes []string) error {
	r.replacedStacks = append([]string(nil), stackCodes...)
	return r.replaceStacksErr
}

func (r *projectFlowRepos) ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return append([]string(nil), r.stackCodes...), r.listStacksErr
}

func (r *projectFlowRepos) GetProjectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	return r.project, r.projectErr
}

func (r *projectFlowRepos) IsActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return r.activeMember, r.activeMemberErr
}

func (r *projectFlowRepos) HasProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) (bool, error) {
	return r.hasRole, r.hasRoleErr
}

func (r *projectFlowRepos) RevokeProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) error {
	r.revokedRoleCode = roleCode
	return r.revokeRoleErr
}

func (r *projectFlowRepos) UpdateProject(ctx context.Context, projectID uuid.UUID, titleSet bool, title string, descriptionSet bool, description string) error {
	r.updatedTitleSet = titleSet
	r.updatedTitle = title
	r.updatedDescSet = descriptionSet
	r.updatedDesc = description
	return r.updateProjectErr
}

func (r *projectFlowRepos) OpenProjectRecruitment(ctx context.Context, projectID uuid.UUID) error {
	return r.openRecruitErr
}

func (r *projectFlowRepos) ListStudentCandidates(ctx context.Context, facultyID, projectID, requesterUserID, projectOwnerID uuid.UUID, term string, limit int) ([]StudentCandidate, error) {
	r.lastFacultyID = facultyID
	r.lastOwnerID = projectOwnerID
	r.lastSearchTerm = term
	r.lastSearchLimit = limit
	return append([]StudentCandidate(nil), r.candidates...), r.candidatesErr
}

func (r *projectFlowRepos) CreateProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (Position, error) {
	r.createPosCode = code
	r.createPosName = name
	r.createPosCap = capacity
	return r.createdPosition, r.createPosErr
}

func (r *projectFlowRepos) EnsureProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (Position, error) {
	return r.ensuredPosition, r.ensurePositionErr
}

func (r *projectFlowRepos) ListProjectPositions(ctx context.Context, projectID uuid.UUID) ([]Position, error) {
	return nil, nil
}

func (r *projectFlowRepos) GetProjectPosition(ctx context.Context, projectID, positionID uuid.UUID) (Position, error) {
	return r.position, r.positionErr
}

func (r *projectFlowRepos) GetProjectPositionCapacity(ctx context.Context, projectID, positionID uuid.UUID) (int, error) {
	return r.positionCapacity, r.positionCapacityErr
}

func (r *projectFlowRepos) SumProjectPositionCapacities(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (r *projectFlowRepos) IsActiveStudentInFaculty(ctx context.Context, studentID, facultyID uuid.UUID) (bool, error) {
	return r.studentOK, r.studentOKErr
}

func (r *projectFlowRepos) UpsertInvitedMember(ctx context.Context, projectID, studentID, invitedBy uuid.UUID, comment string) (Member, error) {
	r.inviteComment = comment
	return r.invitedMember, r.invitedErr
}

func (r *projectFlowRepos) UpsertAppliedMember(ctx context.Context, projectID, userID uuid.UUID, comment string) (Member, error) {
	r.appliedComment = comment
	return r.appliedMember, r.appliedErr
}

func (r *projectFlowRepos) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]Member, error) {
	return append([]Member(nil), r.projectMembers...), r.projectMembersErr
}

func (r *projectFlowRepos) CountActiveMembersByPosition(ctx context.Context, projectID, positionID uuid.UUID, excludeUserID *uuid.UUID) (int, error) {
	return r.activeByPosition, r.activeByPositionErr
}

func (r *projectFlowRepos) GetProjectMemberStatusAndPosition(ctx context.Context, projectID, userID uuid.UUID) (string, *uuid.UUID, error) {
	return r.memberStatus, r.memberPositionID, r.memberStatusErr
}

func (r *projectFlowRepos) CountActiveMembersWithPosition(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (r *projectFlowRepos) ApproveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID, positionID *uuid.UUID) (Member, error) {
	return r.approvedMember, r.approvedErr
}

func (r *projectFlowRepos) RejectProjectMemberApplication(ctx context.Context, projectID, memberUserID uuid.UUID) (Member, error) {
	return r.rejectedMember, r.rejectedErr
}

func (r *projectFlowRepos) RemoveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID) (Member, error) {
	return r.removedMember, r.removedErr
}

func (r *projectFlowRepos) SetActiveMemberPosition(ctx context.Context, projectID, memberUserID, positionID uuid.UUID) (Member, error) {
	return r.positionedMember, r.positionedErr
}

func (r *projectFlowRepos) GetInvitedMemberPosition(ctx context.Context, projectID, userID uuid.UUID) (*uuid.UUID, error) {
	return r.invitePositionID, r.invitePosErr
}

func (r *projectFlowRepos) RespondMemberInvite(ctx context.Context, projectID, userID uuid.UUID, accept bool) (Member, error) {
	return r.respondedMember, r.respondedErr
}

func (r *projectFlowRepos) ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]IncomingInvite, error) {
	return nil, nil
}

func (r *projectFlowRepos) ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]OutgoingApplication, error) {
	return nil, nil
}

func (r *projectFlowRepos) ListProfessorCandidates(ctx context.Context, facultyID uuid.UUID, term string, limit int, requesterUserID, projectOwnerID uuid.UUID) ([]ProfessorCandidate, error) {
	return nil, nil
}

func (r *projectFlowRepos) IsActiveProfessorInFaculty(ctx context.Context, professorID, facultyID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *projectFlowRepos) AssignProjectProfessor(ctx context.Context, projectID, professorID uuid.UUID) error {
	return nil
}

func (r *projectFlowRepos) GetProfessorCandidateByID(ctx context.Context, professorID, facultyID uuid.UUID) (ProfessorCandidate, error) {
	return ProfessorCandidate{}, nil
}

func (r *projectFlowRepos) RespondProfessorInvite(ctx context.Context, projectID, professorID uuid.UUID, accept bool) (domain.Project, error) {
	return domain.Project{}, nil
}

func (r *projectFlowRepos) ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, term string, limit int) ([]domain.Project, error) {
	return nil, nil
}

func (r *projectFlowRepos) CreateProjectCriterion(ctx context.Context, projectID, userID uuid.UUID, title, description string, weight int) (Criterion, error) {
	return r.criteriaCreated, r.createCriterionErr
}

func (r *projectFlowRepos) ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]Criterion, error) {
	return append([]Criterion(nil), r.criteriaList...), r.criteriaListErr
}

func (r *projectFlowRepos) ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]CriterionGrade, error) {
	return append([]CriterionGrade(nil), r.criterionGrades...), r.criterionGradesErr
}

func (r *projectFlowRepos) UpsertProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID, items []CriterionGradeUpsert) error {
	r.upsertedGrades = append([]CriterionGradeUpsert(nil), items...)
	return r.upsertGradesErr
}

func (r *projectFlowRepos) CountProjectCriteria(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.criteriaCount, r.criteriaCountErr
}

func (r *projectFlowRepos) CountProjectGradedCriteria(ctx context.Context, projectID, professorID uuid.UUID) (int, error) {
	return r.gradedCount, r.gradedCountErr
}

func (r *projectFlowRepos) ActivateProject(ctx context.Context, projectID uuid.UUID) error {
	return r.activatedErr
}

func (r *projectFlowRepos) CountProjectTasksSummary(ctx context.Context, projectID uuid.UUID) (int, int, error) {
	return r.tasksTotal, r.tasksDone, r.tasksSummaryErr
}

func (r *projectFlowRepos) MoveProjectToGrading(ctx context.Context, projectID uuid.UUID) error {
	return r.movedToGradingErr
}

func (r *projectFlowRepos) ReturnProjectToActive(ctx context.Context, projectID uuid.UUID) error {
	return r.returnedToActiveErr
}

func (r *projectFlowRepos) MoveProjectToCompleted(ctx context.Context, projectID uuid.UUID) error {
	return r.movedToCompletedErr
}

func (r *projectFlowRepos) DeleteOwnedProject(ctx context.Context, projectID, ownerID uuid.UUID) error {
	return r.deletedOwnedErr
}

func (r *projectFlowRepos) CreateTask(ctx context.Context, projectID uuid.UUID, title, description string, positionID uuid.UUID, assigneeUserID *uuid.UUID, status string, createdBy uuid.UUID, dueAt *time.Time) (uuid.UUID, error) {
	return r.createdTaskID, r.createTaskErr
}

func (r *projectFlowRepos) GetTaskByID(ctx context.Context, projectID, taskID uuid.UUID) (Task, error) {
	return r.task, r.taskErr
}

func (r *projectFlowRepos) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]Task, error) {
	return append([]Task(nil), r.listedTasks...), r.listTasksErr
}

func (r *projectFlowRepos) EnsureTaskActivityLogAvailable(ctx context.Context) error {
	return r.ensureTaskActivityErr
}

func (r *projectFlowRepos) GetTaskStatusAndTitle(ctx context.Context, projectID, taskID uuid.UUID) (string, string, error) {
	return r.taskStatus, r.taskTitle, r.taskStatusErr
}

func (r *projectFlowRepos) UpdateTaskStatus(ctx context.Context, projectID, taskID uuid.UUID, status string) (uuid.UUID, error) {
	return r.updatedTaskID, r.updateTaskErr
}

func (r *projectFlowRepos) GetTaskAssignContext(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, string, string, *uuid.UUID, error) {
	return r.assignPositionID, r.assignPrevStatus, r.assignTaskTitle, r.assignPrevAssignee, r.assignCtxErr
}

func (r *projectFlowRepos) AssignTaskToUser(ctx context.Context, projectID, taskID, assigneeUserID uuid.UUID) (uuid.UUID, error) {
	return r.assignedTaskID, r.assignTaskErr
}

func (r *projectFlowRepos) ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]TaskActivity, error) {
	return append([]TaskActivity(nil), r.taskActivities...), r.taskActivitiesErr
}

func (r *projectFlowRepos) GetTaskCompleteContext(ctx context.Context, projectID, taskID uuid.UUID) (*uuid.UUID, string, string, error) {
	return r.completeAssigneeID, r.completeStatus, r.completeTitle, r.completeCtxErr
}

func (r *projectFlowRepos) UpsertTaskSubmission(ctx context.Context, projectID, taskID, userID uuid.UUID, comment string, attachments []string) error {
	r.submissionComment = comment
	r.submissionFiles = append([]string(nil), attachments...)
	return r.upsertSubmissionErr
}

func (r *projectFlowRepos) MarkTaskDone(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, error) {
	return r.markDoneID, r.markDoneErr
}

func (r *projectFlowRepos) ClaimTask(ctx context.Context, projectID, taskID, userID uuid.UUID) error {
	return r.claimErr
}

func (r *projectFlowRepos) DeleteTask(ctx context.Context, projectID, taskID uuid.UUID) error {
	return r.deleteTaskErr
}

func (r *projectFlowRepos) InsertTaskActivity(ctx context.Context, projectID, taskID uuid.UUID, actorUserID *uuid.UUID, eventType, fromStatus, toStatus, title, comment string, attachments []string) error {
	r.insertedEventType = eventType
	r.insertedFromStatus = fromStatus
	r.insertedToStatus = toStatus
	r.insertedTitle = title
	r.insertedComment = comment
	r.insertedAttachments = append([]string(nil), attachments...)
	return r.insertTaskActivityErr
}

func (r *projectFlowRepos) ListProjectRoleCodes(ctx context.Context, userID, projectID uuid.UUID) ([]string, error) {
	return append([]string(nil), r.roleCodes...), r.listRoleCodesErr
}

func (r *projectFlowRepos) ReplaceAssignableRoles(ctx context.Context, userID, projectID uuid.UUID, assignableCodes []string, wantCodes []string) error {
	r.replacedAssignable = append([]string(nil), assignableCodes...)
	r.replacedWanted = append([]string(nil), wantCodes...)
	return r.replaceAssignableErr
}

func (r *projectFlowRepos) GetMemberStatusAndCreator(ctx context.Context, userID, projectID uuid.UUID) (string, uuid.UUID, error) {
	return r.accessStatus, r.accessCreatorID, r.accessErr
}

func (r *projectFlowRepos) ListProjectAccessRoles(ctx context.Context, projectID uuid.UUID) ([]AccessCatalogItem, error) {
	return append([]AccessCatalogItem(nil), r.customAccessRoles...), r.customAccessRolesErr
}

func (r *projectFlowRepos) CreateProjectAccessRole(ctx context.Context, projectID, createdBy uuid.UUID, roleCode, displayCode, name, description string, permissionCodes []string) (AccessCatalogItem, error) {
	r.createAccessRoleCode = displayCode
	r.createAccessRoleName = name
	r.createAccessRoleDesc = description
	r.createAccessRolePerm = append([]string(nil), permissionCodes...)
	if r.createAccessRoleErr != nil {
		return AccessCatalogItem{}, r.createAccessRoleErr
	}
	if r.createdAccessRole.Code != "" {
		return r.createdAccessRole, nil
	}
	return AccessCatalogItem{
		Code:            roleCode,
		DisplayCode:     displayCode,
		Name:            name,
		Description:     description,
		PermissionCodes: append([]string(nil), permissionCodes...),
		Custom:          true,
	}, nil
}

func newProjectFlowService(repo *projectFlowRepos, authz *projectFlowAuthz, grantor *projectFlowGrantor) *Service {
	return &Service{
		authz:          authz,
		grantor:        grantor,
		projectsRepo:   repo,
		stacksRepo:     repo,
		positionsRepo:  repo,
		membersRepo:    repo,
		professorsRepo: repo,
		criteriaRepo:   repo,
		lifecycleRepo:  repo,
		tasksRepo:      repo,
		accessRepo:     repo,
		now:            time.Now,
	}
}

func TestNewService_PanicsOnMissingRequiredRepos(t *testing.T) {
	repo := &projectFlowRepos{}
	authz := &projectFlowAuthz{}
	grantor := &projectFlowGrantor{}

	require.NotNil(t, NewService(authz, grantor, repo, repo, repo, repo, repo, repo, repo, repo, repo))

	cases := []struct {
		name string
		call func()
		msg  string
	}{
		{
			name: "projects repo",
			call: func() { NewService(authz, grantor, nil, repo, repo, repo, repo, repo, repo, repo, repo) },
			msg:  "projectflow.NewService: projectsRepo is nil",
		},
		{
			name: "stacks repo",
			call: func() { NewService(authz, grantor, repo, nil, repo, repo, repo, repo, repo, repo, repo) },
			msg:  "projectflow.NewService: stacksRepo is nil",
		},
		{
			name: "positions repo",
			call: func() { NewService(authz, grantor, repo, repo, nil, repo, repo, repo, repo, repo, repo) },
			msg:  "projectflow.NewService: positionsRepo is nil",
		},
		{
			name: "members repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, nil, repo, repo, repo, repo, repo) },
			msg:  "projectflow.NewService: membersRepo is nil",
		},
		{
			name: "professors repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, repo, nil, repo, repo, repo, repo) },
			msg:  "projectflow.NewService: professorsRepo is nil",
		},
		{
			name: "criteria repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, repo, repo, nil, repo, repo, repo) },
			msg:  "projectflow.NewService: criteriaRepo is nil",
		},
		{
			name: "lifecycle repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, repo, repo, repo, nil, repo, repo) },
			msg:  "projectflow.NewService: lifecycleRepo is nil",
		},
		{
			name: "tasks repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, repo, repo, repo, repo, nil, repo) },
			msg:  "projectflow.NewService: tasksRepo is nil",
		},
		{
			name: "access repo",
			call: func() { NewService(authz, grantor, repo, repo, repo, repo, repo, repo, repo, repo, nil) },
			msg:  "projectflow.NewService: accessRepo is nil",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.msg, tc.call)
		})
	}
}

func TestProjectFlowHelpers(t *testing.T) {
	require.Equal(t, []string{"AI", "GO", "MACHINE LEARNING"}, normalizeStackCodes([]string{" go ", "AI", "go", "", "Machine Learning"}))
	require.Equal(t, []Stack{{Code: "GO"}, {Code: "AI"}}, stacksFromCodes([]string{"GO", "AI"}))
	require.Equal(t, "BACKEND_ENGINEER", normalizePositionCode("", " Backend Engineer "))
	require.True(t, isSystemTaskPositionCode(" team_lead "))
	require.Equal(t, "search term", normalizeSearchQuery(" Search TERM "))
	require.Equal(t, 20, clampLimit(0, 0))
	require.Equal(t, 10, clampLimit(50, 10))
	require.Equal(t, []string{"https://one", "https://two"}, normalizeTaskAttachments([]string{" https://one ", "", "https://one", "https://two"}))
	require.Equal(t, []string{"CO_LEAD", "RECRUITER", "TASK_MANAGER"}, assignableRoleCodes())
}

func TestAppendTaskActivity_NormalizesFields(t *testing.T) {
	repo := &projectFlowRepos{}
	svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

	err := svc.appendTaskActivity(context.Background(), uuid.New(), uuid.New(), nil, " status_changed ", " todo ", " done ", "  Task done  ", "  comment  ", []string{" https://one ", "https://one", ""})
	require.NoError(t, err)
	require.Equal(t, "STATUS_CHANGED", repo.insertedEventType)
	require.Equal(t, "TODO", repo.insertedFromStatus)
	require.Equal(t, "DONE", repo.insertedToStatus)
	require.Equal(t, "Task done", repo.insertedTitle)
	require.Equal(t, "comment", repo.insertedComment)
	require.Equal(t, []string{"https://one"}, repo.insertedAttachments)
}

func TestEnsurePositionHelpers(t *testing.T) {
	projectID := uuid.New()
	positionID := uuid.New()

	t.Run("ensure position exists maps not found", func(t *testing.T) {
		repo := &projectFlowRepos{positionCapacityErr: ErrNotFound}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		_, err := svc.ensurePositionExists(context.Background(), projectID, positionID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("member-assignable position rejects reserved code", func(t *testing.T) {
		repo := &projectFlowRepos{position: Position{Code: SystemTaskPositionTeamLeadCode}}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.ensureMemberAssignablePosition(context.Background(), projectID, positionID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("ensure position capacity returns full", func(t *testing.T) {
		repo := &projectFlowRepos{
			positionCapacity: 1,
			activeByPosition: 1,
		}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.ensurePositionCapacity(context.Background(), projectID, positionID, nil)
		require.ErrorIs(t, err, ErrPositionFull)
	})
}

func TestEnsureAssigneeMatchesPosition(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	positionID := uuid.New()

	t.Run("team lead position requires project lead", func(t *testing.T) {
		repo := &projectFlowRepos{
			position: Position{Code: SystemTaskPositionTeamLeadCode},
			project:  domain.Project{CreatedBy: uuid.New()},
			hasRole:  false,
		}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.ensureAssigneeMatchesPosition(context.Background(), projectID, userID, positionID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("member position requires active member with matching role", func(t *testing.T) {
		repo := &projectFlowRepos{
			position:         Position{Code: "BACKEND"},
			memberStatus:     "ACTIVE",
			memberPositionID: &positionID,
		}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.ensureAssigneeMatchesPosition(context.Background(), projectID, userID, positionID)
		require.NoError(t, err)
	})

	t.Run("member position rejects missing membership", func(t *testing.T) {
		repo := &projectFlowRepos{
			position:         Position{Code: "BACKEND"},
			memberStatusErr:  ErrNotFound,
			memberPositionID: nil,
		}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.ensureAssigneeMatchesPosition(context.Background(), projectID, userID, positionID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestProjectAccessHelpers(t *testing.T) {
	projectID := uuid.New()
	facultyID := uuid.New()
	userID := uuid.New()

	t.Run("require project permission allows and stores scope", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		svc := newProjectFlowService(&projectFlowRepos{}, authz, &projectFlowGrantor{})

		err := svc.requireProjectPermission(context.Background(), userID, "member.access.manage", projectID)
		require.NoError(t, err)
		require.Equal(t, "member.access.manage", authz.lastPerm)
		require.Equal(t, rbac.ScopeProject, authz.lastScope.Type)
		require.Equal(t, projectID, *authz.lastScope.ID)
	})

	t.Run("require faculty permission denies forbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: false}
		svc := newProjectFlowService(&projectFlowRepos{}, authz, &projectFlowGrantor{})

		err := svc.requireFacultyPermission(context.Background(), userID, "project.create", facultyID)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("project edit access falls back to active membership", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: false}
		repo := &projectFlowRepos{activeMember: true}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		err := svc.requireProjectEditAccess(context.Background(), userID, projectID)
		require.NoError(t, err)
	})

	t.Run("project submit access returns authz error", func(t *testing.T) {
		authz := &projectFlowAuthz{canErr: errors.New("rbac unavailable")}
		svc := newProjectFlowService(&projectFlowRepos{}, authz, &projectFlowGrantor{})

		err := svc.requireProjectSubmitAccess(context.Background(), userID, projectID)
		require.ErrorContains(t, err, "rbac unavailable")
	})
}

func TestProjectRoleDelegationHelpers(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()

	t.Run("ensure project role grants when absent", func(t *testing.T) {
		repo := &projectFlowRepos{hasRole: false}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, grantor)

		err := svc.ensureProjectRole(context.Background(), userID, "TEAM_LEAD", projectID)
		require.NoError(t, err)
		require.True(t, grantor.called)
		require.Equal(t, "TEAM_LEAD", grantor.roleCode)
		require.Equal(t, projectID, *grantor.scope.ID)
	})

	t.Run("ensure project role skips grant when already assigned", func(t *testing.T) {
		repo := &projectFlowRepos{hasRole: true}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, grantor)

		err := svc.ensureProjectRole(context.Background(), userID, "TEAM_LEAD", projectID)
		require.NoError(t, err)
		require.False(t, grantor.called)
	})

	t.Run("revoke project role delegates to repo", func(t *testing.T) {
		repo := &projectFlowRepos{}
		svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

		err := svc.revokeProjectRole(context.Background(), userID, "TEAM_LEAD", projectID)
		require.NoError(t, err)
		require.Equal(t, "TEAM_LEAD", repo.revokedRoleCode)
	})
}

func TestMemberAccessOperations(t *testing.T) {
	projectID := uuid.New()
	callerID := uuid.New()
	targetUserID := uuid.New()
	creatorID := uuid.New()

	t.Run("get access catalog requires permission", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{customAccessRoles: []AccessCatalogItem{{
			Code:            projectAccessRoleCode(projectID, "TESTER"),
			DisplayCode:     "TESTER",
			Name:            "Tester",
			Description:     "Checks the project",
			PermissionCodes: []string{"task.view"},
			Custom:          true,
		}}}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		items, err := svc.GetAccessCatalog(context.Background(), callerID, projectID)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, "TESTER", items[3].DisplayCode)
		require.True(t, items[3].Custom)
	})

	t.Run("list access permissions returns assignable permission catalog", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		svc := newProjectFlowService(&projectFlowRepos{}, authz, &projectFlowGrantor{})

		items, err := svc.ListProjectAccessPermissions(context.Background(), callerID, projectID)
		require.NoError(t, err)
		require.NotEmpty(t, items)
		require.Contains(t, projectPermissionCodeSet(), "member.access.manage")
	})

	t.Run("create project access role normalizes and validates permissions", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		item, err := svc.CreateProjectAccessRole(context.Background(), callerID, projectID, " manager roles ", " Менеджер ролей ", "  Может выдавать доступ  ", []string{"member.access.manage", "task.view", "task.view"})
		require.NoError(t, err)
		require.True(t, item.Custom)
		require.Equal(t, "MANAGER_ROLES", repo.createAccessRoleCode)
		require.Equal(t, "Менеджер ролей", repo.createAccessRoleName)
		require.Equal(t, "Может выдавать доступ", repo.createAccessRoleDesc)
		require.Equal(t, []string{"member.access.manage", "task.view"}, repo.createAccessRolePerm)

		_, err = svc.CreateProjectAccessRole(context.Background(), callerID, projectID, "team_lead", "Bad", "", nil)
		require.ErrorIs(t, err, ErrInvalidInput)

		_, err = svc.CreateProjectAccessRole(context.Background(), callerID, projectID, "tester", "Tester", "", []string{"project.delete"})
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("get member access rejects non-active member", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{accessStatus: "PENDING"}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		access, err := svc.GetMemberAccess(context.Background(), callerID, projectID, targetUserID)
		require.Nil(t, access)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("replace member access validates one assignable role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true, listCodes: []string{"project.view", "task.view"}}
		repo := &projectFlowRepos{
			accessStatus:    "ACTIVE",
			accessCreatorID: creatorID,
			roleCodes:       []string{"TASK_MANAGER", "MEMBER", "CO_LEAD"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		access, err := svc.ReplaceMemberAccess(context.Background(), callerID, projectID, targetUserID, []string{"TASK_MANAGER", "TASK_MANAGER"})
		require.NoError(t, err)
		require.NotNil(t, access)
		require.Equal(t, []string{"CO_LEAD", "RECRUITER", "TASK_MANAGER"}, repo.replacedAssignable)
		require.Equal(t, []string{"TASK_MANAGER"}, repo.replacedWanted)
		require.Equal(t, []string{"TASK_MANAGER", "MEMBER", "CO_LEAD"}, access.RoleCodes)
		require.Equal(t, []string{"TASK_MANAGER", "CO_LEAD"}, access.ManagedRoleCodes)
		require.Equal(t, []string{"project.view", "task.view"}, access.EffectivePermissionCodes)
	})

	t.Run("replace member access rejects multiple roles", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			accessStatus:    "ACTIVE",
			accessCreatorID: creatorID,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		access, err := svc.ReplaceMemberAccess(context.Background(), callerID, projectID, targetUserID, []string{"TASK_MANAGER", "CO_LEAD"})
		require.Nil(t, access)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("replace member access rejects creator and unknown role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}

		repo := &projectFlowRepos{
			accessStatus:    "ACTIVE",
			accessCreatorID: targetUserID,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		access, err := svc.ReplaceMemberAccess(context.Background(), callerID, projectID, targetUserID, []string{"CO_LEAD"})
		require.Nil(t, access)
		require.ErrorIs(t, err, ErrSystemManagedAccess)

		repo = &projectFlowRepos{
			accessStatus:    "ACTIVE",
			accessCreatorID: creatorID,
		}
		svc = newProjectFlowService(repo, authz, &projectFlowGrantor{})
		access, err = svc.ReplaceMemberAccess(context.Background(), callerID, projectID, targetUserID, []string{"UNKNOWN_ROLE"})
		require.Nil(t, access)
		require.ErrorIs(t, err, ErrUnknownRoleCode)
	})

	t.Run("my permissions delegates to authorizer", func(t *testing.T) {
		authz := &projectFlowAuthz{listCodes: []string{"task.view", "task.update"}}
		svc := newProjectFlowService(&projectFlowRepos{}, authz, &projectFlowGrantor{})

		codes, err := svc.MyPermissions(context.Background(), targetUserID, projectID)
		require.NoError(t, err)
		require.Equal(t, []string{"task.view", "task.update"}, codes)
		require.Equal(t, rbac.ScopeProject, authz.lastScope.Type)
		require.Equal(t, projectID, *authz.lastScope.ID)
	})
}

func TestTaskAccessHelpers(t *testing.T) {
	task := Task{ID: uuid.NewString(), Title: "Task"}
	repo := &projectFlowRepos{
		task:                  task,
		ensureTaskActivityErr: errors.New("missing log"),
	}
	svc := newProjectFlowService(repo, &projectFlowAuthz{}, &projectFlowGrantor{})

	require.ErrorContains(t, svc.ensureTaskActivityAvailable(context.Background()), "missing log")

	got, err := svc.taskByID(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, task, got)
}

func TestMemberManagementFlows(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	facultyID := uuid.New()
	project := domain.Project{
		ID:        projectID,
		Status:    domain.ProjectRecruitment,
		CreatedBy: uuid.New(),
		FacultyID: facultyID,
	}

	t.Run("update project trims payload and returns refreshed project", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: project}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		title := "  New title  "
		description := "  Updated description  "
		got, err := svc.UpdateProject(context.Background(), userID, projectID, &title, &description)
		require.NoError(t, err)
		require.Equal(t, project, got)
		require.True(t, repo.updatedTitleSet)
		require.Equal(t, "New title", repo.updatedTitle)
		require.True(t, repo.updatedDescSet)
		require.Equal(t, "Updated description", repo.updatedDesc)
	})

	t.Run("set and list stacks normalize values", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{stackCodes: []string{"GO", "AI"}}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		stacks, err := svc.SetStacks(context.Background(), userID, projectID, []string{" ai ", "GO", "go"})
		require.NoError(t, err)
		require.Equal(t, []string{"AI", "GO"}, repo.replacedStacks)
		require.Equal(t, []Stack{{Code: "GO"}, {Code: "AI"}}, stacks)
	})

	t.Run("open recruitment maps not found to invalid input", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{openRecruitErr: ErrNotFound}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		_, err := svc.OpenRecruitment(context.Background(), userID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("create position normalizes code and default capacity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{createdPosition: Position{Code: "BACKEND_ENGINEER"}}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		got, err := svc.CreatePosition(context.Background(), userID, projectID, "", " Backend Engineer ", 0)
		require.NoError(t, err)
		require.Equal(t, "BACKEND_ENGINEER", repo.createPosCode)
		require.Equal(t, 1, repo.createPosCap)
		require.Equal(t, Position{Code: "BACKEND_ENGINEER"}, got)
	})

	t.Run("list student candidates normalizes search and limit", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:    project,
			candidates: []StudentCandidate{{UserID: memberID.String(), FullName: "Student"}},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		items, err := svc.ListStudentCandidates(context.Background(), userID, projectID, "  Alice  ", 1000)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, "alice", repo.lastSearchTerm)
		require.Equal(t, 100, repo.lastSearchLimit)
		require.Equal(t, facultyID, repo.lastFacultyID)
		require.Equal(t, project.CreatedBy, repo.lastOwnerID)
	})

	t.Run("invite student trims comment and grants invited role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       project,
			studentOK:     true,
			invitedMember: Member{UserID: memberID.String(), Status: "INVITED"},
		}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)

		comment := strings.Repeat("a", 600)
		member, err := svc.InviteStudent(context.Background(), userID, projectID, memberID, comment)
		require.NoError(t, err)
		require.Equal(t, Member{UserID: memberID.String(), Status: "INVITED"}, member)
		require.Len(t, repo.inviteComment, 500)
		require.True(t, grantor.called)
		require.Equal(t, "INVITED_MEMBER", grantor.roleCode)
	})

	t.Run("apply member trims comment", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       project,
			appliedMember: Member{UserID: userID.String(), Status: "PENDING"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		member, err := svc.ApplyMember(context.Background(), userID, projectID, "  hello  ")
		require.NoError(t, err)
		require.Equal(t, Member{UserID: userID.String(), Status: "PENDING"}, member)
		require.Equal(t, "hello", repo.appliedComment)
	})

	t.Run("approve member assigns project role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		positionID := uuid.New()
		repo := &projectFlowRepos{
			project:          project,
			position:         Position{Code: "BACKEND"},
			positionCapacity: 2,
			activeByPosition: 1,
			approvedMember:   Member{UserID: memberID.String(), Status: "ACTIVE"},
		}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)

		member, err := svc.ApproveMember(context.Background(), userID, projectID, memberID, &positionID)
		require.NoError(t, err)
		require.Equal(t, Member{UserID: memberID.String(), Status: "ACTIVE"}, member)
		require.True(t, grantor.called)
		require.Equal(t, "MEMBER", grantor.roleCode)
	})

	t.Run("respond member invite accepts and revokes invited role", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		positionID := uuid.New()
		repo := &projectFlowRepos{
			project:          project,
			invitePositionID: &positionID,
			position:         Position{Code: "BACKEND"},
			positionCapacity: 3,
			activeByPosition: 1,
			respondedMember:  Member{UserID: userID.String(), Status: "ACTIVE"},
		}
		grantor := &projectFlowGrantor{}
		svc := newProjectFlowService(repo, authz, grantor)

		member, err := svc.RespondMemberInvite(context.Background(), userID, projectID, true)
		require.NoError(t, err)
		require.Equal(t, Member{UserID: userID.String(), Status: "ACTIVE"}, member)
		require.True(t, grantor.called)
		require.Equal(t, "INVITED_MEMBER", repo.revokedRoleCode)
	})
}

func TestGradingFlows(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	professorID := uuid.New()
	project := domain.Project{
		ID:                    projectID,
		Status:                domain.ProjectGrading,
		CreatedBy:             uuid.New(),
		ProfessorID:           &professorID,
		ProfessorReviewStatus: "ACCEPTED",
	}

	t.Run("create criterion normalizes weight", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{criteriaCreated: Criterion{Title: "Architecture", Weight: 100}}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		criterion, err := svc.CreateCriterion(context.Background(), userID, projectID, "  Architecture  ", "  desc ", 150)
		require.NoError(t, err)
		require.Equal(t, Criterion{Title: "Architecture", Weight: 100}, criterion)
	})

	t.Run("get grading uses assigned professor id", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:         project,
			criterionGrades: []CriterionGrade{{CriterionID: uuid.NewString()}},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		items, err := svc.GetGrading(context.Background(), userID, projectID)
		require.NoError(t, err)
		require.Len(t, items, 1)
	})

	t.Run("upsert grading sanitizes duplicates", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		criterionID := uuid.New()
		repo := &projectFlowRepos{
			project:         domain.Project{ID: projectID, Status: domain.ProjectReview, ProfessorID: &professorID},
			criterionGrades: []CriterionGrade{{CriterionID: criterionID.String()}},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		isMet := true

		items, err := svc.UpsertGrading(context.Background(), professorID, projectID, []CriterionGrade{
			{CriterionID: criterionID.String(), IsMet: &isMet, Comment: "  ok  "},
			{CriterionID: criterionID.String(), IsMet: &isMet, Comment: "  override  "},
		})
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Len(t, repo.upsertedGrades, 1)
		require.Equal(t, "override", repo.upsertedGrades[0].Comment)
	})

	t.Run("readiness computes activation state", func(t *testing.T) {
		authz := &projectFlowAuthz{}
		creatorID := uuid.New()
		projectReady := domain.Project{
			ID:                    projectID,
			Status:                domain.ProjectReview,
			CreatedBy:             creatorID,
			ProfessorID:           &professorID,
			ProfessorReviewStatus: "ACCEPTED",
		}
		posID := uuid.NewString()
		repo := &projectFlowRepos{
			project: projectReady,
			projectMembers: []Member{
				{UserID: creatorID.String(), Status: "ACTIVE"},
				{UserID: uuid.NewString(), Status: "ACTIVE", PositionID: &posID},
			},
			criteriaCount: 2,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		ready, err := svc.Readiness(context.Background(), projectID)
		require.NoError(t, err)
		require.True(t, ready.CanActivate)
		require.Equal(t, 1, ready.RequiredMembers)
		require.Equal(t, 1, ready.ActiveMembers)
	})

	t.Run("approve project activates ready project", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		creatorID := uuid.New()
		projectReady := domain.Project{
			ID:                    projectID,
			Status:                domain.ProjectReview,
			CreatedBy:             creatorID,
			ProfessorID:           &professorID,
			ProfessorReviewStatus: "ACCEPTED",
		}
		posID := uuid.NewString()
		repo := &projectFlowRepos{
			project: projectReady,
			projectMembers: []Member{
				{UserID: creatorID.String(), Status: "ACTIVE"},
				{UserID: uuid.NewString(), Status: "ACTIVE", PositionID: &posID},
			},
			criteriaCount: 1,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		got, readiness, err := svc.ApproveProject(context.Background(), userID, projectID)
		require.NoError(t, err)
		require.True(t, readiness.CanActivate)
		require.Equal(t, projectReady, got)
	})

	t.Run("submit for grading checks tasks", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:    domain.Project{ID: projectID, Status: domain.ProjectActive, ProfessorID: &professorID, ProfessorReviewStatus: "ACCEPTED"},
			tasksTotal: 3,
			tasksDone:  3,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		got, err := svc.SubmitProjectForGrading(context.Background(), userID, projectID)
		require.NoError(t, err)
		require.Equal(t, repo.project, got)
	})

	t.Run("publish grading rejects incomplete grading", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       project,
			criteriaCount: 3,
			gradedCount:   2,
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		_, err := svc.PublishGrading(context.Background(), professorID, projectID)
		require.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestTaskFlows(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	positionID := uuid.New()
	taskID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive, CreatedBy: userID}

	t.Run("task helper guards", func(t *testing.T) {
		svc := newProjectFlowService(&projectFlowRepos{project: domain.Project{Status: domain.ProjectReview}}, &projectFlowAuthz{}, &projectFlowGrantor{})
		require.ErrorIs(t, svc.ensureActiveProject(context.Background(), projectID), ErrProjectNotActive)

		status, ok := normalizeTaskStatus(" in_progress ")
		require.True(t, ok)
		require.Equal(t, "IN_PROGRESS", status)
	})

	t.Run("create task succeeds", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:          activeProject,
			positionCapacity: 2,
			createdTaskID:    taskID,
			task:             Task{ID: taskID.String(), Title: "Build API"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		task, err := svc.CreateTask(context.Background(), userID, projectID, " Build API ", " Implement handlers ", positionID, nil, nil)
		require.NoError(t, err)
		require.Equal(t, Task{ID: taskID.String(), Title: "Build API"}, task)
		require.Equal(t, "CREATED", repo.insertedEventType)
	})

	t.Run("list tasks delegates", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			listedTasks: []Task{{ID: uuid.NewString(), Title: "One"}},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		tasks, err := svc.ListTasks(context.Background(), userID, projectID)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
	})

	t.Run("update task status records activity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       activeProject,
			taskStatus:    "OPEN",
			taskTitle:     "Build API",
			updatedTaskID: taskID,
			task:          Task{ID: taskID.String(), Title: "Build API", Status: "DONE"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		task, err := svc.UpdateTaskStatus(context.Background(), userID, projectID, taskID, "done")
		require.NoError(t, err)
		require.Equal(t, Task{ID: taskID.String(), Title: "Build API", Status: "DONE"}, task)
		require.Equal(t, "STATUS_CHANGED", repo.insertedEventType)
		require.Equal(t, "DONE", repo.insertedToStatus)
	})

	t.Run("assign task verifies assignee position", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		assigneeID := uuid.New()
		repo := &projectFlowRepos{
			project:          activeProject,
			assignPositionID: positionID,
			assignPrevStatus: "OPEN",
			assignTaskTitle:  "Build API",
			position:         Position{Code: "BACKEND"},
			memberStatus:     "ACTIVE",
			memberPositionID: &positionID,
			assignedTaskID:   taskID,
			task:             Task{ID: taskID.String(), Title: "Build API"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		task, err := svc.AssignTask(context.Background(), userID, projectID, taskID, assigneeID)
		require.NoError(t, err)
		require.Equal(t, Task{ID: taskID.String(), Title: "Build API"}, task)
		require.Equal(t, "ASSIGNED", repo.insertedEventType)
	})

	t.Run("list task activities delegates", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			taskActivities: []TaskActivity{{TaskID: taskID.String(), EventType: "CREATED"}},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		activities, err := svc.ListTaskActivities(context.Background(), userID, projectID, &taskID)
		require.NoError(t, err)
		require.Len(t, activities, 1)
	})

	t.Run("complete task stores submission and marks done", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:            activeProject,
			completeAssigneeID: &userID,
			completeStatus:     "IN_PROGRESS",
			completeTitle:      "Build API",
			markDoneID:         taskID,
			task:               Task{ID: taskID.String(), Title: "Build API", Status: "DONE"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		task, err := svc.CompleteTask(context.Background(), userID, projectID, taskID, " done ", []string{" https://one ", "https://one"})
		require.NoError(t, err)
		require.Equal(t, Task{ID: taskID.String(), Title: "Build API", Status: "DONE"}, task)
		require.Equal(t, "done", repo.submissionComment)
		require.Equal(t, []string{"https://one"}, repo.submissionFiles)
		require.Equal(t, "COMPLETED", repo.insertedEventType)
	})

	t.Run("claim and delete task succeed", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:    activeProject,
			taskStatus: "OPEN",
			taskTitle:  "Build API",
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})

		require.NoError(t, svc.ClaimTask(context.Background(), userID, projectID, taskID))
		require.Equal(t, "CLAIMED", repo.insertedEventType)
		require.NoError(t, svc.DeleteTask(context.Background(), userID, projectID, taskID))
	})
}
