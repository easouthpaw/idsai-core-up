package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpx "idsai-core-up/internal/http"

	"github.com/stretchr/testify/require"
)

func TestDevTesterRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/login", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Login")
}

func TestDevProjectsRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/projects", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Projects")
}

func TestDevProjectRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/projects/00000000-0000-0000-0000-000000000001", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Project")
}

func TestDevAdminRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dev/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "IDSAI Corp. Admin")
}

func TestAuthorRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, "test-secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/author", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "Aibolat")
}
