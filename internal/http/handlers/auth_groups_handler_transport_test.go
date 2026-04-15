package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/infra/photon"
	"idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeInstitutionSuggester struct {
	items []auth.InstitutionSuggestion
	err   error
}

func (f *fakeInstitutionSuggester) Suggest(ctx context.Context, in auth.InstitutionSuggestRequest) ([]auth.InstitutionSuggestion, error) {
	return f.items, f.err
}

func newAuthHandlerWithRepo(repo *authHandlerRepo) *AuthHandler {
	return NewAuthHandler(auth.NewService(repo, auth.Config{
		JWTSecret: "01234567890123456789012345678901",
	}))
}

func TestAuthHandlerListDepartments_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
	repo := &authHandlerRepo{
		tenantID: uuid.New(),
		departments: []auth.Department{
			{
				ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				FacultyID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Code:      "CS",
				Name:      "Computer Science",
				ShortCode: "CS",
				CreatedAt: now,
			},
		},
	}

	handler := newAuthHandlerWithRepo(repo)
	router := gin.New()
	router.GET("/v2/auth/departments", handler.ListDepartments)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/departments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListDepartmentsResponse{Departments: dto.DepartmentResponsesFromService(repo.departments)}), rec.Body.String())
}

func TestAuthHandlerListFaculties_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 8, 30, 0, 0, time.UTC)
	repo := &authHandlerRepo{
		tenantID: uuid.New(),
		faculties: []auth.Faculty{
			{
				ID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				Code:      "IDSAI_ENU",
				Name:      "IDSAI ENU",
				CreatedAt: now,
			},
		},
	}

	handler := newAuthHandlerWithRepo(repo)
	router := gin.New()
	router.GET("/v2/auth/faculties", handler.ListFaculties)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/faculties", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListFacultiesResponse{Faculties: dto.FacultyResponsesFromService(repo.faculties)}), rec.Body.String())
}

func TestAuthHandlerSuggestInstitutions_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newAuthHandlerWithRepo(&authHandlerRepo{tenantID: uuid.New()})
	handler.SetInstitutionSuggester(&fakeInstitutionSuggester{
		items: []auth.InstitutionSuggestion{
			{
				Provider:   auth.InstitutionProviderPhoton,
				ExternalID: "school-1",
				Name:       "Школа-лицей №17",
				Address:    "Астана, ул. Абая, 1",
			},
		},
	})

	router := gin.New()
	router.GET("/v2/auth/institutions/suggest", handler.SuggestInstitutions)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/institutions/suggest?q=%D0%BB%D0%B8%D1%86%D0%B5%D0%B9&kind=SCHOOL", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.SuggestInstitutionsResponse{
		Items: dto.InstitutionSuggestionResponsesFromService([]auth.InstitutionSuggestion{
			{
				Provider:   auth.InstitutionProviderPhoton,
				ExternalID: "school-1",
				Name:       "Школа-лицей №17",
				Address:    "Астана, ул. Абая, 1",
			},
		}),
	}), rec.Body.String())
}

func TestAuthHandlerSuggestInstitutions_ReturnsPhotonUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newAuthHandlerWithRepo(&authHandlerRepo{tenantID: uuid.New()})
	handler.SetInstitutionSuggester(&fakeInstitutionSuggester{
		err: photon.ErrUnavailable,
	})

	router := gin.New()
	router.GET("/v2/auth/institutions/suggest", handler.SuggestInstitutions)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/institutions/suggest?q=%D0%B5%D0%B2%D1%80%D0%B0%D0%B7%D0%B8&kind=UNIVERSITY", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.JSONEq(t, `{"error":"Photon autocomplete is unavailable"}`, rec.Body.String())
}

func TestAuthHandlerSettingsListGroupChangeRequests_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC)
	repo := &authHandlerRepo{
		tenantID: uuid.New(),
		ownGroupRequests: []auth.GroupChangeRequest{
			{
				ID:               uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				StudentID:        uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				StudentName:      "Student One",
				StudentEmail:     "student@example.edu",
				CurrentGroupID:   uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				CurrentGroupCode: "SE-2201",
				RequestedGroupID: uuid.MustParse("66666666-6666-6666-6666-666666666666"),
				RequestedCode:    "SE-2202",
				Status:           "PENDING",
				CreatedAt:        now,
			},
		},
	}

	handler := newAuthHandlerWithRepo(repo)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenantID", repo.tenantID)
		c.Set("userID", uuid.MustParse("44444444-4444-4444-4444-444444444444"))
		c.Next()
	})
	router.GET("/v2/auth/settings/group-change-requests", handler.SettingsListGroupChangeRequests)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/settings/group-change-requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListGroupChangeRequestsResponse{Requests: dto.GroupChangeRequestResponsesFromService(repo.ownGroupRequests)}), rec.Body.String())
}

func TestAuthHandlerListDepartmentGroupsTree_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &authHandlerRepo{
		tenantID: uuid.New(),
		departmentGroupsTree: []auth.DepartmentGroupsTree{
			{
				ID:        uuid.MustParse("77777777-7777-7777-7777-777777777777"),
				Code:      "CS",
				Name:      "Computer Science",
				ShortCode: "CS",
				Groups: []auth.GroupNode{
					{
						ID:            uuid.MustParse("88888888-8888-8888-8888-888888888888"),
						GroupCode:     "CS-2201",
						GroupNumber:   2201,
						TotalStudents: 1,
						Students: []auth.GroupStudent{
							{
								UserID:    uuid.MustParse("99999999-9999-9999-9999-999999999999"),
								FullName:  "Student Tree",
								Email:     "tree@example.edu",
								AvatarURL: "/avatar.png",
								Status:    "ACTIVE",
								Role:      "STUDENT",
							},
						},
					},
				},
			},
		},
	}

	handler := newAuthHandlerWithRepo(repo)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenantID", repo.tenantID)
		c.Set("isAdmin", true)
		c.Next()
	})
	router.GET("/v2/auth/groups/tree", handler.ListDepartmentGroupsTree)

	req := httptest.NewRequest(http.MethodGet, "/v2/auth/groups/tree", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListDepartmentGroupsTreeResponse{Departments: dto.DepartmentGroupsTreeResponsesFromService(repo.departmentGroupsTree)}), rec.Body.String())
}
