package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterDevAndDocsRoutes_HeadRootReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerDevAndDocsRoutes(r, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
