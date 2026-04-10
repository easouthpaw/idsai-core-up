package dto

import (
	"encoding/json"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/admin"
	"idsai-core-up/internal/services/notifications"
	"idsai-core-up/internal/services/projects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAdminDTOResponsesFromService(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	projectID := uuid.New()
	assigneeID := uuid.New()
	isMet := true

	users := AdminUserResponsesFromService([]admin.User{{
		ID:             userID,
		FullName:       "Admin Example",
		Email:          "admin@example.edu",
		RoleCode:       admin.RoleProfessor,
		Status:         admin.StatusActive,
		FacultyCode:    "FIT",
		DepartmentCode: "CPI",
	}})
	projectsResp := AdminProjectResponsesFromService([]admin.Project{{
		ID:             projectID,
		Title:          "AI Platform",
		Description:    "demo",
		Status:         "ACTIVE",
		Visibility:     "FACULTY",
		IsPublic:       false,
		CreatedBy:      userID,
		AuthorName:     "Admin Example",
		AuthorEmail:    "admin@example.edu",
		FacultyCode:    "FIT",
		DepartmentCode: "CPI",
		CreatedAt:      now,
		UpdatedAt:      now.Add(time.Hour),
	}})
	observation := AdminProjectObservationResponseFromService(admin.ProjectObservation{
		Project: admin.Project{
			ID:        projectID,
			Title:     "AI Platform",
			CreatedBy: userID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Positions: []admin.ProjectPosition{{ID: uuid.New(), Code: "BE", Name: "Backend", Capacity: 2}},
		Members: []admin.ProjectMember{{
			UserID:       userID,
			FullName:     "Student Example",
			Email:        "student@example.edu",
			RoleCode:     "TEAM_LEAD",
			Status:       "ACTIVE",
			PositionCode: "BE",
			PositionName: "Backend",
			JoinedAt:     &now,
		}},
		Tasks: []admin.ProjectTask{{
			ID:             uuid.New(),
			Title:          "Build API",
			Status:         "DONE",
			PositionCode:   "BE",
			AssigneeUserID: &assigneeID,
			AssigneeName:   "Student Example",
			DueAt:          &now,
			UpdatedAt:      now,
		}},
		Criteria: []admin.ProjectCriterion{{
			ID:        uuid.New(),
			Title:     "Architecture",
			Weight:    30,
			CreatedBy: userID,
			CreatedAt: now,
			IsMet:     &isMet,
			Comment:   "ok",
			UpdatedAt: &now,
		}},
		Summary: admin.ProjectObservationSummary{
			MembersTotal:  1,
			MembersActive: 1,
			TasksTotal:    1,
			TasksDone:     1,
			CriteriaTotal: 1,
		},
	})

	require.Len(t, users, 1)
	require.Equal(t, userID, users[0].ID)
	require.Len(t, projectsResp, 1)
	require.Equal(t, projectID, projectsResp[0].ID)
	require.Len(t, observation.Positions, 1)
	require.Len(t, observation.Members, 1)
	require.Len(t, observation.Tasks, 1)
	require.Len(t, observation.Criteria, 1)
	require.Equal(t, 1, observation.Summary.TasksDone)
	require.Nil(t, AdminUserResponsesFromService(nil))
	require.Nil(t, AdminProjectResponsesFromService(nil))
	require.Nil(t, AdminProjectPositionResponsesFromService(nil))
	require.Nil(t, AdminProjectMemberResponsesFromService(nil))
	require.Nil(t, AdminProjectTaskResponsesFromService(nil))
	require.Nil(t, AdminProjectCriterionResponsesFromService(nil))
}

func TestKnowledgeBaseAndNotificationDTOResponses(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	parentID := uuid.New()
	category := domain.KBCategory{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		ParentID:  &parentID,
		Title:     "Guides",
		Slug:      "guides",
		SortOrder: 10,
		CreatedAt: now,
		UpdatedAt: now,
	}
	article := domain.KBArticle{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		CategoryID:   category.ID,
		AuthorID:     uuid.New(),
		Title:        "First Article",
		Slug:         "first-article",
		Content:      "# First Article",
		Status:       "PUBLISHED",
		CreatedAt:    now,
		UpdatedAt:    now,
		PublishedAt:  &now,
		AuthorName:   "Author Example",
		AuthorAvatar: "/media/avatar.jpg",
		CategoryPath: "Guides / Intro",
		Tags:         []string{"go", "rbac"},
	}
	listItem := domain.KBArticleListItem{
		ID:          article.ID,
		CategoryID:  article.CategoryID,
		AuthorID:    article.AuthorID,
		Title:       article.Title,
		Slug:        article.Slug,
		Status:      article.Status,
		CreatedAt:   article.CreatedAt,
		UpdatedAt:   article.UpdatedAt,
		PublishedAt: article.PublishedAt,
		AuthorName:  article.AuthorName,
		Tags:        article.Tags,
	}
	payload := json.RawMessage(`{"project_id":"demo"}`)
	notification := notifications.Notification{
		ID:        uuid.NewString(),
		Type:      "project.updated",
		Title:     "Updated",
		Body:      "Project updated",
		Payload:   payload,
		IsRead:    true,
		CreatedAt: now,
		ReadAt:    &now,
	}

	require.Equal(t, category.Title, KBCategoryResponseFromDomain(category).Title)
	require.Equal(t, article.AuthorAvatar, KBArticleResponseFromDomain(article).AuthorAvatar)
	require.Len(t, KBCategoryResponsesFromDomain([]domain.KBCategory{category}), 1)
	require.Len(t, KBArticleListItemResponsesFromDomain([]domain.KBArticleListItem{listItem}), 1)
	require.Len(t, KBTagResponsesFromDomain([]domain.KBTag{{ID: uuid.New(), Name: "rbac"}}), 1)
	require.Nil(t, KBCategoryResponsesFromDomain(nil))
	require.Nil(t, KBArticleListItemResponsesFromDomain(nil))
	require.Nil(t, KBTagResponsesFromDomain(nil))

	notificationResp := NotificationResponseFromService(notification)
	require.Equal(t, payload, notificationResp.Payload)
	require.Len(t, NotificationResponsesFromService([]notifications.Notification{notification}), 1)
	require.Nil(t, NotificationResponsesFromService(nil))
}

func TestProjectDTOResponsesFromDomainAndView(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	projectID := uuid.New()
	createdBy := uuid.New()
	facultyID := uuid.New()
	groupID := uuid.New()
	professorID := uuid.New()

	require.Equal(t, "CPI", toUIVisibility(" cpi "))
	require.Equal(t, "PRIVATE", toUIVisibility("group"))
	require.Equal(t, "PUBLIC", toUIVisibility("public"))
	department, number := splitGroupCode("cpi-2201")
	require.Equal(t, "CPI", department)
	require.Equal(t, "2201", number)

	project := domain.Project{
		ID:                  projectID,
		Title:               "AI Platform",
		Description:         "demo",
		Status:              domain.ProjectRecruitment,
		IsPublic:            false,
		CreatedBy:           createdBy,
		ProfessorID:         &professorID,
		CreatedAt:           now,
		UpdatedAt:           now.Add(time.Hour),
		FacultyID:           facultyID,
		Visibility:          "GROUP",
		GroupID:             &groupID,
		ImageKey:            "covers/project.jpg",
		DefaultCoverVariant: 0,
		RetakeCount:         2,
	}
	resp := ProjectResponseFromDomain(project)
	viewResp := ProjectResponseFromView(projects.ProjectView{
		Project: project,
		Access: projects.ViewerAccess{
			CanViewWorkspace:      true,
			CanViewProjectDetails: true,
			CanApply:              false,
			CanViewFinalGrade:     true,
		},
		ReviewSummary: &projects.ReviewSummary{
			Score:       "4.5",
			PassPercent: 90,
			Met:         9,
			Total:       10,
			ReviewedAt:  &now,
			Reviewer:    "Professor Example",
		},
	})
	groups := GroupOptionResponsesFromService([]projects.Group{{
		ID:   uuid.New(),
		Code: "CPI-2201",
		Name: "CPI-2201",
	}})

	require.Equal(t, "PRIVATE", resp.Visibility)
	require.Equal(t, groupID.String(), *resp.GroupID)
	require.Equal(t, professorID.String(), *resp.ProfessorID)
	require.Equal(t, "NONE", ProjectResponseFromDomain(domain.Project{
		ID:        projectID,
		CreatedBy: createdBy,
		FacultyID: facultyID,
	}).ProfessorReviewStatus)
	require.Equal(t, 1, resp.DefaultCoverVariant)
	require.True(t, resp.HasCustomImage)
	require.Equal(t, 10, resp.RetakePenaltyPercent)
	require.NotNil(t, viewResp.ViewerAccess)
	require.NotNil(t, viewResp.ReviewSummary)
	require.Equal(t, "CPI", groups[0].Department)
	require.Equal(t, "2201", groups[0].Number)
	require.Nil(t, GroupOptionResponsesFromService(nil))
}
