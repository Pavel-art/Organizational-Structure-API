package mapper

import (
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/internal/api/dto"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/models"
)

func ToEmployeeDTO(m models.Employee) dto.EmployeeDTO {
	var hiredAt *string
	if m.HiredAt != nil {
		s := m.HiredAt.In(time.UTC).Format("2006-01-02")
		hiredAt = &s
	}
	return dto.EmployeeDTO{
		ID:           m.ID,
		DepartmentID: m.DepartmentID,
		FullName:     m.FullName,
		Position:     m.Position,
		HiredAt:      hiredAt,
		CreatedAt:    m.CreatedAt,
	}
}

func ToDepartmentDTO(m *models.Department) dto.DepartmentDTO {
	if m == nil {
		return dto.DepartmentDTO{}
	}

	return dto.DepartmentDTO{
		ID:        m.ID,
		Name:      m.Name,
		ParentID:  m.ParentID,
		CreatedAt: m.CreatedAt,
	}
}

func ToDepartmentNodeResponse(m *models.Department, includeEmployees bool) dto.DepartmentNodeResponse {
	if m == nil {
		return dto.DepartmentNodeResponse{}
	}

	resp := dto.DepartmentNodeResponse{
		Department: ToDepartmentDTO(m),
		Children:   make([]dto.DepartmentNodeResponse, 0, len(m.Children)),
	}

	if includeEmployees {
		employees := make([]dto.EmployeeDTO, 0, len(m.Employees))
		for _, e := range m.Employees {
			employees = append(employees, ToEmployeeDTO(e))
		}
		resp.Employees = &employees
	}

	for i := range m.Children {
		c := m.Children[i]
		resp.Children = append(resp.Children, ToDepartmentNodeResponse(&c, includeEmployees))
	}

	return resp
}
