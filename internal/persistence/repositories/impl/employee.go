package impl

import (
	"context"
	"errors"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(ctx context.Context, emp *models.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id int) (*models.Employee, error) {
	var emp models.Employee
	err := r.db.WithContext(ctx).First(&emp, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepository) GetByDepartment(ctx context.Context, departmentID int) ([]models.Employee, error) {
	var emps []models.Employee
	if err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("created_at ASC").
		Find(&emps).Error; err != nil {
		return nil, err
	}
	return emps, nil
}

func (r *EmployeeRepository) DeleteByDepartment(ctx context.Context, departmentID int) error {
	return r.db.WithContext(ctx).Where("department_id = ?", departmentID).Delete(&models.Employee{}).Error
}

func (r *EmployeeRepository) ReassignDepartment(ctx context.Context, fromDeptID int, toDeptID int) error {
	return r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID).Error
}

func (r *EmployeeRepository) ReassignDepartments(ctx context.Context, fromDeptIDs []int, toDeptID int) error {
	if len(fromDeptIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("department_id IN ?", fromDeptIDs).
		Update("department_id", toDeptID).Error
}
