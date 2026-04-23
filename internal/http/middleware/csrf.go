package middleware

import (
	"net/http"
	"net/url"
	"strings"

	authsvc "idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
)

const csrfCheckHeader = "X-CSRF-Check"

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) || !hasSessionCookie(c) {
			c.Next()
			return
		}
		if strings.TrimSpace(c.GetHeader(csrfCheckHeader)) == "1" || requestHasSameOrigin(c) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf check failed"})
	}
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func hasSessionCookie(c *gin.Context) bool {
	for _, name := range []string{authsvc.AccessCookieName, authsvc.RefreshCookieName} {
		if value, err := c.Cookie(name); err == nil && strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func requestHasSameOrigin(c *gin.Context) bool {
	if sameOrigin(c.GetHeader("Origin"), c) {
		return true
	}
	return sameOrigin(c.GetHeader("Referer"), c)
}

func sameOrigin(raw string, c *gin.Context) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Scheme, requestScheme(c)) &&
		strings.EqualFold(parsed.Host, requestHost(c))
}

func requestScheme(c *gin.Context) string {
	if c.Request != nil && c.Request.TLS != nil {
		return "https"
	}
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		return strings.ToLower(strings.Split(proto, ",")[0])
	}
	return "http"
}

func requestHost(c *gin.Context) string {
	if host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host")); host != "" {
		return strings.TrimSpace(strings.Split(host, ",")[0])
	}
	if c.Request != nil {
		return strings.TrimSpace(c.Request.Host)
	}
	return ""
}
