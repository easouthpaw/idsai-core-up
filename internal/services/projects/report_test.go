package projects_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/services/projects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeReportSource struct {
	stacks     []string
	members    []projectflow.Member
	criteria   []projectflow.Criterion
	grades     []projectflow.CriterionGrade
	tasks      []projectflow.Task
	activities []projectflow.TaskActivity
}

type fakeReportStorage struct {
	objects   map[string][]byte
	available bool
	putErr    error
	getErr    error
	deleteErr error
}

func (f fakeReportSource) ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return f.stacks, nil
}

func (f fakeReportSource) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error) {
	return f.members, nil
}

func (f fakeReportSource) ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]projectflow.Criterion, error) {
	return f.criteria, nil
}

func (f fakeReportSource) ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]projectflow.CriterionGrade, error) {
	return f.grades, nil
}

func (f fakeReportSource) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]projectflow.Task, error) {
	return f.tasks, nil
}

func (f fakeReportSource) ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]projectflow.TaskActivity, error) {
	return f.activities, nil
}

func (f *fakeReportStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	if f.putErr != nil {
		return f.putErr
	}
	if f.objects == nil {
		f.objects = map[string][]byte{}
	}
	f.objects[key] = append([]byte(nil), body...)
	return nil
}

func (f *fakeReportStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	raw, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return append([]byte(nil), raw...), nil
}

func (f *fakeReportStorage) DeleteObject(ctx context.Context, key string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.objects, key)
	return nil
}

func (f *fakeReportStorage) PublicURL(key string) string {
	return ""
}

func (f *fakeReportStorage) Available() bool {
	return f.available
}

func TestService_GetProjectFinalReportPDF_ReturnsPDF(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	professorID := uuid.New()
	viewerID := uuid.New()
	facultyID := uuid.New()
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	isMet := true

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:             projectID,
			Title:          "AI Diploma Platform",
			Description:    strings.Repeat("Безопасный финальный отчёт. ", 20),
			Status:         domain.ProjectCompleted,
			IsPublic:       true,
			CreatedBy:      ownerID,
			CreatedByName:  "Айболат",
			CreatedByEmail: "aibolat.student@idsai.local",
			ProfessorID:    &professorID,
			FacultyID:      facultyID,
			Visibility:     "PUBLIC",
			RetakeCount:    1,
			CreatedAt:      now.Add(-72 * time.Hour),
			UpdatedAt:      now,
		},
		reviewSummary: &projects.ReviewSummary{
			Score:       "4.8",
			PassPercent: 95,
			Met:         2,
			Total:       2,
			ReviewedAt:  ptrTime(now),
			Reviewer:    "Проф. Сейтбек",
		},
	}

	svc := projects.NewService(repo, &fakeGrantor{})
	svc.SetReportSource(fakeReportSource{
		stacks: []string{"Go", "PostgreSQL", "Gin"},
		members: []projectflow.Member{
			{
				UserID:       uuid.NewString(),
				FullName:     "Марат",
				Email:        "marat.student@idsai.local",
				PositionName: ptrString("Backend"),
				Status:       "ACTIVE",
			},
		},
		criteria: []projectflow.Criterion{
			{ID: "c1", Title: "Сервис стабилен", Description: "Ошибок в проде нет", Weight: 60},
			{ID: "c2", Title: "Документация готова", Description: "README и демо-описание обновлены", Weight: 40},
		},
		grades: []projectflow.CriterionGrade{
			{CriterionID: "c1", IsMet: &isMet, Comment: "Прод выровнен и стабилен.", UpdatedAt: ptrTime(now)},
			{CriterionID: "c2", IsMet: &isMet, Comment: "Документация полная.", UpdatedAt: ptrTime(now)},
		},
		tasks: []projectflow.Task{
			{Title: "Собрать финальный билд", Status: "DONE", PositionName: "Backend"},
			{Title: "Проверить демо", Status: "DONE", PositionName: "QA"},
		},
		activities: []projectflow.TaskActivity{
			{EventType: "COMPLETED", Title: "Собрать финальный билд", Comment: "Сборка и smoke test завершены.", CreatedAt: now, ActorName: "Марат"},
		},
	})

	file, err := svc.GetProjectFinalReportPDF(context.Background(), projectID, viewerID, facultyID)
	require.NoError(t, err)
	require.Equal(t, "application/pdf", file.ContentType)
	require.Equal(t, "project-report-"+projectID.String()+".pdf", file.Filename)
	require.True(t, len(file.Data) > 2000)
	require.True(t, strings.HasPrefix(string(file.Data[:4]), "%PDF"))
}

func TestService_GetProjectFinalReportPDF_RequiresCompletedStatus(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	viewerID := uuid.New()
	facultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        projectID,
			Title:     "Draft",
			Status:    domain.ProjectActive,
			IsPublic:  true,
			CreatedBy: ownerID,
			FacultyID: facultyID,
		},
	}

	svc := projects.NewService(repo, &fakeGrantor{})
	svc.SetReportSource(fakeReportSource{})

	_, err := svc.GetProjectFinalReportPDF(context.Background(), projectID, viewerID, facultyID)
	require.ErrorIs(t, err, projects.ErrFinalReportUnavailable)
}

func TestService_GetProjectFinalReportPDF_RequiresReportSource(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	viewerID := uuid.New()
	facultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        projectID,
			Title:     "Complete",
			Status:    domain.ProjectCompleted,
			IsPublic:  true,
			CreatedBy: ownerID,
			FacultyID: facultyID,
		},
	}

	svc := projects.NewService(repo, &fakeGrantor{})

	_, err := svc.GetProjectFinalReportPDF(context.Background(), projectID, viewerID, facultyID)
	require.ErrorIs(t, err, projects.ErrReportSource)
}

func TestService_CaptureProjectFinalReport_StoresArtifactsAndReadUsesStoredPDF(t *testing.T) {
	projectID := uuid.New()
	ownerID := uuid.New()
	professorID := uuid.New()
	viewerID := uuid.New()
	facultyID := uuid.New()
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	isMet := true

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:             projectID,
			Title:          "Stored Report",
			Description:    "Финальный PDF сохраняется в storage.",
			Status:         domain.ProjectCompleted,
			IsPublic:       true,
			CreatedBy:      ownerID,
			CreatedByName:  "Айболат",
			CreatedByEmail: "aibolat.student@idsai.local",
			ProfessorID:    &professorID,
			FacultyID:      facultyID,
			Visibility:     "PUBLIC",
			RetakeCount:    2,
			CreatedAt:      now.Add(-72 * time.Hour),
			UpdatedAt:      now,
		},
		reviewSummary: &projects.ReviewSummary{
			Score:       "4.5",
			PassPercent: 90,
			Met:         1,
			Total:       1,
			ReviewedAt:  ptrTime(now),
			Reviewer:    "Проф. Сейтбек",
		},
	}

	storage := &fakeReportStorage{available: true, objects: map[string][]byte{}}
	svc := projects.NewService(repo, &fakeGrantor{})
	svc.SetStorage(storage)
	svc.SetReportSource(fakeReportSource{
		stacks:   []string{"Go"},
		members:  []projectflow.Member{{UserID: uuid.NewString(), FullName: "Марат", PositionName: ptrString("Backend"), Status: "ACTIVE"}},
		criteria: []projectflow.Criterion{{ID: "c1", Title: "Финал", Description: "Проект завершен", Weight: 100}},
		grades:   []projectflow.CriterionGrade{{CriterionID: "c1", IsMet: &isMet, Comment: "Готово", UpdatedAt: ptrTime(now)}},
		tasks:    []projectflow.Task{{Title: "Собрать PDF", Status: "DONE", PositionName: "Backend"}},
	})

	err := svc.CaptureProjectFinalReport(context.Background(), projectID, viewerID, facultyID)
	require.NoError(t, err)
	require.Contains(t, storage.objects, "projects/final-reports/"+projectID.String()+"/retake-02.pdf")
	require.Contains(t, storage.objects, "projects/final-reports/"+projectID.String()+"/retake-02.json")

	svc.SetReportSource(nil)
	file, err := svc.GetProjectFinalReportPDF(context.Background(), projectID, viewerID, facultyID)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(file.Data[:4]), "%PDF"))
}

func ptrString(value string) *string {
	return &value
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
