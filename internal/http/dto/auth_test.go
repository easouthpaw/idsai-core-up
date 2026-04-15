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

func TestMeResponseFromUser_ExposesSchoolClass(t *testing.T) {
	groupID := uuid.New()
	groupNumber := 1065

	resp := MeResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		FacultyCode:      "CORE_SCHOOL",
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CLASS",
		GroupID:          &groupID,
		GroupCode:        "CLASS-10A",
		GroupNumber:      &groupNumber,
		Email:            "school@example.edu",
		FullName:         "School Example",
		ProfileUpdatedAt: time.Date(2026, time.April, 9, 10, 0, 0, 0, time.UTC),
	})

	require.Equal(t, authsvc.EducationTypeSchool, resp.EducationType)
	require.Equal(t, "10A", resp.SchoolClass)
	require.Equal(t, "CLASS-10A", resp.GroupCode)
}

func TestMeResponseFromUser_ExposesInstitution(t *testing.T) {
	resp := MeResponseFromUser(authsvc.User{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		FacultyID:        uuid.New(),
		DepartmentID:     uuid.New(),
		DepartmentCode:   "CPI",
		Email:            "student@example.edu",
		FullName:         "Student Example",
		ProfileUpdatedAt: time.Date(2026, time.April, 10, 10, 0, 0, 0, time.UTC),
		Institution: authsvc.InstitutionSelection{
			Provider:   authsvc.InstitutionProviderPhoton,
			ExternalID: "uni-1",
			Name:       "Nazarbayev University",
			Address:    "Астана, пр. Кабанбай батыра, 53",
		},
	})

	require.Equal(t, authsvc.InstitutionProviderPhoton, resp.InstitutionProvider)
	require.Equal(t, "uni-1", resp.InstitutionExternalID)
	require.Equal(t, "Nazarbayev University", resp.InstitutionName)
	require.Equal(t, "Астана, пр. Кабанбай батыра, 53", resp.InstitutionAddress)
}
