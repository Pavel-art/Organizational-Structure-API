package apperrors

import "errors"

const (
	CodeNotFound         = "not_found"
	CodeValidation       = "validation_error"
	CodeConflict         = "conflict"
	CodeInternal         = "internal_error"
	CodeBadRequest       = "bad_request"
	CodeRateLimited      = "rate_limited"
	CodeMethodNotAllowed = "method_not_allowed"

	MsgDepartmentNotFound = "department not found"
	MsgEmployeeNotFound   = "employee not found"
	MsgNameNotUnique      = "department name must be unique within parent"
	MsgCycleDetected      = "cycle detected"
)

var (
	ErrNotFound   = errors.New(MsgDepartmentNotFound)
	ErrConflict   = errors.New(MsgNameNotUnique)
	ErrCycle      = errors.New(MsgCycleDetected)
	ErrBadRequest = errors.New("bad request")
)
