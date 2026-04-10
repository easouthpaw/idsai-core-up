package requestctx

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestWithIdentityAndReaders(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	facultyID := uuid.New()
	departmentID := uuid.New()

	ctx := WithIdentity(context.Background(), userID, tenantID, facultyID, departmentID)

	if got, ok := UserID(ctx); !ok || got != userID {
		t.Fatalf("unexpected user id: ok=%v got=%v", ok, got)
	}
	if got, ok := TenantID(ctx); !ok || got != tenantID {
		t.Fatalf("unexpected tenant id: ok=%v got=%v", ok, got)
	}
	if got, ok := FacultyID(ctx); !ok || got != facultyID {
		t.Fatalf("unexpected faculty id: ok=%v got=%v", ok, got)
	}
	if got, ok := DepartmentID(ctx); !ok || got != departmentID {
		t.Fatalf("unexpected department id: ok=%v got=%v", ok, got)
	}
}

func TestReadersReturnFalseWhenIdentityMissing(t *testing.T) {
	ctx := context.Background()

	if _, ok := UserID(ctx); ok {
		t.Fatalf("expected missing user id")
	}
	if _, ok := TenantID(ctx); ok {
		t.Fatalf("expected missing tenant id")
	}
	if _, ok := FacultyID(ctx); ok {
		t.Fatalf("expected missing faculty id")
	}
	if _, ok := DepartmentID(ctx); ok {
		t.Fatalf("expected missing department id")
	}
}
