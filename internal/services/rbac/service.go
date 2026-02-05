package rbac

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidScope = errors.New("invalid scope")

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

// Can checks whether user can perform permission in given scope.
func (s *Service) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope Scope) (bool, error) {
	if !scope.Validate() {
		return false, ErrInvalidScope
	}
	return s.repo.HasPermission(ctx, userID, permissionCode, scope, s.now())
}

// SetNow is useful for deterministic unit tests.
func (s *Service) SetNow(f func() time.Time) {
	s.now = f
}
