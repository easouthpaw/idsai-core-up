package requestctx

import (
	"context"

	"github.com/google/uuid"
)

type key string

const (
	userIDKey       key = "requestctx.user_id"
	tenantIDKey     key = "requestctx.tenant_id"
	facultyIDKey    key = "requestctx.faculty_id"
	departmentIDKey key = "requestctx.department_id"
)

func WithIdentity(ctx context.Context, userID, tenantID, facultyID, departmentID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, tenantIDKey, tenantID)
	ctx = context.WithValue(ctx, facultyIDKey, facultyID)
	ctx = context.WithValue(ctx, departmentIDKey, departmentID)
	return ctx
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func TenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantIDKey).(uuid.UUID)
	return id, ok
}

func FacultyID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(facultyIDKey).(uuid.UUID)
	return id, ok
}

func DepartmentID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(departmentIDKey).(uuid.UUID)
	return id, ok
}
