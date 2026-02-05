package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskDemoHandler struct{}

func NewTaskDemoHandler() *TaskDemoHandler { return &TaskDemoHandler{} }

func (h *TaskDemoHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}})
}

func (h *TaskDemoHandler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}
