package rbac

import (
	"context"

	"github.com/google/uuid"
)

type Authorizer interface {
	Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error)
}
