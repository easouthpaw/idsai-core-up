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

// TestProjectLifecycleDraftToCompleted walks every stage shown in a demo:
//
//	DRAFT → RECRUITMENT → ACTIVE → GRADING → COMPLETED
//
// Each sub-test advances the shared deps state so the next phase sees the
// correct project status, mirroring what the real service would store.
func TestProjectLifecycleDraftToCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ownerID     := uuid.New()
	tenantID    := uuid.New()
	professorID := uuid.New() // must differ from ownerID and project.CreatedBy
	memberID    := uuid.New()
	projectID   := uuid.New()
	facultyID   := uuid.New()
	positionID  := uuid.New()
	criterionID := uuid.New()
	taskID      := uuid.New()
	now         := time.Now().UTC()
	dueAt       := now.Add(24 * time.Hour).Format(time.RFC3339)
	met         := true
	posIDText   := positionID.String()

	deps := &projectFlowTestDeps{
		can: true,
		project: domain.Project{
			ID:                    projectID,
			Title:                 "Diploma: IDS AI Core",
			Description:           "ML platform for student projects",
			Status:                domain.ProjectDraft,
			CreatedBy:             ownerID,
			FacultyID:             facultyID,
			ProfessorReviewStatus: "NONE",
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		stackCodes: []string{"GO", "PYTHON", "POSTGRES"},
		position: projectflow.Position{
			ID:        posIDText,
			ProjectID: projectID.String(),
			Code:      "BACKEND",
			Name:      "Backend Developer",
			Capacity:  2,
			CreatedAt: now,
		},
		member: projectflow.Member{
			ID:         uuid.NewString(),
			ProjectID:  projectID.String(),
			UserID:     memberID.String(),
			PositionID: &posIDText,
			Status:     "ACTIVE",
			CreatedAt:  now,
		},
		studentCandidates: []projectflow.StudentCandidate{
			{UserID: memberID.String(), FullName: "Aibolat", Email: "aibolat@uni.edu", DepartmentCode: "CS"},
		},
		assignedProfessor: projectflow.ProfessorCandidate{
			UserID: professorID.String(), FullName: "Prof Smith", Email: "smith@uni.edu", DepartmentCode: "CS",
		},
		professors: []projectflow.ProfessorCandidate{
			{UserID: professorID.String(), FullName: "Prof Smith", Email: "smith@uni.edu", DepartmentCode: "CS"},
		},
		criterion: projectflow.Criterion{
			ID:          criterionID.String(),
			ProjectID:   projectID.String(),
			Title:       "Code Quality",
			Description: "Clean, tested, well-structured code",
			Weight:      100,
			CreatedBy:   ownerID.String(),
			CreatedAt:   now,
		},
		grades: []projectflow.CriterionGrade{
			{CriterionID: criterionID.String(), IsMet: &met, Comment: "excellent", UpdatedAt: &now},
		},
		taskID: taskID,
		task: projectflow.Task{
			ID:           taskID.String(),
			ProjectID:    projectID.String(),
			Title:        "Build core API",
			Description:  "REST endpoints for project management",
			PositionID:   posIDText,
			PositionCode: "BACKEND",
			PositionName: "Backend Developer",
			Status:       "IN_PROGRESS",
			CreatedBy:    ownerID.String(),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		activities: []projectflow.TaskActivity{
			{ID: uuid.NewString(), ProjectID: projectID.String(), TaskID: taskID.String(), EventType: "CREATED", Title: "Build core API", CreatedAt: now},
		},
		outgoingApps: []projectflow.OutgoingApplication{
			{ProjectID: projectID.String(), ProjectTitle: "Diploma: IDS AI Core", ProjectStatus: "RECRUITMENT", Status: "APPLIED", CreatedAt: now},
		},
	}

	handler := newProjectFlowHandlerForTest(deps)
	base     := "/projects/" + projectID.String()
	taskPath := base + "/tasks/" + taskID.String()

	router := gin.New()
	router.Use(withFlowContext(ownerID, tenantID, facultyID))

	// ── project editing ──────────────────────────────────────────────────────────
	router.PUT("/projects/:project_id", handler.UpdateProject)
	router.PUT("/projects/:project_id/stacks", handler.SetStacks)
	router.GET("/projects/:project_id/stacks", handler.ListStacks)
	// ── recruitment ──────────────────────────────────────────────────────────────
	router.POST("/projects/:project_id/recruitment/open", handler.OpenRecruitment)
	router.POST("/projects/:project_id/positions", handler.CreatePosition)
	router.GET("/projects/:project_id/positions", handler.ListPositions)
	router.GET("/projects/:project_id/candidates/students", handler.ListStudentCandidates)
	router.POST("/projects/:project_id/members/invite", handler.InviteMember)
	router.POST("/projects/:project_id/members/:user_id/approve", handler.ApproveMember)
	router.GET("/projects/:project_id/members", handler.ListMembers)
	// ── professor ────────────────────────────────────────────────────────────────
	router.GET("/projects/:project_id/professors/search", handler.SearchProfessors)
	router.POST("/projects/:project_id/professor", handler.AssignProfessor)
	router.GET("/projects/:project_id/professor", handler.GetAssignedProfessor)
	// ── active + tasks ───────────────────────────────────────────────────────────
	router.POST("/projects/:project_id/approve", handler.ApproveProject)
	router.POST("/projects/:project_id/tasks", handler.CreateTask)
	router.GET("/projects/:project_id/tasks", handler.ListTasks)
	router.GET("/projects/:project_id/tasks/activities", handler.ListTaskActivities)
	router.PUT("/projects/:project_id/tasks/:task_id/status", handler.UpdateTaskStatus)
	router.POST("/projects/:project_id/tasks/:task_id/claim", handler.ClaimTask)
	router.POST("/projects/:project_id/tasks/:task_id/complete", handler.CompleteTask)
	// ── submit for grading (owner action) ───────────────────────────────────────
	router.POST("/projects/:project_id/grading/submit", handler.SubmitProjectForGrading)
	router.POST("/projects/:project_id/criteria", handler.CreateCriterion)
	router.GET("/projects/:project_id/criteria", handler.ListCriteria)
	router.GET("/projects/:project_id/readiness", handler.Readiness)

	// Grading evaluation routes require the caller to BE the professor
	// (grading.view / grading.mark_criteria / grading.publish permissions +
	//  ProfessorID == userID check in PublishGrading / ReturnProjectForRetake).
	profRouter := gin.New()
	profRouter.Use(withFlowContext(professorID, tenantID, facultyID))
	profRouter.GET("/projects/:project_id/grading", handler.GetGrading)
	profRouter.PUT("/projects/:project_id/grading", handler.UpsertGrading)
	profRouter.POST("/projects/:project_id/grading/publish", handler.PublishProjectGrading)

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 1: DRAFT — edit title, description, tech stack
	// ────────────────────────────────────────────────────────────────────────────
	t.Run("phase_1_draft", func(t *testing.T) {
		requireStatus(t, router, http.MethodPut, base,
			`{"title":"Diploma: IDS AI Core","description":"ML platform for student projects"}`,
			http.StatusOK)
		requireStatus(t, router, http.MethodPut, base+"/stacks",
			`{"stacks":["go","python","postgres"]}`,
			http.StatusOK)
		requireStatus(t, router, http.MethodGet, base+"/stacks", "", http.StatusOK)
	})

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 2: RECRUITMENT — open, create position, invite and approve member
	// ────────────────────────────────────────────────────────────────────────────
	deps.project.Status = domain.ProjectRecruitment

	t.Run("phase_2_recruitment", func(t *testing.T) {
		requireStatus(t, router, http.MethodPost, base+"/recruitment/open", `{}`, http.StatusOK)

		requireStatus(t, router, http.MethodPost, base+"/positions",
			`{"code":"backend","name":"Backend Developer","capacity":2}`,
			http.StatusCreated)
		requireStatus(t, router, http.MethodGet, base+"/positions", "", http.StatusOK)

		requireStatus(t, router, http.MethodGet,
			base+"/candidates/students?q=aibolat&limit=20",
			"", http.StatusOK)

		requireStatus(t, router, http.MethodPost, base+"/members/invite",
			`{"user_id":"`+memberID.String()+`","comment":"join the team"}`,
			http.StatusOK)
		requireStatus(t, router, http.MethodPost,
			base+"/members/"+memberID.String()+"/approve",
			`{"position_id":"`+positionID.String()+`"}`,
			http.StatusOK)
		requireStatus(t, router, http.MethodGet, base+"/members", "", http.StatusOK)
	})

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 3: Professor assignment (still RECRUITMENT)
	// ────────────────────────────────────────────────────────────────────────────
	t.Run("phase_3_professor_assignment", func(t *testing.T) {
		requireStatus(t, router, http.MethodGet,
			base+"/professors/search?q=smith&limit=10",
			"", http.StatusOK)

		// assign professor: ownerID ≠ professorID, so validation passes
		requireStatus(t, router, http.MethodPost, base+"/professor",
			`{"professor_id":"`+professorID.String()+`"}`,
			http.StatusOK)
	})

	deps.project.ProfessorID = &professorID
	deps.project.ProfessorReviewStatus = "ACCEPTED"

	t.Run("phase_3_professor_accepted", func(t *testing.T) {
		requireStatus(t, router, http.MethodGet, base+"/professor", "", http.StatusOK)
	})

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 4: ACTIVE — approve project, task lifecycle
	// ────────────────────────────────────────────────────────────────────────────
	deps.project.Status = domain.ProjectActive

	t.Run("phase_4_active_tasks", func(t *testing.T) {
		requireStatus(t, router, http.MethodPost, base+"/approve", `{}`, http.StatusOK)

		requireStatus(t, router, http.MethodPost, base+"/tasks",
			`{"title":"Build core API","description":"REST endpoints","position_id":"`+positionID.String()+`","assignee_user_id":"`+ownerID.String()+`","due_at":"`+dueAt+`"}`,
			http.StatusCreated)
		requireStatus(t, router, http.MethodGet, base+"/tasks", "", http.StatusOK)
		requireStatus(t, router, http.MethodGet,
			base+"/tasks/activities?task_id="+taskID.String(),
			"", http.StatusOK)

		requireStatus(t, router, http.MethodPost, taskPath+"/claim", `{}`, http.StatusOK)
		requireStatus(t, router, http.MethodPut, taskPath+"/status", `{"status":"done"}`, http.StatusOK)
		requireStatus(t, router, http.MethodPost, taskPath+"/complete",
			`{"comment":"implementation complete","attachments":[]}`,
			http.StatusOK)
	})

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 5: GRADING — submit, set criteria, grade, publish
	// ────────────────────────────────────────────────────────────────────────────
	t.Run("phase_5_submit_for_grading", func(t *testing.T) {
		requireStatus(t, router, http.MethodPost, base+"/grading/submit", `{}`, http.StatusOK)
	})

	deps.project.Status = domain.ProjectGrading

	t.Run("phase_5_grading_criteria_and_evaluation", func(t *testing.T) {
		// Owner sets criteria
		requireStatus(t, router, http.MethodPost, base+"/criteria",
			`{"title":"Code Quality","description":"Clean, tested, well-structured code","weight":100}`,
			http.StatusCreated)
		requireStatus(t, router, http.MethodGet, base+"/criteria", "", http.StatusOK)
		requireStatus(t, router, http.MethodGet, base+"/readiness", "", http.StatusOK)

		// Professor views, fills in grades, and publishes
		requireStatus(t, profRouter, http.MethodGet, base+"/grading", "", http.StatusOK)
		requireStatus(t, profRouter, http.MethodPut, base+"/grading",
			`{"items":[{"criterion_id":"`+criterionID.String()+`","is_met":true,"comment":"excellent work"}]}`,
			http.StatusOK)
		requireStatus(t, profRouter, http.MethodPost, base+"/grading/publish", `{}`, http.StatusOK)
	})

	// ────────────────────────────────────────────────────────────────────────────
	// Phase 6: COMPLETED — project remains readable, grading is visible
	// ────────────────────────────────────────────────────────────────────────────
	deps.project.Status = domain.ProjectCompleted

	t.Run("phase_6_completed", func(t *testing.T) {
		requireStatus(t, profRouter, http.MethodGet, base+"/grading", "", http.StatusOK)
		requireStatus(t, router, http.MethodGet, base+"/members", "", http.StatusOK)
		requireStatus(t, router, http.MethodGet, base+"/criteria", "", http.StatusOK)
	})
}
