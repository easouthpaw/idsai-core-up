package dto

import (
	"testing"
	"time"

	authsvc "idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMeResponseFromUser_StripsGroupForProfessor(t *testing.T) {
	groupID := uuid.New()
	groupNumber := 2201

	resp := MeResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CPI",
		GroupID:          &groupID,
		GroupCode:        "CPI-2201",
		GroupNumber:      &groupNumber,
		Email:            "prof@example.edu",
		FullName:         "Professor Example",
		ProfileUpdatedAt: time.Date(2026, time.April, 6, 10, 0, 0, 0, time.UTC),
		IsProfessor:      true,
	})

	require.Equal(t, "", resp.GroupID)
	require.Equal(t, "", resp.GroupCode)
	require.Nil(t, resp.GroupNumber)
	require.Equal(t, "CPI", resp.DepartmentCode)
}

func TestMeResponseFromUser_KeepsGroupForStudent(t *testing.T) {
	groupID := uuid.New()
	groupNumber := 2201

	resp := MeResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CPI",
		GroupID:          &groupID,
		GroupCode:        "CPI-2201",
		GroupNumber:      &groupNumber,
		Email:            "student@example.edu",
		FullName:         "Student Example",
		ProfileUpdatedAt: time.Date(2026, time.April, 6, 10, 0, 0, 0, time.UTC),
	})

	require.Equal(t, groupID.String(), resp.GroupID)
	require.Equal(t, "CPI-2201", resp.GroupCode)
	require.NotNil(t, resp.GroupNumber)
	require.Equal(t, 2201, *resp.GroupNumber)
}

func TestMeResponseFromUser_StripsGroupForAdmin(t *testing.T) {
	groupID := uuid.New()
	groupNumber := 2201

	resp := MeResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CPI",
		GroupID:          &groupID,
		GroupCode:        "CPI-2201",
		GroupNumber:      &groupNumber,
		Email:            "admin@example.edu",
		FullName:         "Admin Example",
		ProfileUpdatedAt: time.Date(2026, time.April, 8, 10, 0, 0, 0, time.UTC),
		IsAdmin:          true,
	})

	require.Equal(t, "", resp.GroupID)
	require.Equal(t, "", resp.GroupCode)
	require.Nil(t, resp.GroupNumber)
	require.Equal(t, "CPI", resp.DepartmentCode)
}
