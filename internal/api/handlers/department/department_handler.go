package department

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/internal/api/constants"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/dto"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/mapper"
	"github.com/Pavel-art/Organizational-Structure-API/internal/application/services"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/apperrors"
	"github.com/Pavel-art/Organizational-Structure-API/pkg/response"
	"github.com/Pavel-art/Organizational-Structure-API/pkg/validator"
)

const (
	pathDepartments = "/departments"
)

type Handler struct {
	deptService services.DepartmentService
	empService  services.EmployeeService
}

func NewDepartmentHandler(deptService services.DepartmentService, empService services.EmployeeService) *Handler {
	return &Handler{deptService: deptService, empService: empService}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == pathDepartments {
		if r.Method == http.MethodPost {
			h.createDepartment(w, r)
			return
		}
		response.WriteError(w, r, http.StatusMethodNotAllowed, apperrors.CodeMethodNotAllowed, "method not allowed")
		return
	}

	if !strings.HasPrefix(path, pathDepartments+"/") {
		http.NotFound(w, r)
		return
	}

	rest := strings.TrimPrefix(path, pathDepartments+"/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid department id")
		return
	}

	if len(parts) == 2 && parts[1] == "employees" {
		if r.Method == http.MethodPost {
			h.createEmployee(w, r, id)
			return
		}
		response.WriteError(w, r, http.StatusMethodNotAllowed, apperrors.CodeMethodNotAllowed, "method not allowed")
		return
	}

	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getDepartment(w, r, id)
	case http.MethodPatch:
		h.updateDepartment(w, r, id)
	case http.MethodDelete:
		h.deleteDepartment(w, r, id)
	default:
		response.WriteError(w, r, http.StatusMethodNotAllowed, apperrors.CodeMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createDepartment(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepartmentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	name, ok := validator.TrimAndValidateString(req.Name, constants.NameMaxLen)
	if !ok {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid name")
		return
	}

	dept, err := h.deptService.Create(r.Context(), name, req.ParentID)
	if err != nil {
		response.WriteErrorFromErr(w, r, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, mapper.ToDepartmentDTO(dept))
}

func (h *Handler) createEmployee(w http.ResponseWriter, r *http.Request, deptID int) {
	var req dto.CreateEmployeeRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fullName, ok := validator.TrimAndValidateString(req.FullName, constants.FullNameMaxLen)
	if !ok {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid full_name")
		return
	}
	position, ok := validator.TrimAndValidateString(req.Position, constants.PositionMaxLen)
	if !ok {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid position")
		return
	}

	var hiredAt *time.Time
	if req.HiredAt != nil {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.HiredAt))
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid hired_at (expected YYYY-MM-DD)")
			return
		}
		hiredAt = &parsed
	}

	emp, err := h.empService.Create(r.Context(), deptID, fullName, position, hiredAt)
	if err != nil {
		response.WriteErrorFromErr(w, r, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, mapper.ToEmployeeDTO(*emp))
}

func (h *Handler) getDepartment(w http.ResponseWriter, r *http.Request, deptID int) {
	depth, ok := validator.ParseIntQuery(r, constants.QueryDepth, constants.DefaultDepth)
	if !ok || depth < 1 || depth > constants.MaxDepth {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid depth")
		return
	}
	includeEmployees, ok := validator.ParseBoolQuery(r, constants.QueryIncludeEmployees, true)
	if !ok {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid include_employees")
		return
	}

	dept, err := h.deptService.Get(r.Context(), deptID, depth, includeEmployees)
	if err != nil {
		response.WriteErrorFromErr(w, r, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, mapper.ToDepartmentNodeResponse(dept, includeEmployees))
}

func (h *Handler) updateDepartment(w http.ResponseWriter, r *http.Request, deptID int) {
	var req dto.UpdateDepartmentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Name != nil {
		name, ok := validator.TrimAndValidateString(*req.Name, constants.NameMaxLen)
		if !ok {
			response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid name")
			return
		}
		req.Name = &name
	}

	dept, err := h.deptService.Update(r.Context(), deptID, req.Name, req.ParentID)
	if err != nil {
		response.WriteErrorFromErr(w, r, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, mapper.ToDepartmentDTO(dept))
}

func (h *Handler) deleteDepartment(w http.ResponseWriter, r *http.Request, deptID int) {
	mode := r.URL.Query().Get(constants.QueryMode)
	if mode == "" {
		mode = implDeleteModeCascade
	}

	var reassignTo *int
	if mode == implDeleteModeReassign {
		raw := r.URL.Query().Get(constants.QueryReassignTo)
		if raw == "" {
			response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "reassign_to_department_id is required for mode=reassign")
			return
		}
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid reassign_to_department_id")
			return
		}
		reassignTo = &v
	}

	if err := h.deptService.Delete(r.Context(), deptID, mode, reassignTo); err != nil {
		response.WriteErrorFromErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const (
	implDeleteModeCascade  = "cascade"
	implDeleteModeReassign = "reassign"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, "invalid json body")
		return false
	}
	return true
}
