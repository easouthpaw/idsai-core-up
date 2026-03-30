package rbac

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidScope = errors.New("invalid scope")

// Service is the core RBAC authorizer that delegates to a Repository.
// It implements the Authorizer interface.
type Service struct {
	repo       Repository
	conditions *ConditionRegistry
	now        func() time.Time
}

// NewService creates a new RBAC service.
// If conditions is nil, an empty registry is used (no ABAC conditions).
func NewService(repo Repository, conditions *ConditionRegistry) *Service {
	if conditions == nil {
		conditions = NewConditionRegistry()
	}
	return &Service{
		repo:       repo,
		conditions: conditions,
		now:        time.Now,
	}
}

// Can checks whether user can perform permissionCode in given scope.
func (s *Service) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error) {
	if !scope.Validate() {
		return false, ErrInvalidScope
	}
	return s.repo.HasPermission(ctx, userID, permissionCode, scope, s.now())
}

// CanAll checks whether the user has ALL listed permissions in the given scope.
func (s *Service) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope Scope) (bool, error) {
	if !scope.Validate() {
		return false, ErrInvalidScope
	}
	for _, perm := range permissions {
		ok, err := s.repo.HasPermission(ctx, userID, perm, scope, s.now())
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// CanWithAttributes performs an ABAC check: RBAC permission + attribute conditions.
func (s *Service) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope, attrs map[string]interface{}) (bool, error) {
	ok, err := s.Can(ctx, userID, permissionCode, scope)
	if err != nil || !ok {
		return false, err
	}

	// Evaluate ABAC conditions
	if !s.conditions.Evaluate(permissionCode, attrs) {
		return false, nil
	}
	return true, nil
}

// Conditions returns the condition registry for external registration.
func (s *Service) Conditions() *ConditionRegistry {
	return s.conditions
}

func (s *Service) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope Scope) ([]string, error) {
	if !scope.Validate() {
		return nil, ErrInvalidScope
	}
	return s.repo.ListPermissionCodes(ctx, userID, scope, s.now())
}

// SetNow is useful for deterministic unit tests.
func (s *Service) SetNow(f func() time.Time) {
	s.now = f
}
