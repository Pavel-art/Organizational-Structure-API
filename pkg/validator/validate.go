package validator

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Pavel-art/Organizational-Structure-API/internal/api/constants"
)

func TrimAndValidateString(v string, maxLen int) (string, bool) {
	s := strings.TrimSpace(v)
	if len(s) < constants.MinStringLength || len(s) > maxLen {
		return "", false
	}
	return s, true
}

func ParseIntQuery(r *http.Request, key string, defaultValue int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, true
	}
	i, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return i, true
}

func ParseBoolQuery(r *http.Request, key string, defaultValue bool) (bool, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, true
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}
	return b, true
}
