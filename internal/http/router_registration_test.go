package httpx

import (
	"net/http"
	"testing"

	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
)

func TestRegisterAuthRoutes_WithHandlerRegistersPublicAndProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v2 := router.Group("/v2")
	registerAuthRoutes(v2, noopAuthMW, nil, &handlers.AuthHandler{})

	paths := registeredRouteSet(router)
	for _, route := range []string{
		http.MethodGet + " /v2/auth/faculties",
		http.MethodGet + " /v2/auth/departments/:department_code/groups",
		http.MethodPost + " /v2/auth/register",
		http.MethodGet + " /v2/auth/me",
		http.MethodPatch + " /v2/auth/settings/profile",
		http.MethodPost + " /v2/auth/settings/group-change-requests",
		http.MethodGet + " /v2/auth/admin/group-change-requests",
		http.MethodPost + " /v2/auth/admin/group-change-requests/:request_id/review",
	} {
		if !paths[route] {
			t.Fatalf("route %s was not registered", route)
		}
	}

	before := len(router.Routes())
	registerAuthRoutes(v2, noopAuthMW, nil, nil)
	if got := len(router.Routes()); got != before {
		t.Fatalf("nil auth handler should not register routes: before=%d after=%d", before, got)
	}
}

func TestRegisterProjectFlowRoutes_WithHandlerRegistersWorkspaceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v2 := router.Group("/v2")
	registerProjectFlowRoutes(v2, noopAuthMW, nil, &handlers.ProjectFlowHandler{})

	paths := registeredRouteSet(router)
	for _, route := range []string{
		http.MethodPatch + " /v2/projects/:project_id",
		http.MethodPost + " /v2/projects/:project_id/recruitment/open",
		http.MethodPost + " /v2/projects/:project_id/members/invite",
		http.MethodPost + " /v2/projects/:project_id/professor/respond",
		http.MethodPut + " /v2/projects/:project_id/grading",
		http.MethodPost + " /v2/projects/:project_id/tasks/:task_id/complete",
		http.MethodGet + " /v2/projects/:project_id/access/catalog",
		http.MethodPut + " /v2/projects/:project_id/members/:user_id/access",
		http.MethodGet + " /v2/invites/incoming",
		http.MethodGet + " /v2/professor/review-invites",
	} {
		if !paths[route] {
			t.Fatalf("route %s was not registered", route)
		}
	}

	before := len(router.Routes())
	registerProjectFlowRoutes(v2, noopAuthMW, nil, nil)
	if got := len(router.Routes()); got != before {
		t.Fatalf("nil project flow handler should not register routes: before=%d after=%d", before, got)
	}
}

func noopAuthMW(c *gin.Context) {
	c.Next()
}

func registeredRouteSet(router *gin.Engine) map[string]bool {
	out := make(map[string]bool)
	for _, route := range router.Routes() {
		out[route.Method+" "+route.Path] = true
	}
	return out
}
