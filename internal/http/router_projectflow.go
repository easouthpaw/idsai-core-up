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

	enforce := rbacFeatureEnabled("RBAC_ENFORCE_PROJECTFLOW", true)
	requireProject := func(permission string) gin.HandlerFunc {
		return middleware.RequirePermissionIf(enforce && rbacSvc != nil, rbacSvc, permission, middleware.ProjectScopeFromParam("project_id"))
	}
	requireFaculty := func(permission string) gin.HandlerFunc {
		return middleware.RequirePermissionIf(enforce && rbacSvc != nil, rbacSvc, permission, middleware.FacultyScopeFromCtx())
	}

	projectFlow := v2.Group("/projects/:project_id")
	projectFlow.Use(authMW)
	projectFlow.PATCH("", requireProject("project.edit"), projectFlowH.UpdateProject)
	projectFlow.DELETE("", requireProject("project.delete"), projectFlowH.DeleteProject)
	projectFlow.PUT("/stacks", requireProject("project.edit"), projectFlowH.SetStacks)
	projectFlow.GET("/stacks", requireProject("project.view"), projectFlowH.ListStacks)
	projectFlow.POST("/recruitment/open", requireProject("project.edit"), projectFlowH.OpenRecruitment)
	projectFlow.POST("/positions", requireProject("position.create"), projectFlowH.CreatePosition)
	projectFlow.GET("/positions", requireProject("project.view"), projectFlowH.ListPositions)
	projectFlow.GET("/candidates/students", requireProject("member.approve"), projectFlowH.ListStudentCandidates)
	projectFlow.GET("/candidates/professors", requireProject("project.invite_professor"), projectFlowH.SearchProfessors)
	projectFlow.POST("/members/apply", requireProject("member.apply"), projectFlowH.ApplyMember)
	projectFlow.POST("/members/invite", requireProject("member.approve"), projectFlowH.InviteMember)
	projectFlow.POST("/members/respond", requireProject("member.apply"), projectFlowH.RespondMemberInvite)
	projectFlow.GET("/members", requireProject("project.view"), projectFlowH.ListMembers)
	projectFlow.POST("/members/:user_id/approve", requireProject("member.approve"), projectFlowH.ApproveMember)
	projectFlow.POST("/members/:user_id/reject", requireProject("member.approve"), projectFlowH.RejectMember)
	projectFlow.PATCH("/members/:user_id/position", requireProject("member.approve"), projectFlowH.SetMemberPosition)
	projectFlow.GET("/professor", requireProject("project.view"), projectFlowH.GetAssignedProfessor)
	projectFlow.POST("/professor", requireProject("project.invite_professor"), projectFlowH.AssignProfessor)
	projectFlow.POST("/professor/respond", requireProject("project.review.respond"), projectFlowH.RespondProfessorInvite)
	projectFlow.POST("/criteria", requireProject("project.set_criteria"), projectFlowH.CreateCriterion)
	projectFlow.GET("/criteria", requireProject("grading.view"), projectFlowH.ListCriteria)
	projectFlow.GET("/grading", requireProject("grading.view"), projectFlowH.GetGrading)
	projectFlow.PUT("/grading", requireProject("grading.mark_criteria"), projectFlowH.UpsertGrading)
	projectFlow.POST("/grading/submit", requireProject("project.submit_for_review"), projectFlowH.SubmitProjectForGrading)
	projectFlow.POST("/grading/publish", requireProject("grading.publish"), projectFlowH.PublishProjectGrading)
	projectFlow.GET("/readiness", requireProject("project.view"), projectFlowH.Readiness)
	projectFlow.POST("/approve", requireProject("project.approve"), projectFlowH.ApproveProject)
	projectFlow.GET("/tasks", requireProject("task.view"), projectFlowH.ListTasks)
	projectFlow.GET("/tasks/activity", requireProject("task.view"), projectFlowH.ListTaskActivities)
	projectFlow.POST("/tasks", requireProject("task.create"), projectFlowH.CreateTask)
	projectFlow.PATCH("/tasks/:task_id/status", requireProject("task.update"), projectFlowH.UpdateTaskStatus)
	projectFlow.PATCH("/tasks/:task_id/assignee", requireProject("task.assign"), projectFlowH.AssignTask)
	projectFlow.POST("/tasks/:task_id/claim", requireProject("task.claim"), projectFlowH.ClaimTask)
	projectFlow.POST("/tasks/:task_id/complete", requireProject("task.update"), projectFlowH.CompleteTask)
	projectFlow.GET("/access/catalog", requireProject("member.access.manage"), projectFlowH.GetAccessCatalog)
	projectFlow.GET("/members/:user_id/access", requireProject("member.access.manage"), projectFlowH.GetMemberAccess)
	projectFlow.PUT("/members/:user_id/access", requireProject("member.access.manage"), projectFlowH.ReplaceMemberAccess)
	projectFlow.GET("/my-permissions", requireProject("project.view"), projectFlowH.MyPermissions)

	invites := v2.Group("/invites")
	invites.Use(authMW)
	invites.GET("/incoming", projectFlowH.ListIncomingInvites)
	invites.GET("/outgoing", projectFlowH.ListOutgoingApplications)

	professor := v2.Group("/professor")
	professor.Use(authMW)
	professor.GET("/review-invites", requireFaculty("project.review.respond"), projectFlowH.ListProfessorReviewInvites)
}
