package rbac

import (
	"context"

	"github.com/google/uuid"
)

// Authorizer is the primary interface for permission checks.
// Implementations include the base Service and the CachedAuthorizer decorator.
type Authorizer interface {
	// Can checks whether user has a single permission in the given scope.
	Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error)

	// CanAll checks whether user has ALL listed permissions in the given scope.
	// Returns false as soon as any permission is denied.
	CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope Scope) (bool, error)

	// CanWithAttributes performs an ABAC check: standard RBAC + attribute conditions.
	// The attrs map carries runtime context (e.g. task_author_id, project_status).
	CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope, attrs map[string]interface{}) (bool, error)

	// ListPermissionCodes returns all permission codes the user has in the given scope.
	ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope Scope) ([]string, error)
}
