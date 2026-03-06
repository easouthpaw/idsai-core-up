package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"idsai-core-up/internal/http/frontend"
)

var devFrontendFS = http.FS(frontend.Files)

func DevLoginPage(c *gin.Context) {
	c.FileFromFS("login.html", devFrontendFS)
}

func DevProjectsPage(c *gin.Context) {
	c.FileFromFS("projects.html", devFrontendFS)
}

func DevAdminPage(c *gin.Context) {
	c.FileFromFS("admin.html", devFrontendFS)
}

func DevProjectPage(c *gin.Context) {
	c.FileFromFS("project.html", devFrontendFS)
}
