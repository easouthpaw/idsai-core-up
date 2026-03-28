package postgres

import (
	"context"
	"errors"

	"idsai-core-up/internal/requestctx"

	"github.com/google/uuid"
)

var errTenantContextMissing = errors.New("tenant context missing")

func tenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := requestctx.TenantID(ctx)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, errTenantContextMissing
	}
	return tenantID, nil
}
