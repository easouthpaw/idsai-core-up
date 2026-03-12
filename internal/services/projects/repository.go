package projects

import (
	"context"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

type Group struct {
	ID   uuid.UUID
	Code string
	Name string
}

type Repository interface {
	Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error)
	HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error)
	ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error)
	ListPublic(ctx context.Context) ([]domain.Project, error)
	FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error)
	ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]Group, error)
}
