package repositories

import (
	"context"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *models.Employee) error

	GetByID(ctx context.Context, id int) (*models.Employee, error)

	GetByDepartment(ctx context.Context, departmentID int) ([]models.Employee, error)

	DeleteByDepartment(ctx context.Context, departmentID int) error

	ReassignDepartment(ctx context.Context, fromDeptID int, toDeptID int) error

	ReassignDepartments(ctx context.Context, fromDeptIDs []int, toDeptID int) error
}
