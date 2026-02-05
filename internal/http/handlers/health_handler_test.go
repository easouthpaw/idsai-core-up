package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Unit_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := handlers.NewHealthHandler(nil)
	r.GET("/health", h.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"status":"ok"`)
}
