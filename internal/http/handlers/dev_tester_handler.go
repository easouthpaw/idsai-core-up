package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"idsai-core-up/internal/http/frontend"
)

var devFrontendFS = http.FS(frontend.Files)

func serveFrontendHTML(c *gin.Context, status int, name string) {
	data, err := frontend.Files.ReadFile(name)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load page")
		return
	}
	c.Data(status, "text/html; charset=utf-8", data)
}

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

func DevInvitesPage(c *gin.Context) {
	c.FileFromFS("invites.html", devFrontendFS)
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

func DevSettingsPage(c *gin.Context) {
	c.FileFromFS("settings.html", devFrontendFS)
}

func DevProfilePage(c *gin.Context) {
	c.FileFromFS("profile.html", devFrontendFS)
}

func DevGroupsPage(c *gin.Context) {
	c.FileFromFS("groups.html", devFrontendFS)
}

func DevNotFoundPage(c *gin.Context) {
	serveFrontendHTML(c, http.StatusNotFound, "404.html")
}

func DevKBPage(c *gin.Context) {
	c.FileFromFS("kb.html", devFrontendFS)
}

func DevKBArticlePage(c *gin.Context) {
	c.FileFromFS("kb-article.html", devFrontendFS)
}
