package dto

import (
	"testing"
	"time"

	"idsai-core-up/internal/services/projectflow"

	"github.com/stretchr/testify/require"
)

func TestProjectFlowDTOResponsesFromService(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	isMet := true
	positionID := "position-1"
	positionCode := "BE"
	positionName := "Backend"
	userID := "user-1"
	invitedBy := "lead-1"

	grades := CriterionGradesFromRequest([]GradingItemRequest{{
		CriterionID: " criterion-1 ",
		IsMet:       &isMet,
		Comment:     " done ",
	}})
	stackResp := ProjectStackResponsesFromService([]projectflow.Stack{{Code: "go"}})
	criteria := ProjectFlowCriterionResponsesFromService([]projectflow.Criterion{{
		ID:          "criterion-1",
		ProjectID:   "project-1",
		Title:       "Architecture",
		Description: "Must be clean",
		Weight:      30,
		CreatedBy:   userID,
		CreatedAt:   now,
	}})
	criterionGrades := ProjectFlowCriterionGradeResponsesFromService([]projectflow.CriterionGrade{{
		CriterionID: "criterion-1",
		IsMet:       &isMet,
		Comment:     "ok",
		UpdatedAt:   &now,
	}})
	positions := ProjectFlowPositionResponsesFromService([]projectflow.Position{{
		ID:        positionID,
		ProjectID: "project-1",
		Code:      positionCode,
		Name:      positionName,
		Capacity:  2,
		CreatedAt: now,
	}})
	members := ProjectFlowMemberResponsesFromService([]projectflow.Member{{
		ID:            "member-1",
		ProjectID:     "project-1",
		UserID:        userID,
		FullName:      "Student Example",
		Email:         "student@example.edu",
		PositionID:    &positionID,
		PositionCode:  &positionCode,
		PositionName:  &positionName,
		Status:        "ACTIVE",
		InviteComment: "join us",
		InvitedBy:     &invitedBy,
		RespondedAt:   &now,
		JoinedAt:      &now,
		CreatedAt:     now,
	}})
	tasks := ProjectFlowTaskResponsesFromService([]projectflow.Task{{
		ID:             "task-1",
		ProjectID:      "project-1",
		Title:          "Build API",
		Description:    "Create endpoints",
		PositionID:     positionID,
		PositionCode:   positionCode,
		PositionName:   positionName,
		AssigneeUserID: &userID,
		Status:         "DONE",
		CreatedBy:      userID,
		DueAt:          &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}})
	activities := ProjectFlowTaskActivityResponsesFromService([]projectflow.TaskActivity{{
		ID:          "activity-1",
		ProjectID:   "project-1",
		TaskID:      "task-1",
		ActorUserID: &userID,
		ActorName:   "Student Example",
		ActorEmail:  "student@example.edu",
		EventType:   "completed",
		FromStatus:  "IN_PROGRESS",
		ToStatus:    "DONE",
		Title:       "Task completed",
		Comment:     "done",
		Attachments: []string{"artifact.pdf"},
		CreatedAt:   now,
	}})
	students := ProjectFlowStudentCandidateResponsesFromService([]projectflow.StudentCandidate{{
		UserID:         userID,
		FullName:       "Student Example",
		Email:          "student@example.edu",
		DepartmentCode: "CPI",
	}})
	professor := projectflow.ProfessorCandidate{
		UserID:         "prof-1",
		FullName:       "Professor Example",
		Email:          "prof@example.edu",
		DepartmentCode: "CPI",
	}
	professors := ProjectFlowProfessorCandidateResponsesFromService([]projectflow.ProfessorCandidate{professor})
	assignedProfessor := AssignedProfessorResponseFromService(&professor)
	invites := ProjectFlowIncomingInviteResponsesFromService([]projectflow.IncomingInvite{{
		ProjectID:     "project-1",
		ProjectTitle:  "AI Platform",
		ProjectStatus: "RECRUITMENT",
		Status:        "PENDING",
		UserID:        userID,
		InviteComment: "join us",
		InvitedBy:     &invitedBy,
		InviterName:   "Lead Example",
		InviterEmail:  "lead@example.edu",
		CreatedAt:     now,
		RespondedAt:   &now,
	}})
	applications := ProjectFlowOutgoingApplicationResponsesFromService([]projectflow.OutgoingApplication{{
		ProjectID:     "project-1",
		ProjectTitle:  "AI Platform",
		ProjectStatus: "RECRUITMENT",
		Status:        "PENDING",
		CreatedAt:     now,
		RespondedAt:   &now,
	}})
	accessCatalog := ProjectFlowAccessCatalogItemResponsesFromService([]projectflow.AccessCatalogItem{{
		Code:            "TEAM_LEAD",
		Name:            "Team Lead",
		Description:     "Manages members",
		PermissionCodes: []string{"project.manage", "member.manage"},
	}})
	memberAccess := ProjectFlowMemberAccessResponseFromService(&projectflow.MemberAccess{
		UserID:                   userID,
		RoleCodes:                []string{"TEAM_MEMBER"},
		ManagedRoleCodes:         []string{"TEAM_MEMBER"},
		EffectivePermissionCodes: []string{"task.view"},
	})
	readiness := ProjectFlowReadinessResponseFromService(projectflow.Readiness{
		ProjectID:       "project-1",
		Status:          "RECRUITMENT",
		RequiredMembers: 3,
		ActiveMembers:   2,
		HasProfessor:    true,
		ProfessorStatus: "ASSIGNED",
		CriteriaCount:   5,
		CanActivate:     false,
	})

	require.Len(t, grades, 1)
	require.Equal(t, "criterion-1", grades[0].CriterionID)
	require.Equal(t, "done", grades[0].Comment)
	require.Equal(t, "go", stackResp[0].Code)
	require.Len(t, criteria, 1)
	require.Len(t, criterionGrades, 1)
	require.Len(t, positions, 1)
	require.Len(t, members, 1)
	require.Len(t, tasks, 1)
	require.Len(t, activities, 1)
	require.Len(t, students, 1)
	require.Len(t, professors, 1)
	require.NotNil(t, assignedProfessor.Professor)
	require.Len(t, invites, 1)
	require.Len(t, applications, 1)
	require.Len(t, accessCatalog, 1)
	require.NotNil(t, memberAccess)
	require.Equal(t, "project-1", readiness.ProjectID)

	require.Nil(t, CriterionGradesFromRequest(nil))
	require.Nil(t, ProjectStackResponsesFromService(nil))
	require.Nil(t, ProjectFlowCriterionResponsesFromService(nil))
	require.Nil(t, ProjectFlowCriterionGradeResponsesFromService(nil))
	require.Nil(t, ProjectFlowPositionResponsesFromService(nil))
	require.Nil(t, ProjectFlowMemberResponsesFromService(nil))
	require.Nil(t, ProjectFlowTaskResponsesFromService(nil))
	require.Nil(t, ProjectFlowTaskActivityResponsesFromService(nil))
	require.Nil(t, ProjectFlowStudentCandidateResponsesFromService(nil))
	require.Nil(t, ProjectFlowProfessorCandidateResponsesFromService(nil))
	require.Equal(t, AssignedProfessorResponse{}, AssignedProfessorResponseFromService(nil))
	require.Nil(t, ProjectFlowIncomingInviteResponsesFromService(nil))
	require.Nil(t, ProjectFlowOutgoingApplicationResponsesFromService(nil))
	require.Nil(t, ProjectFlowAccessCatalogItemResponsesFromService(nil))
	require.Nil(t, ProjectFlowMemberAccessResponseFromService(nil))
}
