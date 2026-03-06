package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger writes compact request logs and skips noisy polling/static endpoints.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if strings.TrimSpace(path) == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()

		if shouldSkipRequestLog(c.Request.Method, path, c.Request.URL.Path, status) {
			return
		}

		level := "INFO"
		if status >= 500 {
			level = "ERROR"
		} else if status >= 400 {
			level = "WARN"
		}

		log.Printf(
			"http level=%s status=%d method=%s path=%s latency=%s ip=%s",
			level,
			status,
			c.Request.Method,
			path,
			time.Since(start).Round(time.Microsecond),
			c.ClientIP(),
		)
	}
}

func shouldSkipRequestLog(method, fullPath, rawPath string, status int) bool {
	_ = status
	m := strings.ToUpper(strings.TrimSpace(method))
	fp := strings.TrimSpace(fullPath)
	rp := strings.TrimSpace(rawPath)

	if strings.HasPrefix(rp, "/dev/static/") {
		return true
	}
	if fp == "/health" {
		return true
	}

	// Skip high-frequency notification polling noise.
	if m == "GET" && (fp == "/v2/notifications" || fp == "/v2/notifications/unread-count" || rp == "/v2/notifications" || rp == "/v2/notifications/unread-count") {
		return true
	}
	return false
}
