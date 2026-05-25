package department

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
	"github.com/stretchr/testify/require"
)

type mockDeptService struct{}

func (m mockDeptService) Create(ctx context.Context, name string, parentID *int) (*models.Department, error) {
	return &models.Department{ID: 1, Name: name, ParentID: parentID}, nil
}
func (m mockDeptService) Get(ctx context.Context, id int, depth int, includeEmployees bool) (*models.Department, error) {
	return &models.Department{ID: id, Name: "Root"}, nil
}
func (m mockDeptService) Update(ctx context.Context, id int, name *string, parentID *int) (*models.Department, error) {
	return &models.Department{ID: id, Name: "Updated"}, nil
}
func (m mockDeptService) Delete(ctx context.Context, id int, mode string, reassignTo *int) error {
	return nil
}

type mockEmpService struct{}

func (m mockEmpService) Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *time.Time) (*models.Employee, error) {
	return &models.Employee{ID: 1, DepartmentID: departmentID, FullName: fullName, Position: position}, nil
}

func TestCreateDepartmentValidation(t *testing.T) {
	h := NewDepartmentHandler(mockDeptService{}, mockEmpService{})

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(`{"name":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}
