package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/services/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type adminHandlerRepo struct {
	users       []admin.User
	projects    []admin.Project
	createdUser admin.User
	project     admin.Project
	observation admin.ProjectObservation
}

func (f *adminHandlerRepo) ListUsers(ctx context.Context, roleCode, search string) ([]admin.User, error) {
	return f.users, nil
}

func (f *adminHandlerRepo) ListProjects(ctx context.Context, status, search string) ([]admin.Project, error) {
	return f.projects, nil
}

func (f *adminHandlerRepo) GetProjectObservation(ctx context.Context, projectID uuid.UUID) (admin.ProjectObservation, error) {
	return f.observation, nil
}

func (f *adminHandlerRepo) CreateUser(ctx context.Context, in admin.CreateUserParams) (admin.User, error) {
	return f.createdUser, nil
}

func (f *adminHandlerRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (admin.User, error) {
	if f.createdUser.ID != uuid.Nil {
		f.createdUser.ID = userID
		return f.createdUser, nil
	}
	return admin.User{}, admin.ErrUserNotFound
}

func (f *adminHandlerRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return nil
}

func (f *adminHandlerRepo) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode string) error {
	return nil
}

func (f *adminHandlerRepo) UpdateUserPasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

func (f *adminHandlerRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (f *adminHandlerRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	return nil
}

func (f *adminHandlerRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (f *adminHandlerRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (f *adminHandlerRepo) GetProjectByID(ctx context.Context, projectID uuid.UUID) (admin.Project, error) {
	if f.project.ID != uuid.Nil {
		f.project.ID = projectID
		return f.project, nil
	}
	return admin.Project{}, admin.ErrProjectNotFound
}

func TestAdminHandlerListUsers_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)

	repo := &adminHandlerRepo{
		users: []admin.User{
			{
				ID:             uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				FullName:       "Alice Student",
				Email:          "alice@example.edu",
				RoleCode:       admin.RoleStudent,
				Status:         admin.StatusActive,
				FacultyCode:    "IDSAI",
				DepartmentCode: "CS",
				CreatedAt:      now.Add(-time.Hour),
				UpdatedAt:      now,
			},
		},
	}
	handler := NewAdminHandler(admin.NewService(repo))
	router := gin.New()
	router.GET("/v2/admin/users", handler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/v2/admin/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListUsersResponse{Users: dto.AdminUserResponsesFromService(repo.users)}), rec.Body.String())
}

func TestAdminHandlerCreateStudent_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &adminHandlerRepo{
		createdUser: admin.User{
			ID:             uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			FullName:       "Bob Student",
			Email:          "bob@example.edu",
			RoleCode:       admin.RoleStudent,
			Status:         admin.StatusPending,
			FacultyCode:    "IDSAI",
			DepartmentCode: "SE",
		},
	}
	handler := NewAdminHandler(admin.NewService(repo))
	router := gin.New()
	router.POST("/v2/admin/users/students", handler.CreateStudent)

	body := []byte(`{"email":"bob@example.edu","password":"Password123","full_name":"Bob Student","department_code":"SE"}`)
	req := httptest.NewRequest(http.MethodPost, "/v2/admin/users/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.AdminUserResponseFromService(repo.createdUser)), rec.Body.String())
}

func TestAdminHandlerObserveProject_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	projectID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	assigneeID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	repo := &adminHandlerRepo{
		observation: admin.ProjectObservation{
			Project: admin.Project{
				ID:             projectID,
				Title:          "Core Platform",
				Description:    "Platform refresh",
				Status:         "ACTIVE",
				Visibility:     "PUBLIC",
				IsPublic:       true,
				CreatedBy:      uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				AuthorName:     "Team Lead",
				AuthorEmail:    "lead@example.edu",
				FacultyCode:    "IDSAI",
				DepartmentCode: "CS",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			Positions: []admin.ProjectPosition{
				{ID: uuid.MustParse("66666666-6666-6666-6666-666666666666"), Code: "BE", Name: "Backend", Capacity: 2},
			},
			Members: []admin.ProjectMember{
				{
					UserID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
					FullName:     "Member One",
					Email:        "member@example.edu",
					RoleCode:     "STUDENT",
					Status:       "ACTIVE",
					PositionCode: "BE",
					PositionName: "Backend",
				},
			},
			Tasks: []admin.ProjectTask{
				{
					ID:             uuid.MustParse("88888888-8888-8888-8888-888888888888"),
					Title:          "Implement API",
					Status:         "TODO",
					PositionCode:   "BE",
					AssigneeUserID: &assigneeID,
					AssigneeName:   "Member One",
					UpdatedAt:      now,
				},
			},
			Criteria: []admin.ProjectCriterion{
				{
					ID:        uuid.MustParse("99999999-9999-9999-9999-999999999999"),
					Title:     "Docs",
					Weight:    20,
					CreatedBy: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
					CreatedAt: now,
				},
			},
			Summary: admin.ProjectObservationSummary{
				MembersTotal:   1,
				MembersActive:  1,
				TasksTotal:     1,
				CriteriaTotal:  1,
				MembersApplied: 0,
				MembersInvited: 0,
				TasksDone:      0,
			},
		},
	}
	handler := NewAdminHandler(admin.NewService(repo))
	router := gin.New()
	router.GET("/v2/admin/projects/:project_id/observe", handler.ObserveProject)

	req := httptest.NewRequest(http.MethodGet, "/v2/admin/projects/"+projectID.String()+"/observe", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.AdminProjectObservationResponseFromService(repo.observation)), rec.Body.String())
}

func TestAdminHandlerUserMutationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	adminID := uuid.New()
	repo := &adminHandlerRepo{
		createdUser: admin.User{
			ID:             userID,
			FullName:       "Managed User",
			Email:          "managed@example.edu",
			RoleCode:       admin.RoleProfessor,
			Status:         admin.StatusActive,
			FacultyCode:    "IDSAI",
			DepartmentCode: "CS",
		},
	}
	handler := NewAdminHandler(admin.NewService(repo))
	router := gin.New()
	router.Use(withProjectFlowUser(adminID))
	router.POST("/v2/admin/users/professors", handler.CreateProfessor)
	router.PUT("/v2/admin/users/:user_id/status", handler.SetStatus)
	router.PUT("/v2/admin/users/:user_id/role", handler.SetRole)
	router.PUT("/v2/admin/users/:user_id/password", handler.ResetPassword)
	router.DELETE("/v2/admin/users/:user_id", handler.DeleteUser)

	requireStatus(t, router, http.MethodPost, "/v2/admin/users/professors", `{"email":"managed@example.edu","password":"Password123","full_name":"Managed User","department_code":"CS"}`, http.StatusCreated)
	requireStatus(t, router, http.MethodPut, "/v2/admin/users/"+userID.String()+"/status", `{"status":"disabled"}`, http.StatusNoContent)
	requireStatus(t, router, http.MethodPut, "/v2/admin/users/"+userID.String()+"/role", `{"role":"student"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPut, "/v2/admin/users/"+userID.String()+"/password", `{"password":"Password123"}`, http.StatusNoContent)
	requireStatus(t, router, http.MethodDelete, "/v2/admin/users/"+userID.String(), "", http.StatusNoContent)
	requireStatus(t, router, http.MethodDelete, "/v2/admin/users/"+adminID.String(), "", http.StatusBadRequest)
}

func TestAdminHandlerProjectMutationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	projectID := uuid.New()
	adminID := uuid.New()
	repo := &adminHandlerRepo{
		project: admin.Project{
			ID:             projectID,
			Title:          "Admin Project",
			Description:    "demo",
			Status:         "COMPLETED",
			Visibility:     "FACULTY",
			CreatedBy:      uuid.New(),
			AuthorName:     "Lead",
			AuthorEmail:    "lead@example.edu",
			FacultyCode:    "IDSAI",
			DepartmentCode: "CS",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		projects: []admin.Project{
			{
				ID:          projectID,
				Title:       "Admin Project",
				Description: "demo",
				Status:      "COMPLETED",
				Visibility:  "FACULTY",
				CreatedBy:   uuid.New(),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
	handler := NewAdminHandler(admin.NewService(repo))
	router := gin.New()
	router.Use(withFlowContext(adminID, uuid.New(), uuid.New()))
	router.GET("/v2/admin/projects", handler.ListProjects)
	router.PUT("/v2/admin/projects/:project_id/status", handler.SetProjectStatus)
	router.DELETE("/v2/admin/projects/:project_id", handler.DeleteProject)

	requireStatus(t, router, http.MethodGet, "/v2/admin/projects?status=completed&q=admin", "", http.StatusOK)
	requireStatus(t, router, http.MethodPut, "/v2/admin/projects/"+projectID.String()+"/status", `{"status":"archive"}`, http.StatusNoContent)
	requireStatus(t, router, http.MethodDelete, "/v2/admin/projects/"+projectID.String(), "", http.StatusNoContent)
}
