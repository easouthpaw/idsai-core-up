package handlers

import (
	"errors"
	"net/http"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type listUsersResp struct {
	Users []admin.User `json:"users"`
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

type setRoleReq struct {
	Role string `json:"role"`
}

type resetPasswordReq struct {
	Password string `json:"password"`
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
		if errors.Is(err, admin.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) SetRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req setRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	user, err := h.svc.SetUserRole(c.Request.Context(), userID, req.Role)
	if err != nil {
		if errors.Is(err, admin.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, admin.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.svc.ResetUserPassword(c.Request.Context(), userID, req.Password); err != nil {
		if errors.Is(err, admin.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, admin.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	adminUserID, ok := middleware.UserIDFromCtx(c)
	if ok && adminUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete current admin user"})
		return
	}

	if err := h.svc.DeleteUser(c.Request.Context(), userID); err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
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
		if errors.Is(err, admin.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		if errors.Is(err, admin.ErrDepartmentNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}
