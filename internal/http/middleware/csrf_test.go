package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authsvc "idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
)

func TestCSRFProtectionRejectsCrossSiteCookieMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "http://idsai.local/v2/projects", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: authsvc.AccessCookieName, Value: "token", Path: "/"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}
}

func TestCSRFProtectionAllowsSameOriginCookieMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "http://idsai.local/v2/projects", nil)
	req.Header.Set("Origin", "http://idsai.local")
	req.AddCookie(&http.Cookie{Name: authsvc.AccessCookieName, Value: "token", Path: "/"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d", w.Code)
	}
}

func TestCSRFProtectionAllowsBearerStyleMutationWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "http://idsai.local/v2/projects", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d", w.Code)
	}
}

func csrfTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/v2/projects", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return r
}
