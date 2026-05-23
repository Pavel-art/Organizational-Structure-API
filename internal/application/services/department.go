package services

import (
	"context"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

type DepartmentService interface {
	Create(ctx context.Context, name string, parentID *int) (*models.Department, error)

	Get(ctx context.Context, id int, depth int, includeEmployees bool) (*models.Department, error)

	Update(ctx context.Context, id int, name *string, parentID *int) (*models.Department, error)

	Delete(ctx context.Context, id int, mode string, reassignTo *int) error
}
