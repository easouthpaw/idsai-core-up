package httpx

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
)

func registerProjectFlowRoutes(
	v2 *gin.RouterGroup,
	authMW gin.HandlerFunc,
	rbacSvc *rbac.Service,
	projectFlowH *handlers.ProjectFlowHandler,
) {
	if projectFlowH == nil {
		return
	}

	projectFlow := v2.Group("/projects/:project_id")
	projectFlow.Use(authMW)
	projectFlow.PATCH("", projectFlowH.UpdateProject)
	projectFlow.DELETE("", projectFlowH.DeleteProject)
	projectFlow.PUT("/stacks", projectFlowH.SetStacks)
	projectFlow.GET("/stacks",
		middleware.RequirePermission(rbacSvc, "grading.view", middleware.ProjectScopeFromParam("project_id")),
		projectFlowH.ListStacks,
	)
	projectFlow.POST("/recruitment/open", projectFlowH.OpenRecruitment)
	projectFlow.POST("/positions", projectFlowH.CreatePosition)
	projectFlow.GET("/positions",
		middleware.RequirePermission(rbacSvc, "grading.view", middleware.ProjectScopeFromParam("project_id")),
		projectFlowH.ListPositions,
	)
	projectFlow.GET("/candidates/students", projectFlowH.ListStudentCandidates)
	projectFlow.GET("/candidates/professors", projectFlowH.SearchProfessors)
	projectFlow.POST("/members/apply", projectFlowH.ApplyMember)
	projectFlow.POST("/members/invite", projectFlowH.InviteMember)
	projectFlow.POST("/members/respond", projectFlowH.RespondMemberInvite)
	projectFlow.GET("/members",
		middleware.RequirePermission(rbacSvc, "grading.view", middleware.ProjectScopeFromParam("project_id")),
		projectFlowH.ListMembers,
	)
	projectFlow.POST("/members/:user_id/approve", projectFlowH.ApproveMember)
	projectFlow.POST("/members/:user_id/reject", projectFlowH.RejectMember)
	projectFlow.PATCH("/members/:user_id/position", projectFlowH.SetMemberPosition)
	projectFlow.GET("/professor", projectFlowH.GetAssignedProfessor)
	projectFlow.POST("/professor", projectFlowH.AssignProfessor)
	projectFlow.POST("/professor/respond", projectFlowH.RespondProfessorInvite)
	projectFlow.POST("/criteria", projectFlowH.CreateCriterion)
	projectFlow.GET("/criteria",
		middleware.RequirePermission(rbacSvc, "grading.view", middleware.ProjectScopeFromParam("project_id")),
		projectFlowH.ListCriteria,
	)
	projectFlow.GET("/grading", projectFlowH.GetGrading)
	projectFlow.PUT("/grading", projectFlowH.UpsertGrading)
	projectFlow.POST("/grading/submit", projectFlowH.SubmitProjectForGrading)
	projectFlow.POST("/grading/publish", projectFlowH.PublishProjectGrading)
	projectFlow.GET("/readiness",
		middleware.RequirePermission(rbacSvc, "grading.view", middleware.ProjectScopeFromParam("project_id")),
		projectFlowH.Readiness,
	)
	projectFlow.POST("/approve", projectFlowH.ApproveProject)
	projectFlow.GET("/tasks", projectFlowH.ListTasks)
	projectFlow.GET("/tasks/activity", projectFlowH.ListTaskActivities)
	projectFlow.POST("/tasks", projectFlowH.CreateTask)
	projectFlow.PATCH("/tasks/:task_id/status", projectFlowH.UpdateTaskStatus)
	projectFlow.PATCH("/tasks/:task_id/assignee", projectFlowH.AssignTask)
	projectFlow.POST("/tasks/:task_id/claim", projectFlowH.ClaimTask)
	projectFlow.POST("/tasks/:task_id/complete", projectFlowH.CompleteTask)

	invites := v2.Group("/invites")
	invites.Use(authMW)
	invites.GET("/incoming", projectFlowH.ListIncomingInvites)
	invites.GET("/outgoing", projectFlowH.ListOutgoingApplications)

	professor := v2.Group("/professor")
	professor.Use(authMW)
	professor.GET("/review-invites", projectFlowH.ListProfessorReviewInvites)
}
