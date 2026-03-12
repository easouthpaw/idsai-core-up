package handlers

import (
	"errors"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type listProjectsResp struct {
	Projects []admin.Project `json:"projects"`
}

type setProjectStatusReq struct {
	Status string `json:"status"`
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

	project, err := h.svc.SetProjectStatus(c.Request.Context(), projectID, req.Status)
	if err != nil {
		if errors.Is(err, admin.ErrInvalidProjectStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, admin.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project status"})
		return
	}

	tenantID, tenantOK := middleware.TenantIDFromCtx(c)
	adminUserID, adminOK := middleware.UserIDFromCtx(c)
	if tenantOK && adminOK {
		status := strings.ToUpper(strings.TrimSpace(req.Status))
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			adminUserID,
			"project.status.changed",
			"Статус проекта изменён",
			"Статус проекта успешно обновлён администратором.",
			map[string]any{
				"project_id": projectID.String(),
				"title":      project.Title,
				"status":     status,
			},
			false,
		))

		if project.CreatedBy != adminUserID {
			body := "Статус вашего проекта был обновлён администратором."
			if status == "ARCHIVE" {
				body = "Ваш проект был закрыт (ARCHIVE) администратором."
			}
			notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
				tenantID,
				project.CreatedBy,
				"project.status.changed",
				"Статус проекта обновлён",
				body,
				map[string]any{
					"project_id": projectID.String(),
					"title":      project.Title,
					"status":     status,
				},
				true,
			))
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) DeleteProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if err := h.svc.DeleteProject(c.Request.Context(), projectID); err != nil {
		if errors.Is(err, admin.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) ObserveProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	ob, err := h.svc.ObserveProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, admin.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to observe project"})
		return
	}

	c.JSON(http.StatusOK, ob)
}
