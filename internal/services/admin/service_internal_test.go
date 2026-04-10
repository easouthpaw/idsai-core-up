package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizationHelpers(t *testing.T) {
	role, err := normalizeRole(" student ")
	require.NoError(t, err)
	require.Equal(t, RoleStudent, role)
	require.ErrorIs(t, errInvalid(normalizeRole("guest")), ErrInvalidRole)

	roleFilter, err := normalizeRoleFilter("")
	require.NoError(t, err)
	require.Empty(t, roleFilter)
	roleFilter, err = normalizeRoleFilter(" super_admin ")
	require.NoError(t, err)
	require.Equal(t, "SUPER_ADMIN", roleFilter)
	require.ErrorIs(t, errInvalid(normalizeRoleFilter("guest")), ErrInvalidRole)

	status, err := normalizeStatus(" active ")
	require.NoError(t, err)
	require.Equal(t, StatusActive, status)
	require.ErrorIs(t, errInvalid(normalizeStatus("archived")), ErrInvalidStatus)

	projectStatus, err := normalizeProjectStatus(" grading ")
	require.NoError(t, err)
	require.Equal(t, "GRADING", projectStatus)
	require.ErrorIs(t, errInvalid(normalizeProjectStatus("paused")), ErrInvalidProjectStatus)

	filter, err := normalizeProjectStatusFilter("")
	require.NoError(t, err)
	require.Empty(t, filter)
	filter, err = normalizeProjectStatusFilter(" archive ")
	require.NoError(t, err)
	require.Equal(t, "ARCHIVE", filter)
}

func TestValidateAdminProjectStatusChange(t *testing.T) {
	require.NoError(t, validateAdminProjectStatusChange("review", "active"))
	require.NoError(t, validateAdminProjectStatusChange("completed", "archive"))
	require.ErrorIs(t, validateAdminProjectStatusChange("draft", "active"), ErrInvalidInput)
	require.ErrorIs(t, validateAdminProjectStatusChange("active", "archive"), ErrInvalidInput)
	require.ErrorIs(t, validateAdminProjectStatusChange("active", "review"), ErrInvalidInput)
}

func errInvalid[T any](value T, err error) error {
	return err
}
