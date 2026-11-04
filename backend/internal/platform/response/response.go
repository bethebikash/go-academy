package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// SuccessEnvelope wraps all successful API responses.
type SuccessEnvelope struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

// ErrorDetail provides field-level error messages.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorEnvelope wraps all error API responses.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody details the error code, summary, and details.
type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// MetaInfo provides standard metadata for responses.
type MetaInfo struct {
	Timestamp string `json:"timestamp"`
}

// JSON sends a standard success JSON response.
func JSON(w http.ResponseWriter, status int, data any, meta any) {
	if meta == nil {
		meta = MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	envelope := SuccessEnvelope{
		Data: data,
		Meta: meta,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

// Error sends a standard error JSON response.
func Error(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	envelope := ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}
