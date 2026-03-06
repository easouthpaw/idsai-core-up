package handlers

import (
	"errors"
	"net/http"

	"idsai-core-up/internal/services/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AdminHandler struct {
	svc *admin.Service
}

func NewAdminHandler(svc *admin.Service) *AdminHandler {
	return &AdminHandler{svc: svc}
}

type listUsersResp struct {
	Users []admin.User `json:"users"`
}

type listProjectsResp struct {
	Projects []admin.Project `json:"projects"`
}

type createUserReq struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FullName       string `json:"full_name"`
	DepartmentCode string `json:"department_code"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type setProjectStatusReq struct {
	Status string `json:"status"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	role := c.Query("role")
	search := c.Query("q")

	users, err := h.svc.ListUsers(c.Request.Context(), role, search)
	if err != nil {
		if errors.Is(err, admin.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, listUsersResp{Users: users})
}

func (h *AdminHandler) CreateStudent(c *gin.Context) {
	h.createByRole(c, admin.RoleStudent)
}

func (h *AdminHandler) CreateProfessor(c *gin.Context) {
	h.createByRole(c, admin.RoleProfessor)
}

func (h *AdminHandler) ListProjects(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("q")

	projects, err := h.svc.ListProjects(c.Request.Context(), status, search)
	if err != nil {
		if errors.Is(err, admin.ErrInvalidProjectStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, listProjectsResp{Projects: projects})
}

func (h *AdminHandler) SetStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req setStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.SetUserStatus(c.Request.Context(), userID, req.Status); err != nil {
		if errors.Is(err, admin.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) SetProjectStatus(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req setProjectStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.SetProjectStatus(c.Request.Context(), projectID, req.Status); err != nil {
		if errors.Is(err, admin.ErrInvalidProjectStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project status"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) createByRole(c *gin.Context, role string) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), admin.CreateUserInput{
		Email:          req.Email,
		Password:       req.Password,
		FullName:       req.FullName,
		DepartmentCode: req.DepartmentCode,
		RoleCode:       role,
	})
	if err != nil {
		if errors.Is(err, admin.ErrInvalidInput) || errors.Is(err, admin.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}
