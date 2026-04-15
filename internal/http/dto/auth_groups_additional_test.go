package dto

import (
	"testing"
	"time"

	authsvc "idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPublicProfileResponseFromUser_HidesPrivateFields(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	groupID := uuid.New()
	groupNumber := 2201
	pendingAt := now

	resp := PublicProfileResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CPI",
		GroupID:          &groupID,
		GroupCode:        "CPI-2201",
		GroupNumber:      &groupNumber,
		Email:            "student@example.edu",
		PendingEmail:     " next@example.edu ",
		PendingEmailAt:   &pendingAt,
		FullName:         "Student Example",
		ProfileUpdatedAt: now,
	})

	require.Equal(t, groupID.String(), resp.GroupID)
	require.Equal(t, "CPI-2201", resp.GroupCode)
	require.Empty(t, resp.PendingEmail)
	require.Empty(t, resp.PendingStatus)
	require.Equal(t, now.Format(time.RFC3339), resp.UpdatedAt)
}

func TestPendingEmailStatus(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)

	require.Equal(t, "", pendingEmailStatus(authsvc.User{}))
	require.Equal(t, "pending_verification", pendingEmailStatus(authsvc.User{PendingEmail: " next@example.edu "}))
	require.Equal(t, "verification_sent", pendingEmailStatus(authsvc.User{
		PendingEmail:   " next@example.edu ",
		PendingEmailAt: &now,
	}))
}

func TestDepartmentAndGroupResponsesFromService(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	departmentID := uuid.New()
	facultyID := uuid.New()
	faculties := FacultyResponsesFromService([]authsvc.Faculty{{
		ID:        facultyID,
		Code:      "IDSAI_ENU",
		Name:      "IDSAI ENU",
		CreatedAt: now,
	}})

	departments := DepartmentResponsesFromService([]authsvc.Department{{
		ID:        departmentID,
		FacultyID: facultyID,
		Code:      "CPI",
		Name:      "Computer Science",
		ShortCode: "CS",
		CreatedAt: now,
	}})
	groups := StudentGroupResponsesFromService([]authsvc.StudentGroup{{
		ID:           uuid.New(),
		DepartmentID: departmentID,
		GroupCode:    "CPI-2201",
		GroupNumber:  2201,
		CreatedAt:    now,
		UpdatedAt:    now.Add(time.Hour),
	}})

	require.Len(t, faculties, 1)
	require.Equal(t, "IDSAI_ENU", faculties[0].Code)
	require.Len(t, departments, 1)
	require.Equal(t, "CPI", departments[0].Code)
	require.Equal(t, "Computer Science", departments[0].Name)
	require.Len(t, groups, 1)
	require.Equal(t, "CPI-2201", groups[0].GroupCode)
	require.Equal(t, 2201, groups[0].GroupNumber)
	require.Nil(t, FacultyResponsesFromService(nil))
	require.Nil(t, DepartmentResponsesFromService(nil))
	require.Nil(t, StudentGroupResponsesFromService(nil))
}

func TestGroupChangeAndDepartmentTreeResponsesFromService(t *testing.T) {
	now := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	reviewerID := uuid.New()
	groupStudentID := uuid.New()
	groupNodeID := uuid.New()
	departmentID := uuid.New()

	request := authsvc.GroupChangeRequest{
		ID:               uuid.New(),
		StudentID:        uuid.New(),
		StudentName:      "Student Example",
		StudentEmail:     "student@example.edu",
		CurrentGroupID:   uuid.New(),
		CurrentGroupCode: "CPI-2201",
		RequestedGroupID: uuid.New(),
		RequestedCode:    "CPI-2202",
		Status:           "APPROVED",
		AdminComment:     "ok",
		CreatedAt:        now,
		ReviewedAt:       &now,
		ReviewedBy:       &reviewerID,
		ReviewedByName:   "Admin Example",
	}

	requestResp := GroupChangeRequestResponseFromService(request)
	requests := GroupChangeRequestResponsesFromService([]authsvc.GroupChangeRequest{request})
	tree := DepartmentGroupsTreeResponsesFromService([]authsvc.DepartmentGroupsTree{{
		ID:        departmentID,
		Code:      "CPI",
		Name:      "Computer Science",
		ShortCode: "CS",
		Groups: []authsvc.GroupNode{{
			ID:            groupNodeID,
			GroupCode:     "CPI-2201",
			GroupNumber:   2201,
			TotalStudents: 1,
			Students: []authsvc.GroupStudent{{
				UserID:    groupStudentID,
				FullName:  "Student Example",
				Email:     "student@example.edu",
				AvatarURL: "/avatars/student.jpg",
				Status:    "ACTIVE",
				Role:      "STUDENT",
			}},
		}},
	}})

	require.Equal(t, request.RequestedCode, requestResp.RequestedCode)
	require.Len(t, requests, 1)
	require.Equal(t, reviewerID, *requests[0].ReviewedBy)
	require.Len(t, tree, 1)
	require.Len(t, tree[0].Groups, 1)
	require.Len(t, tree[0].Groups[0].Students, 1)
	require.Equal(t, groupStudentID, tree[0].Groups[0].Students[0].UserID)
	require.Nil(t, GroupChangeRequestResponsesFromService(nil))
	require.Nil(t, DepartmentGroupsTreeResponsesFromService(nil))
	require.Nil(t, GroupNodeResponsesFromService(nil))
	require.Nil(t, GroupStudentResponsesFromService(nil))
}
