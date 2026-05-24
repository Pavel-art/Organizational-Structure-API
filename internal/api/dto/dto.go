package dto

import "time"

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id,omitempty"`
}

type UpdateDepartmentRequest struct {
	Name     *string `json:"name,omitempty"`
	ParentID *int    `json:"parent_id,omitempty"`
}

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at,omitempty"`
}

type DepartmentDTO struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int      `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type EmployeeDTO struct {
	ID           int    `json:"id"`
	DepartmentID int    `json:"department_id"`
	FullName     string `json:"full_name"`
	Position     string `json:"position"`
	// Date-only (YYYY-MM-DD), per ТЗ.
	HiredAt   *string   `json:"hired_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type DepartmentNodeResponse struct {
	Department DepartmentDTO `json:"department"`
	// Employees must be present as [] when include_employees=true.
	// When include_employees=false the field is omitted.
	Employees *[]EmployeeDTO           `json:"employees,omitempty"`
	Children  []DepartmentNodeResponse `json:"children"`
}
