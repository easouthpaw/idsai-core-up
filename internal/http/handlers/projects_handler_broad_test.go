package handlers

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	notifsvc "idsai-core-up/internal/services/notifications"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type projectsHandlerRepo struct {
	createdID uuid.UUID
	createErr error

	createTitle      string
	createVisibility string
	createGroupID    *uuid.UUID
	createUserID     uuid.UUID

	project              domain.Project
	getErr               error
	hasProjectPermission bool
	hasResolvedPerm      bool
	reviewSummary        *projects.ReviewSummary
	reviewSummaryErr     error

	list        []domain.Project
	listErr     error
	listFaculty []domain.Project
	facultyErr  error
	listPublic  []domain.Project
	publicErr   error
	groupID     uuid.UUID
	groupErr    error
	groups      []projects.Group
	groupsErr   error

	setProjectImageErr   error
	clearProjectImageErr error
}

func (r *projectsHandlerRepo) Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	r.createTitle = title
	r.createVisibility = visibility
	r.createGroupID = groupID
	r.createUserID = createdBy
	if r.createdID == uuid.Nil {
		r.createdID = uuid.New()
	}
	return r.createdID, r.createErr
}

func (r *projectsHandlerRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	return r.project, r.getErr
}

func (r *projectsHandlerRepo) HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	return r.hasProjectPermission, nil
}

func (r *projectsHandlerRepo) HasResolvedProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	return r.hasResolvedPerm, nil
}

func (r *projectsHandlerRepo) GetProjectReviewSummary(ctx context.Context, projectID uuid.UUID) (*projects.ReviewSummary, error) {
	return r.reviewSummary, r.reviewSummaryErr
}

func (r *projectsHandlerRepo) SetProjectImage(ctx context.Context, projectID uuid.UUID, imageKey string, updatedAt time.Time) error {
	if r.setProjectImageErr != nil {
		return r.setProjectImageErr
	}
	r.project.ImageKey = imageKey
	r.project.ImageUpdatedAt = &updatedAt
	return nil
}

func (r *projectsHandlerRepo) ClearProjectImage(ctx context.Context, projectID uuid.UUID) error {
	if r.clearProjectImageErr != nil {
		return r.clearProjectImageErr
	}
	r.project.ImageKey = ""
	r.project.ImageUpdatedAt = nil
	return nil
}

func (r *projectsHandlerRepo) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
	return r.list, r.listErr
}

func (r *projectsHandlerRepo) ListByFaculty(ctx context.Context, facultyID uuid.UUID, userID uuid.UUID) ([]domain.Project, error) {
	return r.listFaculty, r.facultyErr
}

func (r *projectsHandlerRepo) ListPublic(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	return r.listPublic, r.publicErr
}

func (r *projectsHandlerRepo) FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error) {
	return r.groupID, r.groupErr
}

func (r *projectsHandlerRepo) GroupBelongsToFaculty(ctx context.Context, facultyID, groupID uuid.UUID) (bool, error) {
	if r.groupErr != nil {
		return false, r.groupErr
	}
	if r.groupID != uuid.Nil {
		return groupID == r.groupID, nil
	}
	return true, nil
}

func (r *projectsHandlerRepo) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]projects.Group, error) {
	return r.groups, r.groupsErr
}

type projectsHandlerGrantor struct {
	calls int
}

func (g *projectsHandlerGrantor) GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error {
	g.calls++
	return nil
}

type projectsHandlerNotifier struct {
	inputs []notifsvc.CreateInput
	err    error
}

func (n *projectsHandlerNotifier) Notify(ctx context.Context, in notifsvc.CreateInput) (notifsvc.Notification, error) {
	n.inputs = append(n.inputs, in)
	return notifsvc.Notification{ID: uuid.NewString()}, n.err
}

type projectsHandlerStorage struct {
	available bool
	putErr    error
	putKeys   []string
	deletes   []string
}

func (s *projectsHandlerStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	s.putKeys = append(s.putKeys, key)
	if s.putErr != nil {
		return s.putErr
	}
	return nil
}

func (s *projectsHandlerStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (s *projectsHandlerStorage) DeleteObject(ctx context.Context, key string) error {
	s.deletes = append(s.deletes, key)
	return nil
}

func (s *projectsHandlerStorage) PublicURL(key string) string {
	return "https://cdn.example.local/" + key
}

func (s *projectsHandlerStorage) Available() bool {
	return s.available
}

func TestProjectsHandlerReadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tenantID := uuid.New()
	facultyID := uuid.New()
	projectID := uuid.New()
	groupID := uuid.New()
	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	project := domain.Project{
		ID:                    projectID,
		Title:                 "IDS AI Core",
		Description:           "Diploma project",
		Status:                domain.ProjectDraft,
		IsPublic:              true,
		CreatedBy:             uuid.New(),
		CreatedByName:         "Lead Student",
		ProfessorReviewStatus: "NONE",
		FacultyID:             facultyID,
		Visibility:            "PUBLIC",
		DefaultCoverVariant:   2,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	repo := &projectsHandlerRepo{
		project:     project,
		list:        []domain.Project{project},
		listFaculty: []domain.Project{project},
		listPublic:  []domain.Project{project},
		groups: []projects.Group{
			{ID: groupID, Code: "CS-101", Name: "CS-101"},
		},
	}
	handler := NewProjectsHandler(projects.NewService(repo, &projectsHandlerGrantor{}))

	router := gin.New()
	router.Use(withProjectsActor(userID, tenantID, facultyID))
	router.GET("/project/:project_id", handler.Get)
	router.GET("/projects/my", handler.ListMine)
	router.GET("/projects/faculty", handler.ListFaculty)
	router.GET("/projects/public", handler.ListPublic)
	router.GET("/projects/groups", handler.ListGroups)

	requireStatus(t, router, http.MethodGet, "/project/"+projectID.String(), "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/project/not-a-uuid", "", http.StatusBadRequest)
	requireStatus(t, router, http.MethodGet, "/projects/my", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/projects/faculty", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/projects/public", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/projects/groups", "", http.StatusOK)

	noActorRouter := gin.New()
	noActorRouter.GET("/projects/my", handler.ListMine)
	requireStatus(t, noActorRouter, http.MethodGet, "/projects/my", "", http.StatusUnauthorized)
}

func TestProjectsHandlerCreateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tenantID := uuid.New()
	facultyID := uuid.New()
	projectID := uuid.New()
	groupID := uuid.New()
	repo := &projectsHandlerRepo{
		createdID: projectID,
		groupID:   groupID,
	}
	grantor := &projectsHandlerGrantor{}
	notifier := &projectsHandlerNotifier{}
	handler := NewProjectsHandler(projects.NewService(repo, grantor))
	handler.SetNotifier(notifier)

	router := gin.New()
	router.Use(withProjectsActor(userID, tenantID, facultyID))
	router.POST("/projects", handler.Create)

	requireStatus(t, router, http.MethodPost, "/projects", `{"title":"Private","description":"by code","visibility":"private","group_code":" cs-101 "}`, http.StatusCreated)
	require.Equal(t, "Private", repo.createTitle)
	require.Equal(t, "GROUP", repo.createVisibility)
	require.NotNil(t, repo.createGroupID)
	require.Equal(t, groupID, *repo.createGroupID)
	require.Equal(t, userID, repo.createUserID)
	require.Equal(t, 1, grantor.calls)
	require.Len(t, notifier.inputs, 1)
	require.Equal(t, "project.created", notifier.inputs[0].Type)
	require.Equal(t, projectID.String(), notifier.inputs[0].Payload["project_id"])

	requireStatus(t, router, http.MethodPost, "/projects", `{"title":"Public","visibility":"public","group_code":"CS-101"}`, http.StatusCreated)
	require.Equal(t, "PUBLIC", repo.createVisibility)
	require.Nil(t, repo.createGroupID)

	requireStatus(t, router, http.MethodPost, "/projects", `{"title":"Bad","visibility":"hidden"}`, http.StatusBadRequest)
	requireStatus(t, router, http.MethodPost, "/projects", `{"title":"Bad","visibility":"private"}`, http.StatusBadRequest)

	repo.groupErr = projects.ErrGroupNotFound
	requireStatus(t, router, http.MethodPost, "/projects", `{"title":"Bad","visibility":"private","group_code":"NOPE"}`, http.StatusBadRequest)
}

func TestProjectsHandlerImageRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tenantID := uuid.New()
	facultyID := uuid.New()
	projectID := uuid.New()
	now := time.Date(2026, 4, 21, 11, 0, 0, 0, time.UTC)
	repo := &projectsHandlerRepo{
		project: domain.Project{
			ID:                  projectID,
			Title:               "Image Project",
			Status:              domain.ProjectDraft,
			CreatedBy:           userID,
			FacultyID:           facultyID,
			Visibility:          "GROUP",
			ImageKey:            "projects/covers/old.jpg",
			DefaultCoverVariant: 1,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}
	storage := &projectsHandlerStorage{available: true}
	svc := projects.NewService(repo, &projectsHandlerGrantor{})
	svc.SetStorage(storage)
	handler := NewProjectsHandler(svc)

	router := gin.New()
	router.Use(withProjectsActor(userID, tenantID, facultyID))
	router.POST("/projects/:project_id/image", handler.UploadImage)
	router.DELETE("/projects/:project_id/image", handler.DeleteImage)

	req := multipartImageRequest(t, "/projects/"+projectID.String()+"/image", "image", "cover.jpg", testJPEG(t, 900, 900))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, storage.putKeys, 1)
	require.Contains(t, repo.project.ImageKey, projectID.String())
	require.Contains(t, storage.deletes, "projects/covers/old.jpg")

	requireStatus(t, router, http.MethodDelete, "/projects/"+projectID.String()+"/image", "", http.StatusOK)
	require.Empty(t, repo.project.ImageKey)
	require.Contains(t, storage.deletes, storage.putKeys[0])

	requireStatus(t, router, http.MethodPost, "/projects/not-a-uuid/image", "", http.StatusBadRequest)
	requireStatus(t, router, http.MethodPost, "/projects/"+projectID.String()+"/image", "", http.StatusBadRequest)

	badReq := multipartImageRequest(t, "/projects/"+projectID.String()+"/image", "image", "bad.txt", []byte("not an image"))
	badRec := httptest.NewRecorder()
	router.ServeHTTP(badRec, badReq)
	require.Equal(t, http.StatusBadRequest, badRec.Code, badRec.Body.String())
}

func withProjectsActor(userID, tenantID, facultyID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("tenantID", tenantID)
		c.Set("facultyID", facultyID)
		c.Next()
	}
}

func multipartImageRequest(t *testing.T, target, fieldName, fileName string, data []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func testJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x * 255) / max(width-1, 1)),
				G: uint8((y * 255) / max(height-1, 1)),
				B: 160,
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}
