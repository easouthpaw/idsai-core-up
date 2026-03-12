package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createTaskReq struct {
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	PositionID     string  `json:"position_id" binding:"required"`
	AssigneeUserID *string `json:"assignee_user_id,omitempty"`
	DueAt          *string `json:"due_at,omitempty"` // RFC3339
}

func (h *ProjectFlowHandler) CreateTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	positionID, err := uuid.Parse(strings.TrimSpace(req.PositionID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position_id"})
		return
	}

	var assignee *uuid.UUID
	if req.AssigneeUserID != nil && strings.TrimSpace(*req.AssigneeUserID) != "" {
		x, err := uuid.Parse(strings.TrimSpace(*req.AssigneeUserID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee_user_id"})
			return
		}
		assignee = &x
	}

	var dueAt *time.Time
	if req.DueAt != nil && strings.TrimSpace(*req.DueAt) != "" {
		tm, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.DueAt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_at (RFC3339 expected)"})
			return
		}
		dueAt = &tm
	}

	item, err := h.svc.CreateTask(c.Request.Context(), uid, pid, req.Title, req.Description, positionID, assignee, dueAt)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ProjectFlowHandler) ListTasks(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListTasks(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProjectFlowHandler) ListTaskActivities(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	rawTaskID := strings.TrimSpace(c.Query("task_id"))
	var taskID *uuid.UUID
	if rawTaskID != "" {
		tid, err := uuid.Parse(rawTaskID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
			return
		}
		taskID = &tid
	}

	items, err := h.svc.ListTaskActivities(c.Request.Context(), uid, pid, taskID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type updateTaskStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *ProjectFlowHandler) UpdateTaskStatus(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	var req updateTaskStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.UpdateTaskStatus(c.Request.Context(), uid, pid, tid, req.Status)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

type assignTaskReq struct {
	AssigneeUserID string `json:"assignee_user_id" binding:"required"`
}

func (h *ProjectFlowHandler) AssignTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	var req assignTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	assigneeID, err := uuid.Parse(strings.TrimSpace(req.AssigneeUserID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee_user_id"})
		return
	}

	item, err := h.svc.AssignTask(c.Request.Context(), uid, pid, tid, assigneeID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ProjectFlowHandler) ClaimTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	if err := h.svc.ClaimTask(c.Request.Context(), uid, pid, tid); err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "claimed"})
}

type completeTaskReq struct {
	Comment     string   `json:"comment"`
	Attachments []string `json:"attachments"`
}

func (h *ProjectFlowHandler) CompleteTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	var req completeTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.CompleteTask(c.Request.Context(), uid, pid, tid, req.Comment, req.Attachments)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}
