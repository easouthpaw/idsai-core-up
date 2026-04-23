//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/requestctx"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestProjectFlowRepo_Integration_EndToEndProjectFlow(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	baseCtx := context.Background()
	pool, err := db.NewPool(baseCtx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	authRepo := postgres.NewAuthRepo(pool)
	repo := postgres.NewProjectFlowRepo(pool)
	projectsRepo := postgres.NewProjectsRepo(pool)

	scope := seedAuthScope(t, baseCtx, pool)
	groupID1, err := authRepo.CreateGroupInDepartment(baseCtx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode1, 201)
	require.NoError(t, err)
	groupID2, err := authRepo.CreateGroupInDepartment(baseCtx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode2, 202)
	require.NoError(t, err)

	ownerID := seedAuthUserProfile(t, baseCtx, pool, scope, "owner", groupID1)
	studentID := seedAuthUserProfile(t, baseCtx, pool, scope, "candidate student", groupID1)
	applicantID := seedAuthUserProfile(t, baseCtx, pool, scope, "applicant student", groupID1)
	professorID := seedAuthUserProfile(t, baseCtx, pool, scope, "professor", groupID2)
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, ownerID, "STUDENT", "FACULTY", &scope.facultyID))
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, studentID, "STUDENT", "FACULTY", &scope.facultyID))
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, applicantID, "STUDENT", "FACULTY", &scope.facultyID))
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, professorID, "PROFESSOR", "FACULTY", &scope.facultyID))

	ctx := requestctx.WithIdentity(baseCtx, ownerID, scope.tenantID, scope.facultyID, scope.departmentID)
	projectID, err := projectsRepo.CreateWithLeadRole(ctx, "Project Flow", "Integration project", scope.facultyID, "FACULTY", nil, ownerID)
	require.NoError(t, err)

	project, err := repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, projectID, project.ID)
	require.Equal(t, domain.ProjectDraft, project.Status)

	status, creatorID, err := repo.GetMemberStatusAndCreator(ctx, ownerID, projectID)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", status)
	require.Equal(t, ownerID, creatorID)

	hasRole, err := repo.HasProjectRole(ctx, ownerID, projectID, "TEAM_LEAD")
	require.NoError(t, err)
	require.True(t, hasRole)
	require.NoError(t, repo.RevokeProjectRole(ctx, ownerID, projectID, "TEAM_LEAD"))
	hasRole, err = repo.HasProjectRole(ctx, ownerID, projectID, "TEAM_LEAD")
	require.NoError(t, err)
	require.False(t, hasRole)

	require.NoError(t, repo.UpdateProject(ctx, projectID, true, "Project Flow Updated", true, "Updated description"))
	project, err = repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, "Project Flow Updated", project.Title)

	require.NoError(t, repo.ReplaceProjectStacks(ctx, projectID, []string{"go", "postgres", "vue"}))
	stackCodes, err := repo.ListProjectStackCodes(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, []string{"go", "postgres", "vue"}, stackCodes)

	backend, err := repo.CreateProjectPosition(ctx, projectID, "BACKEND", "Backend Developer", 2)
	require.NoError(t, err)
	frontend, err := repo.EnsureProjectPosition(ctx, projectID, "FRONTEND", "Frontend Developer", 1)
	require.NoError(t, err)
	frontend, err = repo.EnsureProjectPosition(ctx, projectID, "FRONTEND", "Frontend Engineer", 2)
	require.NoError(t, err)
	require.Equal(t, 2, frontend.Capacity)
	backendID := mustUUID(t, backend.ID)
	frontendID := mustUUID(t, frontend.ID)

	positions, err := repo.ListProjectPositions(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, positions, 2)
	position, err := repo.GetProjectPosition(ctx, projectID, backendID)
	require.NoError(t, err)
	require.Equal(t, "BACKEND", position.Code)
	capacity, err := repo.GetProjectPositionCapacity(ctx, projectID, frontendID)
	require.NoError(t, err)
	require.Equal(t, 2, capacity)
	totalCapacity, err := repo.SumProjectPositionCapacities(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, 4, totalCapacity)

	ok, err := repo.IsActiveStudentInFaculty(ctx, studentID, scope.facultyID)
	require.NoError(t, err)
	require.True(t, ok)
	candidates, err := repo.ListStudentCandidates(ctx, scope.facultyID, projectID, ownerID, ownerID, "candidate", 10)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Equal(t, studentID.String(), candidates[0].UserID)

	require.NoError(t, repo.OpenProjectRecruitment(ctx, projectID))
	project, err = repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, domain.ProjectRecruitment, project.Status)

	invited, err := repo.UpsertInvitedMember(ctx, projectID, studentID, ownerID, "join the team")
	require.NoError(t, err)
	require.Equal(t, "INVITED", invited.Status)
	invitedPosition, err := repo.GetInvitedMemberPosition(ctx, projectID, studentID)
	require.NoError(t, err)
	require.Nil(t, invitedPosition)

	studentCtx := requestctx.WithIdentity(baseCtx, studentID, scope.tenantID, scope.facultyID, scope.departmentID)
	incoming, err := repo.ListIncomingInvites(studentCtx, studentID, 10)
	require.NoError(t, err)
	require.Len(t, incoming, 1)
	require.Equal(t, "INVITED", incoming[0].Status)

	accepted, err := repo.RespondMemberInvite(studentCtx, projectID, studentID, true)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", accepted.Status)
	accepted, err = repo.SetActiveMemberPosition(ctx, projectID, studentID, backendID)
	require.NoError(t, err)
	require.Equal(t, backendID.String(), *accepted.PositionID)

	ok, err = repo.IsActiveProjectMember(ctx, studentID, projectID)
	require.NoError(t, err)
	require.True(t, ok)
	memberStatus, memberPosition, err := repo.GetProjectMemberStatusAndPosition(ctx, projectID, studentID)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", memberStatus)
	require.NotNil(t, memberPosition)
	require.Equal(t, backendID, *memberPosition)

	countByPosition, err := repo.CountActiveMembersByPosition(ctx, projectID, backendID, nil)
	require.NoError(t, err)
	require.Equal(t, 1, countByPosition)
	countByPosition, err = repo.CountActiveMembersByPosition(ctx, projectID, backendID, &studentID)
	require.NoError(t, err)
	require.Equal(t, 0, countByPosition)
	withPosition, err := repo.CountActiveMembersWithPosition(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, 1, withPosition)

	dueAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	taskID, err := repo.CreateTask(ctx, projectID, "Build API", "Implement endpoint", backendID, nil, "OPEN", ownerID, &dueAt)
	require.NoError(t, err)
	task, err := repo.GetTaskByID(ctx, projectID, taskID)
	require.NoError(t, err)
	require.Equal(t, "Build API", task.Title)
	require.Equal(t, "OPEN", task.Status)
	tasks, err := repo.ListProjectTasks(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.NoError(t, repo.EnsureTaskActivityLogAvailable(ctx))

	taskStatus, taskTitle, err := repo.GetTaskStatusAndTitle(ctx, projectID, taskID)
	require.NoError(t, err)
	require.Equal(t, "OPEN", taskStatus)
	require.Equal(t, "Build API", taskTitle)
	positionID, prevStatus, assignTitle, prevAssignee, err := repo.GetTaskAssignContext(ctx, projectID, taskID)
	require.NoError(t, err)
	require.Equal(t, backendID, positionID)
	require.Equal(t, "OPEN", prevStatus)
	require.Equal(t, "Build API", assignTitle)
	require.Nil(t, prevAssignee)

	assignedTaskID, err := repo.AssignTaskToUser(ctx, projectID, taskID, studentID)
	require.NoError(t, err)
	require.Equal(t, taskID, assignedTaskID)
	updatedTaskID, err := repo.UpdateTaskStatus(ctx, projectID, taskID, "OPEN")
	require.NoError(t, err)
	require.Equal(t, taskID, updatedTaskID)
	require.NoError(t, repo.ClaimTask(studentCtx, projectID, taskID, studentID))
	assigneeID, completeStatus, completeTitle, err := repo.GetTaskCompleteContext(ctx, projectID, taskID)
	require.NoError(t, err)
	require.NotNil(t, assigneeID)
	require.Equal(t, studentID, *assigneeID)
	require.Equal(t, "IN_PROGRESS", completeStatus)
	require.Equal(t, "Build API", completeTitle)

	require.NoError(t, repo.InsertTaskActivity(ctx, projectID, taskID, &studentID, "STATUS_CHANGED", "OPEN", "IN_PROGRESS", "Build API", "started", []string{"a.txt", "a.txt", "  ", strings.Repeat("x", 1200)}))
	activities, err := repo.ListProjectTaskActivities(ctx, projectID, &taskID)
	require.NoError(t, err)
	require.Len(t, activities, 1)
	require.Equal(t, "STATUS_CHANGED", activities[0].EventType)
	require.Len(t, activities[0].Attachments, 2)

	require.NoError(t, repo.UpsertTaskSubmission(ctx, projectID, taskID, studentID, "done", []string{"repo-url"}))
	doneTaskID, err := repo.MarkTaskDone(ctx, projectID, taskID)
	require.NoError(t, err)
	require.Equal(t, taskID, doneTaskID)
	totalTasks, doneTasks, err := repo.CountProjectTasksSummary(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, 1, totalTasks)
	require.Equal(t, 1, doneTasks)

	extraTaskID, err := repo.CreateTask(ctx, projectID, "Delete me", "", frontendID, nil, "OPEN", ownerID, nil)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteTask(ctx, projectID, extraTaskID))
	err = repo.DeleteTask(ctx, projectID, extraTaskID)
	require.ErrorIs(t, err, projectflow.ErrNotFound)

	applicantCtx := requestctx.WithIdentity(baseCtx, applicantID, scope.tenantID, scope.facultyID, scope.departmentID)
	applied, err := repo.UpsertAppliedMember(applicantCtx, projectID, applicantID, "please review")
	require.NoError(t, err)
	require.Equal(t, "APPLIED", applied.Status)
	ownerIncoming, err := repo.ListIncomingInvites(ctx, ownerID, 10)
	require.NoError(t, err)
	require.NotEmpty(t, ownerIncoming)
	outgoing, err := repo.ListOutgoingApplications(applicantCtx, applicantID, 10)
	require.NoError(t, err)
	require.Len(t, outgoing, 1)
	require.Equal(t, "APPLIED", outgoing[0].Status)

	rejected, err := repo.RejectProjectMemberApplication(ctx, projectID, applicantID)
	require.NoError(t, err)
	require.Equal(t, "REJECTED", rejected.Status)
	applied, err = repo.UpsertAppliedMember(applicantCtx, projectID, applicantID, "second try")
	require.NoError(t, err)
	require.Equal(t, "APPLIED", applied.Status)
	approved, err := repo.ApproveProjectMember(ctx, projectID, applicantID, &frontendID)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", approved.Status)

	members, err := repo.ListProjectMembers(ctx, projectID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(members), 2)

	removed, err := repo.RemoveProjectMember(ctx, projectID, studentID)
	require.NoError(t, err)
	require.Equal(t, "REMOVED", removed.Status)

	criterionA, err := repo.CreateProjectCriterion(ctx, projectID, ownerID, "Quality", "Code quality", 40)
	require.NoError(t, err)
	criterionB, err := repo.CreateProjectCriterion(ctx, projectID, ownerID, "Delivery", "Delivered scope", 60)
	require.NoError(t, err)
	_, err = repo.CreateProjectCriterion(ctx, projectID, ownerID, "Too much", "Overflow", 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, projectflow.ErrInvalidInput))

	criteria, err := repo.ListProjectCriteria(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, criteria, 2)
	criteriaCount, err := repo.CountProjectCriteria(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, 2, criteriaCount)

	met := true
	criterionAID := mustUUID(t, criterionA.ID)
	criterionBID := mustUUID(t, criterionB.ID)
	require.NoError(t, repo.UpsertProjectCriterionGrades(ctx, projectID, professorID, []projectflow.CriterionGradeUpsert{
		{CriterionID: criterionAID, IsMet: &met, Comment: "good"},
		{CriterionID: criterionBID, IsMet: nil, Comment: "needs review"},
	}))
	grades, err := repo.ListProjectCriterionGrades(ctx, projectID, professorID)
	require.NoError(t, err)
	require.Len(t, grades, 2)
	gradedCount, err := repo.CountProjectGradedCriteria(ctx, projectID, professorID)
	require.NoError(t, err)
	require.Equal(t, 1, gradedCount)
	err = repo.UpsertProjectCriterionGrades(ctx, projectID, professorID, []projectflow.CriterionGradeUpsert{{CriterionID: uuid.New(), IsMet: &met}})
	require.ErrorIs(t, err, projectflow.ErrInvalidInput)

	professorCtx := requestctx.WithIdentity(baseCtx, professorID, scope.tenantID, scope.facultyID, scope.departmentID)
	ok, err = repo.IsActiveProfessorInFaculty(ctx, professorID, scope.facultyID)
	require.NoError(t, err)
	require.True(t, ok)
	profCandidates, err := repo.ListProfessorCandidates(ctx, scope.facultyID, "professor", 10, ownerID, ownerID)
	require.NoError(t, err)
	require.Len(t, profCandidates, 1)
	profCandidate, err := repo.GetProfessorCandidateByID(ctx, professorID, scope.facultyID)
	require.NoError(t, err)
	require.Equal(t, professorID.String(), profCandidate.UserID)
	require.NoError(t, repo.AssignProjectProfessor(ctx, projectID, professorID))
	reviewInvites, err := repo.ListProfessorReviewInvites(professorCtx, professorID, "flow", 10)
	require.NoError(t, err)
	require.Len(t, reviewInvites, 1)
	project, err = repo.RespondProfessorInvite(professorCtx, projectID, professorID, true)
	require.NoError(t, err)
	require.Equal(t, "ACCEPTED", project.ProfessorReviewStatus)
	_, err = repo.RespondProfessorInvite(professorCtx, projectID, professorID, true)
	require.ErrorIs(t, err, projectflow.ErrNotFound)

	require.NoError(t, repo.ActivateProject(ctx, projectID))
	project, err = repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, domain.ProjectActive, project.Status)
	require.NoError(t, repo.MoveProjectToGrading(ctx, projectID))
	require.NoError(t, repo.ReturnProjectToActive(ctx, projectID))
	project, err = repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, domain.ProjectActive, project.Status)
	require.Equal(t, 1, project.RetakeCount)
	require.NoError(t, repo.MoveProjectToGrading(ctx, projectID))
	require.NoError(t, repo.MoveProjectToCompleted(ctx, projectID))
	project, err = repo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, domain.ProjectCompleted, project.Status)

	customRoleCode := "PF_CUSTOM_" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	accessRole, err := repo.CreateProjectAccessRole(ctx, projectID, ownerID, customRoleCode, "REVIEWER", "Reviewer", "Can review tasks", []string{"project.view", "task.create"})
	require.NoError(t, err)
	require.True(t, accessRole.Custom)
	accessRoles, err := repo.ListProjectAccessRoles(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, accessRoles, 1)
	require.NoError(t, repo.ReplaceAssignableRoles(ctx, applicantID, projectID, []string{customRoleCode}, []string{customRoleCode}))
	roleCodes, err := repo.ListProjectRoleCodes(ctx, applicantID, projectID)
	require.NoError(t, err)
	require.Contains(t, roleCodes, customRoleCode)
	require.NoError(t, repo.ReplaceAssignableRoles(ctx, applicantID, projectID, []string{customRoleCode}, nil))
	roleCodes, err = repo.ListProjectRoleCodes(ctx, applicantID, projectID)
	require.NoError(t, err)
	require.NotContains(t, roleCodes, customRoleCode)

	deleteProjectID, err := projectsRepo.CreateWithLeadRole(ctx, "Delete Owned", "temporary", scope.facultyID, "FACULTY", nil, ownerID)
	require.NoError(t, err)
	err = repo.DeleteOwnedProject(ctx, deleteProjectID, applicantID)
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.NoError(t, repo.DeleteOwnedProject(ctx, deleteProjectID, ownerID))
	err = repo.DeleteOwnedProject(ctx, deleteProjectID, ownerID)
	require.ErrorIs(t, err, projectflow.ErrNotFound)
}

func grantProjectFlowRole(ctx context.Context, pool *pgxpool.Pool, tenantID, userID uuid.UUID, roleCode, scopeType string, scopeID *uuid.UUID) error {
	_, err := pool.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT $1, $2, r.id, $3, $4
FROM roles r
WHERE r.code = $5
ON CONFLICT DO NOTHING;
`, tenantID, userID, scopeType, scopeID, roleCode)
	return err
}

func mustUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(raw)
	require.NoError(t, err)
	return id
}
