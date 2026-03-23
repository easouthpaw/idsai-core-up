package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerReq struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	FullName       string `json:"full_name"`
	DepartmentCode string `json:"department_code" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

type resendVerificationReq struct {
	Email string `json:"email" binding:"required,email"`
}

type passwordResetRequestReq struct {
	Email string `json:"email" binding:"required,email"`
}

type passwordResetConfirmReq struct {
	Token    string `json:"token"`
	Password string `json:"password" binding:"required"`
}

type authStatusResp struct {
	Status string `json:"status"`
}

type accessResp struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Status       string `json:"status,omitempty"`
}

type meResp struct {
	UserID        string `json:"user_id"`
	TenantID      string `json:"tenant_id"`
	FacultyID     string `json:"faculty_id"`
	DepartmentID  string `json:"department_id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	IsAdmin       bool   `json:"is_admin"`
	IsProfessor   bool   `json:"is_professor"`
	EmailVerified bool   `json:"email_verified"`
}

func tenantCodeFromHeader(c *gin.Context) string {
	code := strings.ToUpper(strings.TrimSpace(c.GetHeader("X-Tenant-Code")))
	if code == "" {
		return "CORE"
	}
	return code
}

func wantsTokenMode(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Auth-Mode")), "token")
}

func actorKey(c *gin.Context) string {
	return strings.TrimSpace(c.ClientIP())
}

func authRedirectURL(c *gin.Context, query string) string {
	url := "/dev/login"
	if query != "" {
		url += "?" + query
	}
	return url
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrTooManyAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
	case errors.Is(err, auth.ErrEmailVerificationRequired):
		c.JSON(http.StatusForbidden, gin.H{"error": "email verification required", "code": "email_verification_required"})
	case errors.Is(err, auth.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case errors.Is(err, auth.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, auth.ErrNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource not found"})
	case errors.Is(err, auth.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, auth.ErrSessionExpired), errors.Is(err, auth.ErrSessionInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session invalid"})
	case errors.Is(err, auth.ErrTokenExpired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
	case errors.Is(err, auth.ErrTokenInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "token invalid"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth operation failed"})
	}
}

func buildMeResp(u auth.User) meResp {
	return meResp{
		UserID:        u.ID.String(),
		TenantID:      u.TenantID.String(),
		FacultyID:     u.FacultyID.String(),
		DepartmentID:  u.DepartmentID.String(),
		Email:         u.Email,
		FullName:      u.FullName,
		IsAdmin:       u.IsAdmin,
		IsProfessor:   u.IsProfessor,
		EmailVerified: u.EmailVerifiedAt != nil,
	}
}

func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	authResponseNoStore(c)

	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.RegisterStudent(c.Request.Context(), tenantCodeFromHeader(c), req.Email, req.Password, req.FullName, req.DepartmentCode); err != nil {
		writeAuthError(c, err)
		return
	}

	clearSessionCookies(c)
	c.JSON(http.StatusAccepted, authStatusResp{Status: "verification_required"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	authResponseNoStore(c)

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	session, err := h.svc.Login(c.Request.Context(), actorKey(c), tenantCodeFromHeader(c), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	if wantsTokenMode(c) {
		c.JSON(http.StatusOK, accessResp{
			AccessToken:  session.Tokens.AccessToken,
			RefreshToken: session.Tokens.RefreshToken,
			Status:       "authenticated",
		})
		return
	}

	setSessionCookies(c, h.svc, session.Tokens)
	c.JSON(http.StatusOK, buildMeResp(session.User))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	authResponseNoStore(c)

	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		refreshToken = readCookie(c, auth.RefreshCookieName)
	}
	if refreshToken == "" {
		clearSessionCookies(c)
		c.Status(http.StatusNoContent)
		return
	}

	session, err := h.svc.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		clearSessionCookies(c)
		writeAuthError(c, err)
		return
	}

	if wantsTokenMode(c) {
		c.JSON(http.StatusOK, accessResp{
			AccessToken:  session.Tokens.AccessToken,
			RefreshToken: session.Tokens.RefreshToken,
			Status:       "refreshed",
		})
		return
	}

	setSessionCookies(c, h.svc, session.Tokens)
	c.JSON(http.StatusOK, authStatusResp{Status: "refreshed"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authResponseNoStore(c)

	var req logoutReq
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		refreshToken = readCookie(c, auth.RefreshCookieName)
	}
	_ = h.svc.Logout(c.Request.Context(), refreshToken)
	clearSessionCookies(c)
	clearPasswordResetCookie(c)
	c.JSON(http.StatusOK, authStatusResp{Status: "logged_out"})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	authResponseNoStore(c)

	var req resendVerificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.ResendVerification(c.Request.Context(), actorKey(c), tenantCodeFromHeader(c), req.Email); err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, authStatusResp{Status: "verification_sent"})
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	authResponseNoStore(c)

	var req passwordResetRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.RequestPasswordReset(c.Request.Context(), actorKey(c), tenantCodeFromHeader(c), req.Email); err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, authStatusResp{Status: "password_reset_sent"})
}

func (h *AuthHandler) PasswordResetLanding(c *gin.Context) {
	authResponseNoStore(c)

	rawToken := strings.TrimSpace(c.Query("token"))
	ttl, err := h.svc.ValidatePasswordResetToken(c.Request.Context(), rawToken)
	if err != nil {
		clearPasswordResetCookie(c)
		c.Redirect(http.StatusSeeOther, authRedirectURL(c, "reset=expired"))
		return
	}

	setPasswordResetCookie(c, rawToken, ttl)
	c.Redirect(http.StatusSeeOther, authRedirectURL(c, "reset=1"))
}

func (h *AuthHandler) PasswordResetConfirm(c *gin.Context) {
	authResponseNoStore(c)

	var req passwordResetConfirmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	rawToken := strings.TrimSpace(req.Token)
	if rawToken == "" {
		rawToken = readCookie(c, auth.PasswordResetCookieName)
	}
	if err := h.svc.ResetPassword(c.Request.Context(), rawToken, req.Password); err != nil {
		writeAuthError(c, err)
		return
	}

	clearPasswordResetCookie(c)
	clearSessionCookies(c)
	c.JSON(http.StatusOK, authStatusResp{Status: "password_reset"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	authResponseNoStore(c)

	if err := h.svc.VerifyEmail(c.Request.Context(), c.Query("token")); err != nil {
		c.Redirect(http.StatusSeeOther, authRedirectURL(c, "verified=0"))
		return
	}

	c.Redirect(http.StatusSeeOther, authRedirectURL(c, "verified=1"))
}

func (h *AuthHandler) Me(c *gin.Context) {
	authResponseNoStore(c)

	uid, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tid, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.svc.Me(c.Request.Context(), tid, uid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, buildMeResp(user))
}
