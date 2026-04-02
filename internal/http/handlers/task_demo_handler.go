package handlers

import (
	"net/http"

	"idsai-core-up/internal/http/dto"

	"github.com/gin-gonic/gin"
)

type TaskDemoHandler struct{}

func NewTaskDemoHandler() *TaskDemoHandler { return &TaskDemoHandler{} }

func (h *TaskDemoHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, dto.TaskDemoListResponse{Items: []string{}})
}

func (h *TaskDemoHandler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.StatusResponse{Status: "created"})
}
