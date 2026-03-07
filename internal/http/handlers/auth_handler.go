package handlers

import (
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
	Password       string `json:"password" binding:"required,min=6"`
	FullName       string `json:"full_name"`
	DepartmentCode string `json:"department_code" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type accessResp struct {
	AccessToken string `json:"access_token"`
}

type meResp struct {
	UserID       string `json:"user_id"`
	TenantID     string `json:"tenant_id"`
	FacultyID    string `json:"faculty_id"`
	DepartmentID string `json:"department_id"`
	IsAdmin      bool   `json:"is_admin"`
	IsProfessor  bool   `json:"is_professor"`
}

func tenantCodeFromHeader(c *gin.Context) string {
	code := strings.ToUpper(strings.TrimSpace(c.GetHeader("X-Tenant-Code")))
	if code == "" {
		return "CORE"
	}
	return code
}

// RegisterStudent godoc
// @Summary Register student
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerReq true "Register payload"
// @Success 201 {object} auth.Tokens
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	tokens, err := h.svc.RegisterStudent(c.Request.Context(), tenantCodeFromHeader(c), req.Email, req.Password, req.FullName, req.DepartmentCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tokens)
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginReq true "Login payload"
// @Success 200 {object} auth.Tokens
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	tokens, err := h.svc.Login(c.Request.Context(), tenantCodeFromHeader(c), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body refreshReq true "Refresh payload"
// @Success 200 {object} accessResp
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	access, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accessResp{AccessToken: access})
}

// Me godoc
// @Summary Current user (from JWT)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} meResp
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
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
	fid, ok := middleware.FacultyIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	didAny, ok := c.Get("departmentID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	isAdmin, _ := middleware.IsAdminFromCtx(c)
	isProfessor, _ := middleware.IsProfessorFromCtx(c)

	c.JSON(http.StatusOK, meResp{
		UserID:       uid.String(),
		TenantID:     tid.String(),
		FacultyID:    fid.String(),
		DepartmentID: didAny.(interface{ String() string }).String(),
		IsAdmin:      isAdmin,
		IsProfessor:  isProfessor,
	})
}
