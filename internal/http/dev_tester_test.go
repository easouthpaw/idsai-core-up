package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpx "idsai-core-up/internal/http"

	"github.com/stretchr/testify/require"
)

func TestDevTesterRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/login", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Login")
}

func TestDevProjectsRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/projects", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Projects")
}

func TestDevProjectRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/projects/00000000-0000-0000-0000-000000000001", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Project")
}

func TestDevAdminRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Admin")
}

func TestAuthorRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/author", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "Айболат")
}

func TestDevProfessorRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/professor", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Professor Dashboard")
}

func TestDevProfessorReviewsRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/professor/reviews", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Review Requests")
}

func TestDevProfessorCriteriaRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/professor/criteria", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Criteria Setup")
}

func TestDevProfessorGradingRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/professor/grading", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Project Grading")
}

func TestDevGroupsRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/groups", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Groups")
}

func TestDevKBRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/kb", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. — База знаний")
	require.Contains(t, w.Body.String(), "/dev/static/js/auth-session.js")
}

func TestDevKBArticleRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/kb/article?id=00000000-0000-0000-0000-000000000001", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. — Статья")
	require.Contains(t, w.Body.String(), "/dev/static/js/auth-session.js")
}
