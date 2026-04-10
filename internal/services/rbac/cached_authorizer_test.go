package rbac

import (
	"context"
	"errors"
	"testing"
	"time"

	"idsai-core-up/internal/infra/cache"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubAuthorizer struct {
	canResult       bool
	canErr          error
	canCalls        int
	canAllResult    bool
	canAllErr       error
	canAllCalls     int
	canWithAttrsOK  bool
	canWithAttrsErr error
	attrCalls       int
	listCodes       []string
	listErr         error
	listCalls       int
}

func (s *stubAuthorizer) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error) {
	s.canCalls++
	return s.canResult, s.canErr
}

func (s *stubAuthorizer) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope Scope) (bool, error) {
	s.canAllCalls++
	return s.canAllResult, s.canAllErr
}

func (s *stubAuthorizer) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope, attrs map[string]interface{}) (bool, error) {
	s.attrCalls++
	return s.canWithAttrsOK, s.canWithAttrsErr
}

func (s *stubAuthorizer) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope Scope) ([]string, error) {
	s.listCalls++
	return append([]string(nil), s.listCodes...), s.listErr
}

func newTestCache(t *testing.T) (*miniredis.Miniredis, *cache.RedisClient) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := cache.NewRedisClient(mr.Addr(), "", 0)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
		mr.Close()
	})

	return mr, client
}

func TestCacheKey_FormatsSystemAndProjectScopes(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	projectID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	require.Equal(t,
		"rbac:user:11111111-1111-1111-1111-111111111111:perm:task.view:scope:SYSTEM:nil",
		cacheKey(userID, "task.view", Scope{Type: ScopeSystem}),
	)
	require.Equal(t,
		"rbac:user:11111111-1111-1111-1111-111111111111:perm:task.view:scope:PROJECT:22222222-2222-2222-2222-222222222222",
		cacheKey(userID, "task.view", Scope{Type: ScopeProject, ID: &projectID}),
	)
}

func TestCachedAuthorizer_Can_CachesPositiveResult(t *testing.T) {
	mr, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canResult: true}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	userID := uuid.New()
	projectID := uuid.New()
	scope := Scope{Type: ScopeProject, ID: &projectID}
	ctx := context.Background()

	ok, err := cached.Can(ctx, userID, "task.view", scope)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1, inner.canCalls)

	ok, err = cached.Can(ctx, userID, "task.view", scope)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1, inner.canCalls)
	val, err := mr.Get(cacheKey(userID, "task.view", scope))
	require.NoError(t, err)
	require.Equal(t, "1", val)
}

func TestCachedAuthorizer_Can_CachesNegativeResult(t *testing.T) {
	mr, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canResult: false}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	userID := uuid.New()
	projectID := uuid.New()
	scope := Scope{Type: ScopeProject, ID: &projectID}

	ok, err := cached.Can(context.Background(), userID, "task.delete", scope)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = cached.Can(context.Background(), userID, "task.delete", scope)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, 1, inner.canCalls)
	val, err := mr.Get(cacheKey(userID, "task.delete", scope))
	require.NoError(t, err)
	require.Equal(t, "0", val)
}

func TestCachedAuthorizer_Can_DoesNotCacheErrors(t *testing.T) {
	_, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canErr: errors.New("db unavailable")}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	userID := uuid.New()
	projectID := uuid.New()
	scope := Scope{Type: ScopeProject, ID: &projectID}

	ok, err := cached.Can(context.Background(), userID, "task.view", scope)
	require.ErrorContains(t, err, "db unavailable")
	require.False(t, ok)

	ok, err = cached.Can(context.Background(), userID, "task.view", scope)
	require.ErrorContains(t, err, "db unavailable")
	require.False(t, ok)
	require.Equal(t, 2, inner.canCalls)
}

func TestCachedAuthorizer_CanAll_UsesPerPermissionCache(t *testing.T) {
	_, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canResult: true}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	userID := uuid.New()
	projectID := uuid.New()
	scope := Scope{Type: ScopeProject, ID: &projectID}
	permissions := []string{"project.view", "task.view", "task.update"}

	ok, err := cached.CanAll(context.Background(), userID, permissions, scope)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 3, inner.canCalls)

	ok, err = cached.CanAll(context.Background(), userID, permissions, scope)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 3, inner.canCalls)
}

func TestCachedAuthorizer_CanWithAttributes_Delegates(t *testing.T) {
	_, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canWithAttrsOK: true}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	projectID := uuid.New()
	ok, err := cached.CanWithAttributes(context.Background(), uuid.New(), "task.update", Scope{
		Type: ScopeProject,
		ID:   &projectID,
	}, map[string]interface{}{"status": "IN_PROGRESS"})

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1, inner.attrCalls)
}

func TestCachedAuthorizer_ListPermissionCodes_Delegates(t *testing.T) {
	_, redisClient := newTestCache(t)
	inner := &stubAuthorizer{listCodes: []string{"project.view", "task.view"}}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	projectID := uuid.New()
	codes, err := cached.ListPermissionCodes(context.Background(), uuid.New(), Scope{
		Type: ScopeProject,
		ID:   &projectID,
	})

	require.NoError(t, err)
	require.Equal(t, []string{"project.view", "task.view"}, codes)
	require.Equal(t, 1, inner.listCalls)
}

func TestCachedAuthorizer_InvalidateUser_RemovesMatchingKeys(t *testing.T) {
	mr, redisClient := newTestCache(t)
	inner := &stubAuthorizer{canResult: true}
	cached := NewCachedAuthorizer(inner, redisClient, time.Minute)

	userID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	scope := Scope{Type: ScopeProject, ID: &projectID}

	_, err := cached.Can(context.Background(), userID, "task.view", scope)
	require.NoError(t, err)
	_, err = cached.Can(context.Background(), otherUserID, "task.view", scope)
	require.NoError(t, err)

	cached.InvalidateUser(context.Background(), userID)

	require.False(t, mr.Exists(cacheKey(userID, "task.view", scope)))
	require.True(t, mr.Exists(cacheKey(otherUserID, "task.view", scope)))
}
