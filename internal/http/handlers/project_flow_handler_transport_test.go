package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type projectFlowTestDeps struct {
	can               bool
	permissions       []string
	project           domain.Project
	assignedProfessor projectflow.ProfessorCandidate
	incomingInvites   []projectflow.IncomingInvite
	member            projectflow.Member
}

func (d *projectFlowTestDeps) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope) (bool, error) {
	return d.can, nil
}

func (d *projectFlowTestDeps) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope rbac.Scope) (bool, error) {
	return d.can, nil
}

func (d *projectFlowTestDeps) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, attrs map[string]interface{}) (bool, error) {
	return d.can, nil
}

func (d *projectFlowTestDeps) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope) ([]string, error) {
	return d.permissions, nil
}

func (d *projectFlowTestDeps) GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error {
	return nil
}

func (d *projectFlowTestDeps) GetProjectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	return d.project, nil
}

func (d *projectFlowTestDeps) IsActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return false, nil
}

func (d *projectFlowTestDeps) HasProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) (bool, error) {
	return false, nil
}

func (d *projectFlowTestDeps) RevokeProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) error {
	return nil
}

func (d *projectFlowTestDeps) UpdateProject(ctx context.Context, projectID uuid.UUID, titleSet bool, title string, descriptionSet bool, description string) error {
	return nil
}

func (d *projectFlowTestDeps) OpenProjectRecruitment(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) ListStudentCandidates(ctx context.Context, facultyID, projectID, requesterUserID, projectOwnerID uuid.UUID, term string, limit int) ([]projectflow.StudentCandidate, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) ReplaceProjectStacks(ctx context.Context, projectID uuid.UUID, stackCodes []string) error {
	return nil
}

func (d *projectFlowTestDeps) ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) CreateProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (projectflow.Position, error) {
	return projectflow.Position{}, nil
}

func (d *projectFlowTestDeps) EnsureProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (projectflow.Position, error) {
	return projectflow.Position{}, nil
}

func (d *projectFlowTestDeps) ListProjectPositions(ctx context.Context, projectID uuid.UUID) ([]projectflow.Position, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) GetProjectPosition(ctx context.Context, projectID, positionID uuid.UUID) (projectflow.Position, error) {
	return projectflow.Position{}, nil
}

func (d *projectFlowTestDeps) GetProjectPositionCapacity(ctx context.Context, projectID, positionID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) SumProjectPositionCapacities(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) IsActiveStudentInFaculty(ctx context.Context, studentID, facultyID uuid.UUID) (bool, error) {
	return false, nil
}

func (d *projectFlowTestDeps) UpsertInvitedMember(ctx context.Context, projectID, studentID, invitedBy uuid.UUID, comment string) (projectflow.Member, error) {
	return projectflow.Member{}, nil
}

func (d *projectFlowTestDeps) UpsertAppliedMember(ctx context.Context, projectID, userID uuid.UUID, comment string) (projectflow.Member, error) {
	return projectflow.Member{}, nil
}

func (d *projectFlowTestDeps) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) CountActiveMembersByPosition(ctx context.Context, projectID, positionID uuid.UUID, excludeUserID *uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) GetProjectMemberStatusAndPosition(ctx context.Context, projectID, userID uuid.UUID) (string, *uuid.UUID, error) {
	return "", nil, nil
}

func (d *projectFlowTestDeps) CountActiveMembersWithPosition(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) ApproveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID, positionID *uuid.UUID) (projectflow.Member, error) {
	return d.member, nil
}

func (d *projectFlowTestDeps) RejectProjectMemberApplication(ctx context.Context, projectID, memberUserID uuid.UUID) (projectflow.Member, error) {
	return d.member, nil
}

func (d *projectFlowTestDeps) RemoveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID) (projectflow.Member, error) {
	return d.member, nil
}

func (d *projectFlowTestDeps) SetActiveMemberPosition(ctx context.Context, projectID, memberUserID, positionID uuid.UUID) (projectflow.Member, error) {
	return d.member, nil
}

func (d *projectFlowTestDeps) GetInvitedMemberPosition(ctx context.Context, projectID, userID uuid.UUID) (*uuid.UUID, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) RespondMemberInvite(ctx context.Context, projectID, userID uuid.UUID, accept bool) (projectflow.Member, error) {
	return d.member, nil
}

func (d *projectFlowTestDeps) ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]projectflow.IncomingInvite, error) {
	return d.incomingInvites, nil
}

func (d *projectFlowTestDeps) ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]projectflow.OutgoingApplication, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) ListProfessorCandidates(ctx context.Context, facultyID uuid.UUID, term string, limit int, requesterUserID, projectOwnerID uuid.UUID) ([]projectflow.ProfessorCandidate, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) IsActiveProfessorInFaculty(ctx context.Context, professorID, facultyID uuid.UUID) (bool, error) {
	return false, nil
}

func (d *projectFlowTestDeps) AssignProjectProfessor(ctx context.Context, projectID, professorID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) GetProfessorCandidateByID(ctx context.Context, professorID, facultyID uuid.UUID) (projectflow.ProfessorCandidate, error) {
	return d.assignedProfessor, nil
}

func (d *projectFlowTestDeps) RespondProfessorInvite(ctx context.Context, projectID, professorID uuid.UUID, accept bool) (domain.Project, error) {
	return domain.Project{}, nil
}

func (d *projectFlowTestDeps) ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, term string, limit int) ([]domain.Project, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) GetProjectCriteriaWeightSum(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) CreateProjectCriterion(ctx context.Context, projectID, userID uuid.UUID, title, description string, weight int) (projectflow.Criterion, error) {
	return projectflow.Criterion{}, nil
}

func (d *projectFlowTestDeps) ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]projectflow.Criterion, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]projectflow.CriterionGrade, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) UpsertProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID, items []projectflow.CriterionGradeUpsert) error {
	return nil
}

func (d *projectFlowTestDeps) CountProjectCriteria(ctx context.Context, projectID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) CountProjectGradedCriteria(ctx context.Context, projectID, professorID uuid.UUID) (int, error) {
	return 0, nil
}

func (d *projectFlowTestDeps) ActivateProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) CountProjectTasksSummary(ctx context.Context, projectID uuid.UUID) (int, int, error) {
	return 0, 0, nil
}

func (d *projectFlowTestDeps) MoveProjectToGrading(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) ReturnProjectToActive(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) MoveProjectToCompleted(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) DeleteOwnedProject(ctx context.Context, projectID, ownerID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) CreateTask(ctx context.Context, projectID uuid.UUID, title, description string, positionID uuid.UUID, assigneeUserID *uuid.UUID, status string, createdBy uuid.UUID, dueAt *time.Time) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *projectFlowTestDeps) GetTaskByID(ctx context.Context, projectID, taskID uuid.UUID) (projectflow.Task, error) {
	return projectflow.Task{}, nil
}

func (d *projectFlowTestDeps) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]projectflow.Task, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) EnsureTaskActivityLogAvailable(ctx context.Context) error {
	return nil
}

func (d *projectFlowTestDeps) GetTaskStatusAndTitle(ctx context.Context, projectID, taskID uuid.UUID) (string, string, error) {
	return "", "", nil
}

func (d *projectFlowTestDeps) UpdateTaskStatus(ctx context.Context, projectID, taskID uuid.UUID, status string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *projectFlowTestDeps) GetTaskAssignContext(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, string, string, *uuid.UUID, error) {
	return uuid.Nil, "", "", nil, nil
}

func (d *projectFlowTestDeps) AssignTaskToUser(ctx context.Context, projectID, taskID, assigneeUserID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *projectFlowTestDeps) ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]projectflow.TaskActivity, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) GetTaskCompleteContext(ctx context.Context, projectID, taskID uuid.UUID) (*uuid.UUID, string, string, error) {
	return nil, "", "", nil
}

func (d *projectFlowTestDeps) UpsertTaskSubmission(ctx context.Context, projectID, taskID, userID uuid.UUID, comment string, attachments []string) error {
	return nil
}

func (d *projectFlowTestDeps) MarkTaskDone(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *projectFlowTestDeps) ClaimTask(ctx context.Context, projectID, taskID, userID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) DeleteTask(ctx context.Context, projectID, taskID uuid.UUID) error {
	return nil
}

func (d *projectFlowTestDeps) InsertTaskActivity(ctx context.Context, projectID, taskID uuid.UUID, actorUserID *uuid.UUID, eventType, fromStatus, toStatus, title, comment string, attachments []string) error {
	return nil
}

func (d *projectFlowTestDeps) ListProjectRoleCodes(ctx context.Context, userID, projectID uuid.UUID) ([]string, error) {
	return []string{"PROJECT_LEAD"}, nil
}

func (d *projectFlowTestDeps) ReplaceAssignableRoles(ctx context.Context, userID, projectID uuid.UUID, assignableCodes []string, wantCodes []string) error {
	return nil
}

func (d *projectFlowTestDeps) GetMemberStatusAndCreator(ctx context.Context, userID, projectID uuid.UUID) (string, uuid.UUID, error) {
	return "ACTIVE", d.project.CreatedBy, nil
}

func (d *projectFlowTestDeps) ListProjectAccessRoles(ctx context.Context, projectID uuid.UUID) ([]projectflow.AccessCatalogItem, error) {
	return nil, nil
}

func (d *projectFlowTestDeps) CreateProjectAccessRole(ctx context.Context, projectID, createdBy uuid.UUID, roleCode, displayCode, name, description string, permissionCodes []string) (projectflow.AccessCatalogItem, error) {
	return projectflow.AccessCatalogItem{
		Code:            roleCode,
		DisplayCode:     displayCode,
		Name:            name,
		Description:     description,
		PermissionCodes: permissionCodes,
		Custom:          true,
	}, nil
}

func newProjectFlowHandlerForTest(deps *projectFlowTestDeps) *ProjectFlowHandler {
	svc := projectflow.NewService(deps, deps, deps, deps, deps, deps, deps, deps, deps, deps, deps)
	return NewProjectFlowHandler(svc)
}

func withProjectFlowUser(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func TestProjectFlowHandlerGetAccessCatalog_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	deps := &projectFlowTestDeps{can: true}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(uuid.New()))
	router.GET("/v2/projects/:project_id/access/catalog", handler.GetAccessCatalog)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/access/catalog", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListAccessCatalogResponse{Items: dto.ProjectFlowAccessCatalogItemResponsesFromService(projectflow.AssignableRoles)}), rec.Body.String())
}

func TestProjectFlowHandlerGetMemberAccess_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	callerID := uuid.New()
	targetUserID := uuid.New()
	deps := &projectFlowTestDeps{
		can:         true,
		permissions: []string{"task.create", "task.assign"},
		project: domain.Project{
			CreatedBy: uuid.New(),
		},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(callerID))
	router.GET("/v2/projects/:project_id/members/:user_id/access", handler.GetMemberAccess)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/members/"+targetUserID.String()+"/access", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ProjectFlowMemberAccessResponseFromService(&projectflow.MemberAccess{
		UserID:                   targetUserID.String(),
		RoleCodes:                []string{"PROJECT_LEAD"},
		ManagedRoleCodes:         []string{},
		EffectivePermissionCodes: deps.permissions,
	})), rec.Body.String())
}

func TestProjectFlowHandlerGetMemberAccess_DeniesBOLAForForeignMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	callerID := uuid.New()
	foreignUserID := uuid.New()
	deps := &projectFlowTestDeps{
		can: false,
		project: domain.Project{
			CreatedBy: uuid.New(),
		},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(callerID))
	router.GET("/v2/projects/:project_id/members/:user_id/access", handler.GetMemberAccess)

	// BOLA simulation: a regular member swaps the path user_id for another member's identifier.
	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/members/"+foreignUserID.String()+"/access", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.JSONEq(t, `{"error":"forbidden"}`, rec.Body.String())
}

func TestProjectFlowHandlerMyPermissions_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	deps := &projectFlowTestDeps{
		can:         true,
		permissions: []string{"task.create", "task.assign"},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(uuid.New()))
	router.GET("/v2/projects/:project_id/my-permissions", handler.MyPermissions)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/my-permissions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.MyPermissionsResponse{Permissions: deps.permissions}), rec.Body.String())
}

func TestProjectFlowHandlerGetAssignedProfessor_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	professorID := uuid.New()
	deps := &projectFlowTestDeps{
		can: true,
		project: domain.Project{
			ID:          projectID,
			FacultyID:   uuid.New(),
			CreatedBy:   uuid.New(),
			ProfessorID: &professorID,
		},
		assignedProfessor: projectflow.ProfessorCandidate{
			UserID:         professorID.String(),
			FullName:       "Professor DTO",
			Email:          "prof@example.edu",
			DepartmentCode: "CS",
		},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(uuid.New()))
	router.GET("/v2/projects/:project_id/professor", handler.GetAssignedProfessor)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/professor", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	expected := dto.AssignedProfessorResponse{
		Professor: &dto.ProjectFlowProfessorCandidateResponse{
			UserID:         deps.assignedProfessor.UserID,
			FullName:       deps.assignedProfessor.FullName,
			Email:          deps.assignedProfessor.Email,
			DepartmentCode: deps.assignedProfessor.DepartmentCode,
		},
	}
	require.JSONEq(t, mustJSON(t, expected), rec.Body.String())
}

func TestProjectFlowHandlerListIncomingInvites_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	deps := &projectFlowTestDeps{
		incomingInvites: []projectflow.IncomingInvite{
			{
				ProjectID:     uuid.NewString(),
				ProjectTitle:  "Core Upgrade",
				ProjectStatus: "RECRUITMENT",
				Status:        "PENDING",
				InviteComment: "Join us",
				CreatedAt:     time.Date(2026, 4, 1, 14, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(userID))
	router.GET("/v2/invites/incoming", handler.ListIncomingInvites)

	req := httptest.NewRequest(http.MethodGet, "/v2/invites/incoming", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListIncomingInvitesResponse{Items: dto.ProjectFlowIncomingInviteResponsesFromService(deps.incomingInvites)}), rec.Body.String())
}

func TestProjectFlowHandlerRemoveMember_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	memberUserID := uuid.New()
	deps := &projectFlowTestDeps{
		can: true,
		member: projectflow.Member{
			ID:        uuid.NewString(),
			ProjectID: projectID.String(),
			UserID:    memberUserID.String(),
			Status:    "REMOVED",
			CreatedAt: time.Date(2026, 4, 2, 9, 0, 0, 0, time.UTC),
		},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withProjectFlowUser(uuid.New()))
	router.DELETE("/v2/projects/:project_id/members/:user_id", handler.RemoveMember)

	req := httptest.NewRequest(http.MethodDelete, "/v2/projects/"+projectID.String()+"/members/"+memberUserID.String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ProjectFlowMemberResponseFromService(deps.member)), rec.Body.String())
}
