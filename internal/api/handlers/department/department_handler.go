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

type Handler struct {
	mux         *http.ServeMux
	deptService services.DepartmentService
	empService  services.EmployeeService
}

func NewDepartmentHandler(deptService services.DepartmentService, empService services.EmployeeService) *Handler {
	h := &Handler{
		mux:         http.NewServeMux(),
		deptService: deptService,
		empService:  empService,
	}
	h.registerRoutes()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.normalizePath(r)
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("POST /departments", h.CreateDepartment)
	h.mux.HandleFunc("POST /departments/", h.CreateDepartment)

	h.mux.HandleFunc("GET /departments/{id}", h.GetDepartment)
	h.mux.HandleFunc("GET /departments/{id}/", h.GetDepartment)

	h.mux.HandleFunc("PATCH /departments/{id}", h.UpdateDepartment)
	h.mux.HandleFunc("PATCH /departments/{id}/", h.UpdateDepartment)

	h.mux.HandleFunc("DELETE /departments/{id}", h.DeleteDepartment)
	h.mux.HandleFunc("DELETE /departments/{id}/", h.DeleteDepartment)

	h.mux.HandleFunc("POST /departments/{id}/employees", h.CreateEmployee)
	h.mux.HandleFunc("POST /departments/{id}/employees/", h.CreateEmployee)
}

func (h *Handler) normalizePath(r *http.Request) {
	if r.URL == nil {
		return
	}
	if r.URL.Path == "" || r.URL.Path == "/" {
		return
	}
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	deptID, ok := parsePositiveIntPathValue(w, r, "id", "invalid department id")
	if !ok {
		return
	}

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

func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	deptID, ok := parsePositiveIntPathValue(w, r, "id", "invalid department id")
	if !ok {
		return
	}

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

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	deptID, ok := parsePositiveIntPathValue(w, r, "id", "invalid department id")
	if !ok {
		return
	}

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

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	deptID, ok := parsePositiveIntPathValue(w, r, "id", "invalid department id")
	if !ok {
		return
	}

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

func parsePositiveIntPathValue(w http.ResponseWriter, r *http.Request, key string, msg string) (int, bool) {
	raw := r.PathValue(key)
	if raw == "" {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, msg)
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		response.WriteError(w, r, http.StatusBadRequest, apperrors.CodeValidation, msg)
		return 0, false
	}
	return v, true
}
