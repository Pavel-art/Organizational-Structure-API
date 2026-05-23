package services

import (
	"context"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

type EmployeeService interface {
	Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *string) (*models.Employee, error)
}
