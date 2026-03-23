package handlers

import (
	"idsai-core-up/internal/services/auth"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func setSessionCookies(c *gin.Context, svc *auth.Service, tokens auth.Tokens) {
	setCookie(c, auth.AccessCookieName, tokens.AccessToken, int(svc.AccessTTL().Seconds()), "/", true)
	setCookie(c, auth.RefreshCookieName, tokens.RefreshToken, int(svc.RefreshTTL().Seconds()), "/", true)
}

func clearSessionCookies(c *gin.Context) {
	setCookie(c, auth.AccessCookieName, "", -1, "/", true)
	setCookie(c, auth.RefreshCookieName, "", -1, "/", true)
}

func setPasswordResetCookie(c *gin.Context, rawToken string, ttl time.Duration) {
	maxAge := int(ttl.Seconds())
	if maxAge < 0 {
		maxAge = -1
	}
	setCookie(c, auth.PasswordResetCookieName, rawToken, maxAge, "/v2/auth/password-reset", true)
}

func clearPasswordResetCookie(c *gin.Context) {
	setCookie(c, auth.PasswordResetCookieName, "", -1, "/v2/auth/password-reset", true)
}

func readCookie(c *gin.Context, name string) string {
	value, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func setCookie(c *gin.Context, name, value string, maxAge int, path string, httpOnly bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, path, "", requestUsesHTTPS(c), httpOnly)
}

func requestUsesHTTPS(c *gin.Context) bool {
	if c.Request == nil {
		return false
	}
	if c.Request.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https")
}

func authResponseNoStore(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}
