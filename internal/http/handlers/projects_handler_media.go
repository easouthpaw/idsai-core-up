package handlers

import (
	"errors"
	"io"
	"net/http"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/infra/images"
	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *ProjectsHandler) UploadImage(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read image"})
		return
	}
	defer f.Close()

	limited := io.LimitReader(f, images.MaxProjectCoverBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read image"})
		return
	}
	if len(raw) > images.MaxProjectCoverBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large"})
		return
	}

	project, err := h.svc.UpdateProjectImage(c.Request.Context(), userID, projectID, raw)
	if err != nil {
		writeProjectMediaError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ProjectResponseFromDomain(project))
}

func (h *ProjectsHandler) DeleteImage(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	project, err := h.svc.DeleteProjectImage(c.Request.Context(), userID, projectID)
	if err != nil {
		writeProjectMediaError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ProjectResponseFromDomain(project))
}

func writeProjectMediaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projects.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, projects.ErrStorage):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage unavailable"})
	case errors.Is(err, projects.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, images.ErrInvalidImageType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image format"})
	case errors.Is(err, images.ErrInvalidImageData):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image file"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project image"})
	}
}
