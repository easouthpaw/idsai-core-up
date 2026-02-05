package rbac

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope, now time.Time) (bool, error)
}
