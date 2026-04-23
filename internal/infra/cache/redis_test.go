package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func TestRedisClient_GetSetDelAndPattern(t *testing.T) {
	srv := miniredis.RunT(t)
	client := NewRedisClient(srv.Addr(), "", 0)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	ctx := context.Background()

	value, ok, err := client.Get(ctx, "missing")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, value)

	client.Set(ctx, "rbac:user:1:project:a", "allow", time.Minute)
	client.Set(ctx, "rbac:user:1:project:b", "deny", time.Minute)
	client.Set(ctx, "rbac:user:2:project:a", "allow", time.Minute)

	value, ok, err = client.Get(ctx, "rbac:user:1:project:a")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "allow", value)

	client.DelByPattern(ctx, "rbac:user:1:*")

	_, ok, err = client.Get(ctx, "rbac:user:1:project:a")
	require.NoError(t, err)
	require.False(t, ok)
	_, ok, err = client.Get(ctx, "rbac:user:1:project:b")
	require.NoError(t, err)
	require.False(t, ok)

	value, ok, err = client.Get(ctx, "rbac:user:2:project:a")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "allow", value)

	client.Del(ctx)
	client.Del(ctx, "rbac:user:2:project:a")
	_, ok, err = client.Get(ctx, "rbac:user:2:project:a")
	require.NoError(t, err)
	require.False(t, ok)
}
