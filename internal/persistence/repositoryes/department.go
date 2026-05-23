package services

import (
	"context"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

type DepartmentRepository interface {
	Create(ctx context.Context, dept *models.Department) error
	GetByID(ctx context.Context, id int) (*models.Department, error)

	Update(ctx context.Context, dept *models.Department) error

	Delete(ctx context.Context, id int) error

	// для валидации уникальности имени внутри parent
	ExistsByNameAndParent(ctx context.Context, name string, parentID *int, excludeID *int) (bool, error)

	// для проверки циклов и дерева
	GetChildrenRecursive(ctx context.Context, id int) ([]models.Department, error)

	// загрузка дерева
	GetTree(ctx context.Context, id int, depth int) (*models.Department, error)
}
