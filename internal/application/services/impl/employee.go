package impl

import (
	"context"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/apperrors"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
	"github.com/Pavel-art/Organizational-Structure-API/internal/persistence/repositories"
)

type EmployeeService struct {
	deptRepo repositories.DepartmentRepository
	empRepo  repositories.EmployeeRepository
}

func NewEmployeeService(deptRepo repositories.DepartmentRepository, empRepo repositories.EmployeeRepository) *EmployeeService {
	return &EmployeeService{deptRepo: deptRepo, empRepo: empRepo}
}

func (s *EmployeeService) Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *time.Time) (*models.Employee, error) {
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, apperrors.ErrNotFound
	}

	emp := &models.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}
	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}
