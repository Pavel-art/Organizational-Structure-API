package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Pavel-art/Organizational-Structure-API/internal/core/apperrors"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	WriteJSON(w, status, ErrorResponse{Code: code, Message: message})
	log.Ctx(r.Context()).Warn().Int("status", status).Str("code", code).Msg(message)
}

func WriteErrorFromErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		WriteError(w, r, http.StatusNotFound, apperrors.CodeNotFound, err.Error())
	case errors.Is(err, apperrors.ErrConflict) || errors.Is(err, apperrors.ErrCycle):
		WriteError(w, r, http.StatusConflict, apperrors.CodeConflict, err.Error())
	case errors.Is(err, apperrors.ErrBadRequest):
		WriteError(w, r, http.StatusBadRequest, apperrors.CodeBadRequest, err.Error())
	default:
		WriteError(w, r, http.StatusInternalServerError, apperrors.CodeInternal, "internal error")
		log.Ctx(r.Context()).Error().Err(err).Msg("unexpected error")
	}
}
