package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDevTesterPagesServeFrontendHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/", DevLandingPage)
	router.GET("/author", DevAuthorPage)
	router.GET("/login", DevLoginPage)
	router.GET("/projects", DevProjectsPage)
	router.GET("/admin", DevAdminPage)
	router.GET("/project", DevProjectPage)
	router.GET("/invites", DevInvitesPage)
	router.GET("/professor", DevProfessorPage)
	router.GET("/professor/reviews", DevProfessorReviewsPage)
	router.GET("/professor/criteria", DevProfessorCriteriaPage)
	router.GET("/professor/grading", DevProfessorGradingPage)
	router.GET("/settings", DevSettingsPage)
	router.GET("/profile", DevProfilePage)
	router.GET("/groups", DevGroupsPage)
	router.GET("/kb", DevKBPage)
	router.GET("/kb/article", DevKBArticlePage)
	router.NoRoute(DevNotFoundPage)

	for _, path := range []string{
		"/",
		"/author",
		"/login",
		"/projects",
		"/admin",
		"/project",
		"/invites",
		"/professor",
		"/professor/reviews",
		"/professor/criteria",
		"/professor/grading",
		"/settings",
		"/profile",
		"/groups",
		"/kb",
		"/kb/article",
	} {
		requireStatus(t, router, http.MethodGet, path, "", http.StatusOK)
	}
	requireStatus(t, router, http.MethodGet, "/missing", "", http.StatusNotFound)
}
