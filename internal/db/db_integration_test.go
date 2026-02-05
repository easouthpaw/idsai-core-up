//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"

	"github.com/stretchr/testify/require"
)

func TestDB_Integration_Ping(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn, "DATABASE_URL is required for integration tests")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))
}
