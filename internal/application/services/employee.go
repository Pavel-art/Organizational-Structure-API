package services

import (
	"context"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

type EmployeeService interface {
	Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *time.Time) (*models.Employee, error)
}
