package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"idsai-core-up/internal/http/frontend"
)

var devFrontendFS = http.FS(frontend.Files)

func DevLandingPage(c *gin.Context) {
	c.FileFromFS("landing.html", devFrontendFS)
}

func DevAuthorPage(c *gin.Context) {
	c.FileFromFS("author.html", devFrontendFS)
}

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

func DevProfessorPage(c *gin.Context) {
	c.FileFromFS("professor.html", devFrontendFS)
}

func DevProfessorReviewsPage(c *gin.Context) {
	c.FileFromFS("professor-reviews.html", devFrontendFS)
}

func DevProfessorCriteriaPage(c *gin.Context) {
	c.FileFromFS("professor-criteria.html", devFrontendFS)
}

func DevProfessorGradingPage(c *gin.Context) {
	c.FileFromFS("professor-grading.html", devFrontendFS)
}
