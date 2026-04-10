package handlers

import (
	"errors"
	"net/http"
	"strings"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *ProjectsHandler) DownloadFinalReport(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	facultyID, ok := middleware.FacultyIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	file, err := h.svc.GetProjectFinalReportPDF(c.Request.Context(), projectID, userID, facultyID)
	if err != nil {
		switch {
		case errors.Is(err, projects.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		case errors.Is(err, domain.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, projects.ErrFinalReportUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": "final report is available only after grading is published"})
		case errors.Is(err, projects.ErrReportSource):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "final report service unavailable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Header("Cache-Control", "private, no-store")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	disposition := "attachment"
	if strings.EqualFold(strings.TrimSpace(c.Query("disposition")), "inline") {
		disposition = "inline"
	}
	c.Header("Content-Disposition", disposition+`; filename="`+file.Filename+`"`)
	c.Data(http.StatusOK, file.ContentType, file.Data)
}
