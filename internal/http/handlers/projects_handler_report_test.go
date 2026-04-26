package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type reportTestRepo struct {
	project       domain.Project
	reviewSummary *projects.ReviewSummary
}

func (r reportTestRepo) Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (r reportTestRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	return r.project, nil
}

func (r reportTestRepo) HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	return false, nil
}

func (r reportTestRepo) HasResolvedProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	return false, nil
}

func (r reportTestRepo) GetProjectReviewSummary(ctx context.Context, projectID uuid.UUID) (*projects.ReviewSummary, error) {
	return r.reviewSummary, nil
}

func (r reportTestRepo) SetProjectImage(ctx context.Context, projectID uuid.UUID, imageKey string, updatedAt time.Time) error {
	return nil
}

func (r reportTestRepo) ClearProjectImage(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (r reportTestRepo) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
	return nil, nil
}

func (r reportTestRepo) ListByFaculty(ctx context.Context, facultyID uuid.UUID, userID uuid.UUID) ([]domain.Project, error) {
	return nil, nil
}

func (r reportTestRepo) ListPublic(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	return nil, nil
}

func (r reportTestRepo) FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (r reportTestRepo) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]projects.Group, error) {
	return nil, nil
}

type reportTestGrantor struct{}

func (r reportTestGrantor) GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error {
	return nil
}

type reportTestSource struct {
	criteria []projectflow.Criterion
	grades   []projectflow.CriterionGrade
}

func (s reportTestSource) ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return []string{"Go", "Gin"}, nil
}

func (s reportTestSource) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error) {
	return []projectflow.Member{
		{UserID: uuid.NewString(), FullName: "Марат", PositionName: ptrMemberString("Backend"), Status: "ACTIVE"},
	}, nil
}

func (s reportTestSource) ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]projectflow.Criterion, error) {
	return s.criteria, nil
}

func (s reportTestSource) ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]projectflow.CriterionGrade, error) {
	return s.grades, nil
}

func (s reportTestSource) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]projectflow.Task, error) {
	return []projectflow.Task{{Title: "Собрать релиз", Status: "DONE", PositionName: "Backend"}}, nil
}

func (s reportTestSource) ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]projectflow.TaskActivity, error) {
	return []projectflow.TaskActivity{{EventType: "COMPLETED", Title: "Собрать релиз", Comment: "Готово", ActorName: "Марат", CreatedAt: time.Now().UTC()}}, nil
}

func TestProjectsHandler_DownloadFinalReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	ownerID := uuid.New()
	professorID := uuid.New()
	userID := uuid.New()
	facultyID := uuid.New()
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	isMet := true

	svc := projects.NewService(reportTestRepo{
		project: domain.Project{
			ID:             projectID,
			Title:          "Final Report Demo",
			Description:    "Финальный проект",
			Status:         domain.ProjectCompleted,
			IsPublic:       true,
			CreatedBy:      ownerID,
			CreatedByName:  "Айболат",
			CreatedByEmail: "aibolat.student@idsai.local",
			ProfessorID:    &professorID,
			FacultyID:      facultyID,
			CreatedAt:      now.Add(-24 * time.Hour),
			UpdatedAt:      now,
		},
		reviewSummary: &projects.ReviewSummary{
			Score:       "5.0",
			PassPercent: 100,
			Met:         1,
			Total:       1,
			ReviewedAt:  &now,
			Reviewer:    "Проф. Сейтбек",
		},
	}, reportTestGrantor{})
	svc.SetReportSource(reportTestSource{
		criteria: []projectflow.Criterion{{ID: "c1", Title: "Все готово", Description: "Финальная поставка завершена", Weight: 100}},
		grades:   []projectflow.CriterionGrade{{CriterionID: "c1", IsMet: &isMet, Comment: "Отлично", UpdatedAt: &now}},
	})

	handler := NewProjectsHandler(svc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("facultyID", facultyID)
		c.Next()
	})
	router.GET("/v2/projects/:project_id/final-report.pdf", handler.DownloadFinalReport)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/final-report.pdf?lang=ru", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Header().Get("Content-Disposition"), "project-report-"+projectID.String()+"-ru.pdf")
	require.True(t, strings.HasPrefix(rec.Body.String()[:4], "%PDF"))
}

func TestProjectsHandler_DownloadFinalReport_InlineDisposition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	ownerID := uuid.New()
	professorID := uuid.New()
	userID := uuid.New()
	facultyID := uuid.New()
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	isMet := true

	svc := projects.NewService(reportTestRepo{
		project: domain.Project{
			ID:             projectID,
			Title:          "Final Report Demo",
			Description:    "Финальный проект",
			Status:         domain.ProjectCompleted,
			IsPublic:       true,
			CreatedBy:      ownerID,
			CreatedByName:  "Айболат",
			CreatedByEmail: "aibolat.student@idsai.local",
			ProfessorID:    &professorID,
			FacultyID:      facultyID,
			CreatedAt:      now.Add(-24 * time.Hour),
			UpdatedAt:      now,
		},
		reviewSummary: &projects.ReviewSummary{
			Score:       "5.0",
			PassPercent: 100,
			Met:         1,
			Total:       1,
			ReviewedAt:  &now,
			Reviewer:    "Проф. Сейтбек",
		},
	}, reportTestGrantor{})
	svc.SetReportSource(reportTestSource{
		criteria: []projectflow.Criterion{{ID: "c1", Title: "Все готово", Description: "Финальная поставка завершена", Weight: 100}},
		grades:   []projectflow.CriterionGrade{{CriterionID: "c1", IsMet: &isMet, Comment: "Отлично", UpdatedAt: &now}},
	})

	handler := NewProjectsHandler(svc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("facultyID", facultyID)
		c.Next()
	})
	router.GET("/v2/projects/:project_id/final-report.pdf", handler.DownloadFinalReport)

	req := httptest.NewRequest(http.MethodGet, "/v2/projects/"+projectID.String()+"/final-report.pdf?disposition=inline&lang=en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Disposition"), "inline;")
	require.Contains(t, rec.Header().Get("Content-Disposition"), "project-report-"+projectID.String()+"-en.pdf")
	require.True(t, strings.HasPrefix(rec.Body.String()[:4], "%PDF"))
}

func ptrMemberString(value string) *string {
	return &value
}
