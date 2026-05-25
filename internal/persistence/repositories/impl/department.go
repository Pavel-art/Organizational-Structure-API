package impl

import (
	"context"
	"errors"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(ctx context.Context, dept *models.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id int) (*models.Department, error) {
	var dept models.Department
	err := r.db.WithContext(ctx).First(&dept, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, dept *models.Department) error {
	// Используем Updates вместо Save, чтобы не обновлять нулевые значения
	return r.db.WithContext(ctx).
		Model(&models.Department{}).
		Where("id = ?", dept.ID).
		Updates(dept).Error
}

func (r *DepartmentRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Department{}, "id = ?", id).Error
}

func (r *DepartmentRepository) ExistsByNameAndParent(ctx context.Context, name string, parentID *int, excludeID *int) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.Department{}).Where("name = ?", name)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DepartmentRepository) GetChildrenRecursive(ctx context.Context, id int) ([]models.Department, error) {
	var deps []models.Department
	err := r.db.WithContext(ctx).Raw(`
WITH RECURSIVE deps AS (
  SELECT id, name, parent_id, created_at FROM departments WHERE parent_id = ?
  UNION ALL
  SELECT d.id, d.name, d.parent_id, d.created_at
  FROM departments d
  JOIN deps ON d.parent_id = deps.id
)
SELECT id, name, parent_id, created_at FROM deps
`, id).Scan(&deps).Error
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *DepartmentRepository) GetTree(ctx context.Context, id int, depth int, includeEmployees bool) (*models.Department, error) {
	if depth < 0 {
		depth = 0
	}

	// Для глубины 0 возвращаем только корневой отдел
	if depth == 0 {
		q := r.db.WithContext(ctx).Model(&models.Department{}).Where("id = ?", id)
		if includeEmployees {
			q = q.Preload("Employees", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at ASC")
			})
		}
		var dept models.Department
		err := q.First(&dept).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &dept, nil
	}

	// Получаем корневой отдел
	var dept models.Department
	err := r.db.WithContext(ctx).First(&dept, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Рекурсивно загружаем детей до указанной глубины
	if err := r.loadChildren(ctx, &dept, depth, includeEmployees); err != nil {
		return nil, err
	}

	return &dept, nil
}

// loadChildren рекурсивно загружает детей для отдела
func (r *DepartmentRepository) loadChildren(ctx context.Context, dept *models.Department, remainingDepth int, includeEmployees bool) error {
	if remainingDepth <= 0 {
		return nil
	}

	// Загружаем прямых детей
	var children []models.Department
	query := r.db.WithContext(ctx).Where("parent_id = ?", dept.ID).Order("created_at ASC")

	if includeEmployees {
		query = query.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		})
	}

	err := query.Find(&children).Error
	if err != nil {
		return err
	}

	// Рекурсивно загружаем детей для каждого ребенка
	for i := range children {
		if err := r.loadChildren(ctx, &children[i], remainingDepth-1, includeEmployees); err != nil {
			return err
		}
	}

	dept.Children = children
	return nil
}
