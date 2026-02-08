package projects

import (
	"context"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error)
}
