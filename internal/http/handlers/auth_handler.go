package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/security/passwords"
	"idsai-core-up/internal/services/auth"
	rbacsvc "idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	svc   *auth.Service
	authz *rbacsvc.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) SetAuthorizer(authz *rbacsvc.Service) {
	h.authz = authz
}

type registerReq struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	FullName       string `json:"full_name"`
	DepartmentCode string `json:"department_code" binding:"required"`
	GroupCode      string `json:"group_code" binding:"required"`
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
	Email    string `json:"email"`
	Code     string `json:"code"`
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
	UserID         string   `json:"user_id"`
	TenantID       string   `json:"tenant_id"`
	FacultyID      string   `json:"faculty_id"`
	DepartmentID   string   `json:"department_id"`
	DepartmentCode string   `json:"department_code"`
	GroupID        string   `json:"group_id,omitempty"`
	GroupCode      string   `json:"group_code,omitempty"`
	GroupNumber    *int     `json:"group_number,omitempty"`
	Email          string   `json:"email"`
	PendingEmail   string   `json:"pending_email,omitempty"`
	PendingStatus  string   `json:"pending_email_status,omitempty"`
	FullName       string   `json:"full_name"`
	AvatarURL      string   `json:"avatar_url,omitempty"`
	Headline       string   `json:"headline,omitempty"`
	About          string   `json:"about,omitempty"`
	PreferredRole  string   `json:"preferred_role,omitempty"`
	Semester       string   `json:"semester,omitempty"`
	Availability   string   `json:"availability,omitempty"`
	Goals          string   `json:"goals,omitempty"`
	GithubURL      string   `json:"github_url,omitempty"`
	Telegram       string   `json:"telegram,omitempty"`
	PortfolioURL   string   `json:"portfolio_url,omitempty"`
	Stacks         []string `json:"stacks,omitempty"`
	Interests      []string `json:"interests,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
	IsAdmin        bool     `json:"is_admin"`
	IsProfessor    bool     `json:"is_professor"`
	EmailVerified  bool     `json:"email_verified"`
}

type capabilitiesResp struct {
	ScopeType   string   `json:"scope_type"`
	ScopeID     string   `json:"scope_id,omitempty"`
	Permissions []string `json:"permissions"`
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
	case errors.Is(err, passwords.ErrPasswordBlank),
		errors.Is(err, passwords.ErrPasswordTooShort),
		errors.Is(err, passwords.ErrPasswordNeedsLetter),
		errors.Is(err, passwords.ErrPasswordNeedsDigit):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, auth.ErrNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource not found"})
	case errors.Is(err, auth.ErrDepartmentNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "department not found"})
	case errors.Is(err, auth.ErrGroupNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "group not found"})
	case errors.Is(err, auth.ErrGroupMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "selected group does not belong to selected department"})
	case errors.Is(err, auth.ErrGroupUnchanged):
		c.JSON(http.StatusBadRequest, gin.H{"error": "requested group matches current group"})
	case errors.Is(err, auth.ErrPendingGroupRequestExists):
		c.JSON(http.StatusConflict, gin.H{"error": "pending group change request already exists"})
	case errors.Is(err, auth.ErrGroupRequestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "group change request not found"})
	case errors.Is(err, auth.ErrGroupRequestReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": "group change request already reviewed"})
	case errors.Is(err, auth.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, auth.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, auth.ErrEmailInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
	case errors.Is(err, auth.ErrNoPendingEmail):
		c.JSON(http.StatusBadRequest, gin.H{"error": "no pending email"})
	case errors.Is(err, auth.ErrInvalidCurrentPassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current password"})
	case errors.Is(err, auth.ErrStorageUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage unavailable"})
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
	return buildProfileResp(u, true)
}

func buildPublicProfileResp(u auth.User) meResp {
	return buildProfileResp(u, false)
}

func buildProfileResp(u auth.User, includePrivate bool) meResp {
	groupID := ""
	if u.GroupID != nil {
		groupID = u.GroupID.String()
	}
	resp := meResp{
		UserID:         u.ID.String(),
		TenantID:       u.TenantID.String(),
		FacultyID:      u.FacultyID.String(),
		DepartmentID:   u.DepartmentID.String(),
		DepartmentCode: strings.TrimSpace(u.DepartmentCode),
		GroupID:        groupID,
		GroupCode:      strings.TrimSpace(u.GroupCode),
		GroupNumber:    u.GroupNumber,
		Email:          u.Email,
		FullName:       u.FullName,
		AvatarURL:      strings.TrimSpace(u.AvatarURL),
		Headline:       strings.TrimSpace(u.Headline),
		About:          strings.TrimSpace(u.About),
		PreferredRole:  strings.TrimSpace(u.PreferredRole),
		Semester:       strings.TrimSpace(u.Semester),
		Availability:   strings.TrimSpace(u.Availability),
		Goals:          strings.TrimSpace(u.Goals),
		GithubURL:      strings.TrimSpace(u.GithubURL),
		Telegram:       strings.TrimSpace(u.Telegram),
		PortfolioURL:   strings.TrimSpace(u.PortfolioURL),
		Stacks:         append([]string(nil), u.Stacks...),
		Interests:      append([]string(nil), u.Interests...),
		UpdatedAt:      u.ProfileUpdatedAt.UTC().Format(time.RFC3339),
		IsAdmin:        u.IsAdmin,
		IsProfessor:    u.IsProfessor,
		EmailVerified:  u.EmailVerifiedAt != nil,
	}
	if includePrivate {
		resp.PendingEmail = strings.TrimSpace(u.PendingEmail)
		resp.PendingStatus = pendingEmailStatus(u)
	}
	return resp
}

func pendingEmailStatus(u auth.User) string {
	pending := strings.TrimSpace(u.PendingEmail)
	if pending == "" {
		return ""
	}
	if u.PendingEmailAt == nil {
		return "pending_verification"
	}
	return "verification_sent"
}

func parseCapabilitiesScope(c *gin.Context) (rbacsvc.Scope, error) {
	scopeType := strings.ToUpper(strings.TrimSpace(c.Query("scope_type")))
	scopeID := strings.TrimSpace(c.Query("scope_id"))

	if scopeType == "" {
		isAdmin, _ := middleware.IsAdminFromCtx(c)
		if isAdmin {
			return rbacsvc.Scope{Type: rbacsvc.ScopeSystem, ID: nil}, nil
		}
		if facultyID, ok := middleware.FacultyIDFromCtx(c); ok {
			return rbacsvc.Scope{Type: rbacsvc.ScopeFaculty, ID: &facultyID}, nil
		}
		return rbacsvc.Scope{}, rbacsvc.ErrInvalidScope
	}

	parseID := func(raw string) (*uuid.UUID, error) {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, rbacsvc.ErrInvalidScope
		}
		return &id, nil
	}

	switch scopeType {
	case string(rbacsvc.ScopeSystem):
		if scopeID != "" {
			return rbacsvc.Scope{}, rbacsvc.ErrInvalidScope
		}
		return rbacsvc.Scope{Type: rbacsvc.ScopeSystem, ID: nil}, nil
	case string(rbacsvc.ScopeTenant):
		id, err := parseID(scopeID)
		if err != nil {
			return rbacsvc.Scope{}, err
		}
		return rbacsvc.Scope{Type: rbacsvc.ScopeTenant, ID: id}, nil
	case string(rbacsvc.ScopeFaculty):
		id, err := parseID(scopeID)
		if err != nil {
			return rbacsvc.Scope{}, err
		}
		return rbacsvc.Scope{Type: rbacsvc.ScopeFaculty, ID: id}, nil
	case string(rbacsvc.ScopeDepartment):
		id, err := parseID(scopeID)
		if err != nil {
			return rbacsvc.Scope{}, err
		}
		return rbacsvc.Scope{Type: rbacsvc.ScopeDepartment, ID: id}, nil
	case string(rbacsvc.ScopeProject):
		id, err := parseID(scopeID)
		if err != nil {
			return rbacsvc.Scope{}, err
		}
		return rbacsvc.Scope{Type: rbacsvc.ScopeProject, ID: id}, nil
	default:
		return rbacsvc.Scope{}, rbacsvc.ErrInvalidScope
	}
}

func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	authResponseNoStore(c)

	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.RegisterStudent(
		c.Request.Context(),
		tenantCodeFromHeader(c),
		req.Email,
		req.Password,
		req.FullName,
		req.DepartmentCode,
		req.GroupCode,
	); err != nil {
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

	code := strings.TrimSpace(req.Code)
	email := strings.TrimSpace(req.Email)
	if code != "" && email != "" {
		if err := h.svc.ResetPasswordByCode(c.Request.Context(), tenantCodeFromHeader(c), email, code, req.Password); err != nil {
			writeAuthError(c, err)
			return
		}
		clearPasswordResetCookie(c)
		clearSessionCookies(c)
		c.JSON(http.StatusOK, authStatusResp{Status: "password_reset"})
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

func (h *AuthHandler) Capabilities(c *gin.Context) {
	authResponseNoStore(c)

	if h.authz == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rbac unavailable"})
		return
	}

	uid, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	scope, err := parseCapabilitiesScope(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}

	permissions, err := h.authz.ListPermissionCodes(c.Request.Context(), uid, scope)
	if err != nil {
		if errors.Is(err, rbacsvc.ErrInvalidScope) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load capabilities"})
		return
	}

	resp := capabilitiesResp{
		ScopeType:   string(scope.Type),
		Permissions: permissions,
	}
	if scope.ID != nil {
		resp.ScopeID = scope.ID.String()
	}

	c.JSON(http.StatusOK, resp)
}
