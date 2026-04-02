package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/infra/images"
	"idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *AuthHandler) SettingsGet(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	user, err := h.svc.Me(c.Request.Context(), tenantID, userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.MeResponseFromUser(user))
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, _, ok := settingsActorIDs(c)
	if !ok {
		return
	}
	targetUserID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	user, err := h.svc.Me(c.Request.Context(), tenantID, targetUserID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PublicProfileResponseFromUser(user))
}

func (h *AuthHandler) SettingsUpdateProfile(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	user, err := h.svc.UpdateProfile(c.Request.Context(), tenantID, userID, auth.ProfileUpdate{
		FullName:      req.FullName,
		Headline:      req.Headline,
		About:         req.About,
		PreferredRole: req.PreferredRole,
		Semester:      req.Semester,
		Availability:  req.Availability,
		Goals:         req.Goals,
		GithubURL:     req.GithubURL,
		Telegram:      req.Telegram,
		PortfolioURL:  req.PortfolioURL,
		Stacks:        req.Stacks,
		Interests:     req.Interests,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.MeResponseFromUser(user))
}

func (h *AuthHandler) SettingsStartEmailChange(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	var req dto.StartEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.StartEmailChange(c.Request.Context(), actorKey(c), tenantID, userID, req.Email); err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, dto.AuthStatusResponse{Status: "verification_sent"})
}

func (h *AuthHandler) SettingsResendEmailChange(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}
	if err := h.svc.ResendEmailChange(c.Request.Context(), actorKey(c), tenantID, userID); err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, dto.AuthStatusResponse{Status: "verification_sent"})
}

func (h *AuthHandler) SettingsVerifyEmailChange(c *gin.Context) {
	authResponseNoStore(c)

	token := strings.TrimSpace(c.Query("token"))
	if _, err := h.svc.ConfirmEmailChange(c.Request.Context(), token); err != nil {
		c.Redirect(http.StatusSeeOther, "/dev/settings?email_change=expired")
		return
	}

	clearSessionCookies(c)
	c.Redirect(http.StatusSeeOther, "/dev/login?email_change=1")
}

func (h *AuthHandler) SettingsConfirmEmailChange(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	var req dto.ConfirmEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	var err error
	if strings.Contains(token, ".") {
		_, err = h.svc.ConfirmEmailChange(c.Request.Context(), token)
	} else {
		_, err = h.svc.ConfirmEmailChangeCode(c.Request.Context(), tenantID, userID, token)
	}
	if err != nil {
		writeAuthError(c, err)
		return
	}
	clearSessionCookies(c)
	c.JSON(http.StatusOK, dto.AuthStatusResponse{Status: "email_confirmed"})
}

func (h *AuthHandler) SettingsChangePassword(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password confirmation mismatch"})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), tenantID, userID, req.CurrentPassword, req.NewPassword); err != nil {
		writeAuthError(c, err)
		return
	}
	clearSessionCookies(c)
	c.JSON(http.StatusOK, dto.AuthStatusResponse{Status: "password_changed"})
}

func (h *AuthHandler) SettingsUploadAvatar(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read avatar"})
		return
	}
	defer f.Close()

	limited := io.LimitReader(f, images.MaxAvatarBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read avatar"})
		return
	}
	if len(raw) > images.MaxAvatarBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "avatar is too large"})
		return
	}

	user, err := h.svc.UpdateAvatar(c.Request.Context(), tenantID, userID, raw)
	if err != nil {
		switch {
		case errors.Is(err, images.ErrInvalidImageType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image format"})
		case errors.Is(err, images.ErrImageTooSmall):
			c.JSON(http.StatusBadRequest, gin.H{"error": "image is too small (minimum 400x400)"})
		case errors.Is(err, images.ErrInvalidImageData):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image file"})
		default:
			writeAuthError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, dto.MeResponseFromUser(user))
}

func (h *AuthHandler) SettingsDeleteAvatar(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}

	user, err := h.svc.DeleteAvatar(c.Request.Context(), tenantID, userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.MeResponseFromUser(user))
}

func settingsActorIDs(c *gin.Context) (tenantID uuid.UUID, userID uuid.UUID, ok bool) {
	tenantID, ok = middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, uuid.Nil, false
	}
	userID, ok = middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, userID, true
}
