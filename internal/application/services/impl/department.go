package impl

import (
	"context"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/apperrors"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
	"github.com/Pavel-art/Organizational-Structure-API/internal/persistence/repositories"
	repoimpl "github.com/Pavel-art/Organizational-Structure-API/internal/persistence/repositories/impl"
	"gorm.io/gorm"
)

const (
	DeleteModeCascade  = "cascade"
	DeleteModeReassign = "reassign"
)

type DepartmentService struct {
	db       *gorm.DB
	deptRepo repositories.DepartmentRepository
	empRepo  repositories.EmployeeRepository
}

func NewDepartmentService(db *gorm.DB, deptRepo repositories.DepartmentRepository, empRepo repositories.EmployeeRepository) *DepartmentService {
	return &DepartmentService{db: db, deptRepo: deptRepo, empRepo: empRepo}
}

func (s *DepartmentService) Create(ctx context.Context, name string, parentID *int) (*models.Department, error) {
	if parentID != nil {
		parent, err := s.deptRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, apperrors.ErrNotFound
		}
	}

	exists, err := s.deptRepo.ExistsByNameAndParent(ctx, name, parentID, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict
	}

	dept := &models.Department{Name: name, ParentID: parentID}
	if err := s.deptRepo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) Get(ctx context.Context, id int, depth int, includeEmployees bool) (*models.Department, error) {
	dept, err := s.deptRepo.GetTree(ctx, id, depth, includeEmployees)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, apperrors.ErrNotFound
	}
	return dept, nil
}

func (s *DepartmentService) Update(ctx context.Context, id int, name *string, parentID *int) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, apperrors.ErrNotFound
	}

	if parentID != nil {
		if *parentID == id {
			return nil, apperrors.ErrCycle
		}

		parent, err := s.deptRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, apperrors.ErrNotFound
		}

		children, err := s.deptRepo.GetChildrenRecursive(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, c := range children {
			if c.ID == *parentID {
				return nil, apperrors.ErrCycle
			}
		}
	}

	newName := dept.Name
	if name != nil {
		newName = *name
	}
	newParentID := dept.ParentID
	if parentID != nil {
		newParentID = parentID
	}

	excludeID := &id
	exists, err := s.deptRepo.ExistsByNameAndParent(ctx, newName, newParentID, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict
	}

	if name != nil {
		dept.Name = *name
	}
	if parentID != nil {
		dept.ParentID = parentID
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) Delete(ctx context.Context, id int, mode string, reassignTo *int) error {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dept == nil {
		return apperrors.ErrNotFound
	}

	switch mode {
	case DeleteModeCascade:
		return s.deptRepo.Delete(ctx, id)
	case DeleteModeReassign:
		if reassignTo == nil {
			return apperrors.ErrBadRequest
		}
		if *reassignTo == id {
			return apperrors.ErrBadRequest
		}

		target, err := s.deptRepo.GetByID(ctx, *reassignTo)
		if err != nil {
			return err
		}
		if target == nil {
			return apperrors.ErrNotFound
		}

		children, err := s.deptRepo.GetChildrenRecursive(ctx, id)
		if err != nil {
			return err
		}
		fromIDs := make([]int, 0, len(children)+1)
		fromIDs = append(fromIDs, id)
		for _, c := range children {
			fromIDs = append(fromIDs, c.ID)
		}

		return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			deptRepo := repoimpl.NewDepartmentRepository(tx)
			empRepo := repoimpl.NewEmployeeRepository(tx)

			if err := empRepo.ReassignDepartments(ctx, fromIDs, *reassignTo); err != nil {
				return err
			}
			if err := deptRepo.Delete(ctx, id); err != nil {
				return err
			}
			return nil
		})
	default:
		return apperrors.ErrBadRequest
	}
}

var _ interface {
	Create(ctx context.Context, name string, parentID *int) (*models.Department, error)
	Get(ctx context.Context, id int, depth int, includeEmployees bool) (*models.Department, error)
	Update(ctx context.Context, id int, name *string, parentID *int) (*models.Department, error)
	Delete(ctx context.Context, id int, mode string, reassignTo *int) error
} = (*DepartmentService)(nil)
