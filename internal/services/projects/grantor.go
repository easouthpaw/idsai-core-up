package projects

import (
	"context"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

type RoleGrantor interface {
	GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error
}
