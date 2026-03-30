package rbac

import (
	"context"
	"fmt"
	"time"

	"idsai-core-up/internal/infra/cache"

	"github.com/google/uuid"
)

// CachedAuthorizer is a decorator around Authorizer that uses Redis for caching
// permission check results. On cache miss or Redis error, it falls through to
// the inner authorizer (graceful degradation).
type CachedAuthorizer struct {
	inner Authorizer
	cache *cache.RedisClient
	ttl   time.Duration
}

// NewCachedAuthorizer wraps an Authorizer with Redis caching.
func NewCachedAuthorizer(inner Authorizer, redisClient *cache.RedisClient, ttl time.Duration) *CachedAuthorizer {
	return &CachedAuthorizer{
		inner: inner,
		cache: redisClient,
		ttl:   ttl,
	}
}

// cacheKey builds the Redis key: rbac:user:{user_id}:perm:{permission}:scope:{scope_type}:{scope_id}
func cacheKey(userID uuid.UUID, permissionCode string, scope Scope) string {
	scopeID := "nil"
	if scope.ID != nil {
		scopeID = scope.ID.String()
	}
	return fmt.Sprintf("rbac:user:%s:perm:%s:scope:%s:%s",
		userID.String(), permissionCode, scope.Type, scopeID)
}

// Can checks the cache first; on miss delegates to the inner authorizer.
func (c *CachedAuthorizer) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error) {
	key := cacheKey(userID, permissionCode, scope)

	// Try cache
	val, found, _ := c.cache.Get(ctx, key)
	if found {
		return val == "1", nil
	}

	// Cache miss → delegate
	ok, err := c.inner.Can(ctx, userID, permissionCode, scope)
	if err != nil {
		return false, err
	}

	// Store in cache
	cached := "0"
	if ok {
		cached = "1"
	}
	c.cache.Set(ctx, key, cached, c.ttl)

	return ok, nil
}

// CanAll checks all permissions, each individually cached.
func (c *CachedAuthorizer) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope Scope) (bool, error) {
	for _, perm := range permissions {
		ok, err := c.Can(ctx, userID, perm, scope)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// CanWithAttributes delegates directly to the inner authorizer without caching,
// because attribute-based checks depend on runtime context that changes per request.
func (c *CachedAuthorizer) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope, attrs map[string]interface{}) (bool, error) {
	return c.inner.CanWithAttributes(ctx, userID, permissionCode, scope, attrs)
}

// ListPermissionCodes delegates to the inner authorizer without caching.
func (c *CachedAuthorizer) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope Scope) ([]string, error) {
	return c.inner.ListPermissionCodes(ctx, userID, scope)
}

// InvalidateUser removes all cached RBAC entries for a user.
func (c *CachedAuthorizer) InvalidateUser(ctx context.Context, userID uuid.UUID) {
	pattern := fmt.Sprintf("rbac:user:%s:*", userID.String())
	c.cache.DelByPattern(ctx, pattern)
}
