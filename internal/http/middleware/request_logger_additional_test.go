package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLogger_LogsErrorsAndUsesRawPathWhenRouteUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	prevWriter := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(prevWriter)
	defer log.SetFlags(prevFlags)

	r := gin.New()
	r.Use(RequestLogger())
	r.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	r.ServeHTTP(w, req)

	if got := buf.String(); got == "" || !bytes.Contains([]byte(got), []byte("level=ERROR")) || !bytes.Contains([]byte(got), []byte("path=/missing")) {
		t.Fatalf("expected structured error log with raw path, got %q", got)
	}
}

func TestRequestLogger_SkipsNoisyEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	prevWriter := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(prevWriter)
	defer log.SetFlags(prevFlags)

	r := gin.New()
	r.Use(RequestLogger())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if got := buf.String(); got != "" {
		t.Fatalf("expected health request to be skipped, got %q", got)
	}
}
