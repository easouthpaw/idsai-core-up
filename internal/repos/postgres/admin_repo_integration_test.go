//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/requestctx"
	adminsvc "idsai-core-up/internal/services/admin"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAdminRepo_Integration_UserAndProjectAdministration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	baseCtx := context.Background()
	pool, err := db.NewPool(baseCtx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	authRepo := postgres.NewAuthRepo(pool)
	adminRepo := postgres.NewAdminRepo(pool)
	projectRepo := postgres.NewProjectsRepo(pool)
	flowRepo := postgres.NewProjectFlowRepo(pool)

	scope := seedAuthScope(t, baseCtx, pool)
	groupID, err := authRepo.CreateGroupInDepartment(baseCtx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode1, 301)
	require.NoError(t, err)

	email := fmt.Sprintf("admin-created-%s@example.local", uuid.NewString())
	user, err := adminRepo.CreateUser(baseCtx, adminsvc.CreateUserParams{
		Email:          email,
		PasswordHash:   "admin-hash",
		FullName:       "Admin Created",
		DepartmentCode: scope.departmentCode,
		RoleCode:       adminsvc.RoleStudent,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID)
	require.Equal(t, "Admin Created", user.FullName)
	require.Equal(t, adminsvc.RoleStudent, user.RoleCode)
	require.Equal(t, scope.departmentCode, user.DepartmentCode)

	_, err = adminRepo.CreateUser(baseCtx, adminsvc.CreateUserParams{
		Email:          email,
		PasswordHash:   "duplicate",
		FullName:       "Duplicate",
		DepartmentCode: scope.departmentCode,
		RoleCode:       adminsvc.RoleStudent,
	})
	require.ErrorIs(t, err, adminsvc.ErrUserExists)

	_, err = adminRepo.CreateUser(baseCtx, adminsvc.CreateUserParams{
		Email:          fmt.Sprintf("admin-missing-%s@example.local", uuid.NewString()),
		PasswordHash:   "missing",
		FullName:       "Missing Department",
		DepartmentCode: "NO_SUCH_DEPT",
		RoleCode:       adminsvc.RoleStudent,
	})
	require.ErrorIs(t, err, adminsvc.ErrDepartmentNotFound)

	users, err := adminRepo.ListUsers(baseCtx, adminsvc.RoleStudent, email)
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, user.ID, users[0].ID)

	require.NoError(t, adminRepo.UpdateUserStatus(baseCtx, user.ID, adminsvc.StatusDisabled))
	user, err = adminRepo.GetUserByID(baseCtx, user.ID)
	require.NoError(t, err)
	require.Equal(t, adminsvc.StatusDisabled, user.Status)

	require.NoError(t, adminRepo.UpdateUserRole(baseCtx, user.ID, adminsvc.RoleProfessor))
	user, err = adminRepo.GetUserByID(baseCtx, user.ID)
	require.NoError(t, err)
	require.Equal(t, adminsvc.RoleProfessor, user.RoleCode)
	require.ErrorIs(t, adminRepo.UpdateUserRole(baseCtx, uuid.New(), adminsvc.RoleStudent), adminsvc.ErrUserNotFound)

	require.NoError(t, adminRepo.UpdateUserPasswordHash(baseCtx, user.ID, "new-admin-hash"))
	require.NoError(t, authRepo.InsertRefreshToken(baseCtx, scope.tenantID, user.ID, "admin-refresh-"+uuid.NewString(), time.Now().Add(time.Hour)))
	require.NoError(t, adminRepo.RevokeUserSessions(baseCtx, user.ID))

	ctx := requestctx.WithIdentity(baseCtx, user.ID, scope.tenantID, scope.facultyID, scope.departmentID)
	projectID, err := projectRepo.CreateWithLeadRole(ctx, "Admin Observed Project", "admin can inspect", scope.facultyID, "FACULTY", nil, user.ID)
	require.NoError(t, err)

	position, err := flowRepo.CreateProjectPosition(ctx, projectID, "QA", "QA Engineer", 1)
	require.NoError(t, err)
	positionID := mustUUID(t, position.ID)
	memberID := seedAuthUserProfile(t, baseCtx, pool, scope, "admin observed member", groupID)
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, memberID, "STUDENT", "FACULTY", &scope.facultyID))
	_, err = flowRepo.UpsertAppliedMember(ctx, projectID, memberID, "observe me")
	require.NoError(t, err)
	_, err = flowRepo.ApproveProjectMember(ctx, projectID, memberID, &positionID)
	require.NoError(t, err)
	taskID, err := flowRepo.CreateTask(ctx, projectID, "Observed Task", "admin report", positionID, &memberID, "DONE", user.ID, nil)
	require.NoError(t, err)
	require.NoError(t, flowRepo.InsertTaskActivity(ctx, projectID, taskID, &memberID, "STATUS_CHANGED", "OPEN", "DONE", "Observed Task", "finished", nil))
	criterion, err := flowRepo.CreateProjectCriterion(ctx, projectID, user.ID, "Observed Criterion", "admin report", 100)
	require.NoError(t, err)
	met := true
	require.NoError(t, flowRepo.UpsertProjectCriterionGrades(ctx, projectID, user.ID, []projectflow.CriterionGradeUpsert{
		{CriterionID: mustUUID(t, criterion.ID), IsMet: &met, Comment: "ok"},
	}))

	projects, err := adminRepo.ListProjects(ctx, "DRAFT", "Observed")
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, projectID, projects[0].ID)

	project, err := adminRepo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, "Admin Observed Project", project.Title)
	require.Equal(t, user.ID, project.CreatedBy)

	observation, err := adminRepo.GetProjectObservation(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, projectID, observation.Project.ID)
	require.Len(t, observation.Positions, 1)
	require.GreaterOrEqual(t, len(observation.Members), 2)
	require.Len(t, observation.Tasks, 1)
	require.Len(t, observation.Criteria, 1)
	require.GreaterOrEqual(t, observation.Summary.MembersActive, 2)
	require.Equal(t, 1, observation.Summary.TasksDone)
	require.Equal(t, 1, observation.Summary.CriteriaTotal)

	require.NoError(t, adminRepo.UpdateProjectStatus(ctx, projectID, "ACTIVE"))
	project, err = adminRepo.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", project.Status)
	require.ErrorIs(t, adminRepo.UpdateProjectStatus(ctx, uuid.New(), "ACTIVE"), adminsvc.ErrProjectNotFound)

	require.NoError(t, adminRepo.DeleteProject(ctx, projectID))
	_, err = adminRepo.GetProjectByID(ctx, projectID)
	require.ErrorIs(t, err, adminsvc.ErrProjectNotFound)
	require.ErrorIs(t, adminRepo.DeleteProject(ctx, projectID), adminsvc.ErrProjectNotFound)

	require.NoError(t, adminRepo.DeleteUser(baseCtx, user.ID))
	_, err = adminRepo.GetUserByID(baseCtx, user.ID)
	require.ErrorIs(t, err, adminsvc.ErrUserNotFound)
	require.ErrorIs(t, adminRepo.DeleteUser(baseCtx, user.ID), adminsvc.ErrUserNotFound)
}
