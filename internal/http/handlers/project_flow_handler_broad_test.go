package handlers

import (
	"net/http"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestProjectFlowHandlerProjectAndMemberRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	memberID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	positionID := uuid.New()
	now := time.Now().UTC()
	positionIDText := positionID.String()

	deps := &projectFlowTestDeps{
		can:        true,
		stackCodes: []string{"GO", "POSTGRES"},
		project: domain.Project{
			ID:                    projectID,
			Title:                 "Flow Project",
			Description:           "project",
			Status:                domain.ProjectRecruitment,
			CreatedBy:             userID,
			FacultyID:             facultyID,
			ProfessorReviewStatus: "ACCEPTED",
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		position: projectflow.Position{ID: positionID.String(), ProjectID: projectID.String(), Code: "BACKEND", Name: "Backend", Capacity: 2, CreatedAt: now},
		member: projectflow.Member{
			ID:            uuid.NewString(),
			ProjectID:     projectID.String(),
			UserID:        memberID.String(),
			PositionID:    &positionIDText,
			Status:        "ACTIVE",
			CreatedAt:     now,
			InviteComment: "join",
		},
		studentCandidates: []projectflow.StudentCandidate{{UserID: memberID.String(), FullName: "Student", Email: "student@example.edu", DepartmentCode: "CS"}},
		outgoingApps:      []projectflow.OutgoingApplication{{ProjectID: projectID.String(), ProjectTitle: "Flow Project", ProjectStatus: "RECRUITMENT", Status: "APPLIED", CreatedAt: now}},
	}
	handler := newProjectFlowHandlerForTest(deps)

	router := gin.New()
	router.Use(withFlowContext(userID, uuid.New(), facultyID))
	router.PUT("/projects/:project_id", handler.UpdateProject)
	router.PUT("/projects/:project_id/stacks", handler.SetStacks)
	router.GET("/projects/:project_id/stacks", handler.ListStacks)
	router.POST("/projects/:project_id/recruitment/open", handler.OpenRecruitment)
	router.POST("/projects/:project_id/positions", handler.CreatePosition)
	router.GET("/projects/:project_id/positions", handler.ListPositions)
	router.GET("/projects/:project_id/candidates/students", handler.ListStudentCandidates)
	router.POST("/projects/:project_id/members/invite", handler.InviteMember)
	router.POST("/projects/:project_id/members/apply", handler.ApplyMember)
	router.POST("/projects/:project_id/members/respond", handler.RespondMemberInvite)
	router.GET("/projects/:project_id/members", handler.ListMembers)
	router.POST("/projects/:project_id/members/:user_id/approve", handler.ApproveMember)
	router.POST("/projects/:project_id/members/:user_id/reject", handler.RejectMember)
	router.PUT("/projects/:project_id/members/:user_id/position", handler.SetMemberPosition)
	router.GET("/applications/outgoing", handler.ListOutgoingApplications)

	base := "/projects/" + projectID.String()
	requireStatus(t, router, http.MethodPut, base, `{"title":"Updated","description":"Updated desc"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPut, base+"/stacks", `{"stacks":["go","postgres"]}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/stacks", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/recruitment/open", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/positions", `{"code":"backend","name":"Backend","capacity":2}`, http.StatusCreated)
	requireStatus(t, router, http.MethodGet, base+"/positions", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/candidates/students?q=stud&limit=500", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/members/invite", `{"user_id":"`+memberID.String()+`","comment":"join"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/members/apply", `{"comment":"please"}`, http.StatusCreated)
	requireStatus(t, router, http.MethodPost, base+"/members/respond", `{"accept":true}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/members", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/members/"+memberID.String()+"/approve", `{"position_id":"`+positionID.String()+`"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/members/"+memberID.String()+"/reject", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodPut, base+"/members/"+memberID.String()+"/position", `{"position_id":"`+positionID.String()+`"}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/applications/outgoing?limit=999", "", http.StatusOK)
}

func TestProjectFlowHandlerTaskRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	positionID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()
	dueAt := now.Add(time.Hour).Format(time.RFC3339)
	userIDText := userID.String()

	deps := &projectFlowTestDeps{
		can: true,
		project: domain.Project{
			ID:                    projectID,
			Title:                 "Active Project",
			Status:                domain.ProjectActive,
			CreatedBy:             userID,
			FacultyID:             facultyID,
			ProfessorReviewStatus: "ACCEPTED",
		},
		position: projectflow.Position{ID: positionID.String(), ProjectID: projectID.String(), Code: "BACKEND", Name: "Backend", Capacity: 2},
		taskID:   taskID,
		task: projectflow.Task{
			ID:             taskID.String(),
			ProjectID:      projectID.String(),
			Title:          "Build API",
			Description:    "desc",
			PositionID:     positionID.String(),
			PositionCode:   "BACKEND",
			PositionName:   "Backend",
			AssigneeUserID: &userIDText,
			Status:         "IN_PROGRESS",
			CreatedBy:      userID.String(),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		activities: []projectflow.TaskActivity{{ID: uuid.NewString(), ProjectID: projectID.String(), TaskID: taskID.String(), EventType: "CREATED", Title: "Build API", CreatedAt: now}},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withFlowContext(userID, uuid.New(), facultyID))
	router.POST("/projects/:project_id/tasks", handler.CreateTask)
	router.GET("/projects/:project_id/tasks", handler.ListTasks)
	router.GET("/projects/:project_id/tasks/activities", handler.ListTaskActivities)
	router.PUT("/projects/:project_id/tasks/:task_id/status", handler.UpdateTaskStatus)
	router.PUT("/projects/:project_id/tasks/:task_id/assign", handler.AssignTask)
	router.POST("/projects/:project_id/tasks/:task_id/claim", handler.ClaimTask)
	router.POST("/projects/:project_id/tasks/:task_id/complete", handler.CompleteTask)
	router.DELETE("/projects/:project_id/tasks/:task_id", handler.DeleteTask)

	base := "/projects/" + projectID.String()
	taskPath := base + "/tasks/" + taskID.String()
	requireStatus(t, router, http.MethodPost, base+"/tasks", `{"title":"Build API","description":"desc","position_id":"`+positionID.String()+`","assignee_user_id":"`+userID.String()+`","due_at":"`+dueAt+`"}`, http.StatusCreated)
	requireStatus(t, router, http.MethodGet, base+"/tasks", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/tasks/activities?task_id="+taskID.String(), "", http.StatusOK)
	requireStatus(t, router, http.MethodPut, taskPath+"/status", `{"status":"done"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPut, taskPath+"/assign", `{"assignee_user_id":"`+userID.String()+`"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, taskPath+"/claim", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, taskPath+"/complete", `{"comment":"done","attachments":["a.txt"]}`, http.StatusOK)
	requireStatus(t, router, http.MethodDelete, taskPath, "", http.StatusNoContent)
}

func TestProjectFlowHandlerGradingAndProfessorRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	professorID := uuid.New()
	ownerID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	criterionID := uuid.New()
	positionID := uuid.New()
	now := time.Now().UTC()
	positionIDText := positionID.String()
	met := true

	deps := &projectFlowTestDeps{
		can: true,
		project: domain.Project{
			ID:                    projectID,
			Title:                 "Grading Project",
			Status:                domain.ProjectGrading,
			CreatedBy:             ownerID,
			ProfessorID:           &professorID,
			ProfessorReviewStatus: "ACCEPTED",
			FacultyID:             facultyID,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		assignedProfessor: projectflow.ProfessorCandidate{UserID: professorID.String(), FullName: "Professor", Email: "prof@example.edu", DepartmentCode: "CS"},
		criterion:         projectflow.Criterion{ID: criterionID.String(), ProjectID: projectID.String(), Title: "Quality", Description: "desc", Weight: 100, CreatedBy: ownerID.String(), CreatedAt: now},
		grades:            []projectflow.CriterionGrade{{CriterionID: criterionID.String(), IsMet: &met, Comment: "ok", UpdatedAt: &now}},
		member:            projectflow.Member{ID: uuid.NewString(), ProjectID: projectID.String(), UserID: uuid.NewString(), PositionID: &positionIDText, Status: "ACTIVE", CreatedAt: now},
		members:           []projectflow.Member{{ID: uuid.NewString(), ProjectID: projectID.String(), UserID: uuid.NewString(), PositionID: &positionIDText, Status: "ACTIVE", CreatedAt: now}},
		professors:        []projectflow.ProfessorCandidate{{UserID: professorID.String(), FullName: "Professor", Email: "prof@example.edu", DepartmentCode: "CS"}},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withFlowContext(professorID, uuid.New(), facultyID))
	router.POST("/projects/:project_id/criteria", handler.CreateCriterion)
	router.GET("/projects/:project_id/criteria", handler.ListCriteria)
	router.GET("/projects/:project_id/grading", handler.GetGrading)
	router.PUT("/projects/:project_id/grading", handler.UpsertGrading)
	router.GET("/projects/:project_id/readiness", handler.Readiness)
	router.POST("/projects/:project_id/grading/publish", handler.PublishProjectGrading)
	router.POST("/projects/:project_id/grading/return", handler.ReturnProjectForRetake)
	router.GET("/projects/:project_id/professors/search", handler.SearchProfessors)
	router.GET("/projects/:project_id/professor", handler.GetAssignedProfessor)
	router.POST("/projects/:project_id/professor", handler.AssignProfessor)
	router.POST("/projects/:project_id/professor/respond", handler.RespondProfessorInvite)
	router.GET("/professor/reviews", handler.ListProfessorReviewInvites)

	base := "/projects/" + projectID.String()
	requireStatus(t, router, http.MethodPost, base+"/criteria", `{"title":"Quality","description":"desc","weight":100}`, http.StatusCreated)
	requireStatus(t, router, http.MethodGet, base+"/criteria", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/grading", "", http.StatusOK)
	requireStatus(t, router, http.MethodPut, base+"/grading", `{"items":[{"criterion_id":"`+criterionID.String()+`","is_met":true,"comment":"ok"}]}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/readiness", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/grading/publish", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/grading/return", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/professors/search?q=prof&limit=999", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, base+"/professor", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/professor", `{"professor_id":"`+professorID.String()+`"}`, http.StatusBadRequest)
	requireStatus(t, router, http.MethodPost, base+"/professor/respond", `{"accept":true}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/professor/reviews?q=grade&limit=999", "", http.StatusOK)
}

func TestProjectFlowHandlerApproveAndSubmitRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	memberID := uuid.New()
	professorID := uuid.New()
	projectID := uuid.New()
	facultyID := uuid.New()
	positionID := uuid.New()
	positionIDText := positionID.String()

	deps := &projectFlowTestDeps{
		can: true,
		project: domain.Project{
			ID:                    projectID,
			Title:                 "Ready Project",
			Status:                domain.ProjectActive,
			CreatedBy:             userID,
			ProfessorID:           &professorID,
			ProfessorReviewStatus: "ACCEPTED",
			FacultyID:             facultyID,
		},
		criterion: projectflow.Criterion{ID: uuid.NewString(), ProjectID: projectID.String(), Title: "Quality", Weight: 100, CreatedBy: userID.String()},
		members:   []projectflow.Member{{ID: uuid.NewString(), ProjectID: projectID.String(), UserID: memberID.String(), PositionID: &positionIDText, Status: "ACTIVE"}},
	}
	handler := newProjectFlowHandlerForTest(deps)
	router := gin.New()
	router.Use(withFlowContext(userID, uuid.New(), facultyID))
	router.POST("/projects/:project_id/approve", handler.ApproveProject)
	router.POST("/projects/:project_id/grading/submit", handler.SubmitProjectForGrading)

	base := "/projects/" + projectID.String()
	requireStatus(t, router, http.MethodPost, base+"/approve", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, base+"/grading/submit", `{}`, http.StatusOK)
}

func withFlowContext(userID, tenantID, facultyID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("tenantID", tenantID)
		c.Set("facultyID", facultyID)
		c.Next()
	}
}
