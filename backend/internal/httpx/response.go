package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ErrorCode represents a stable machine-readable error identifier.
type ErrorCode string

const (
	// ErrorCodeBadRequest indicates validation or input issues.
	ErrorCodeBadRequest ErrorCode = "BadRequest"
	// ErrorCodeUnauthorized indicates authentication failures.
	ErrorCodeUnauthorized ErrorCode = "Unauthorized"
	// ErrorCodeForbidden indicates authorization failures.
	ErrorCodeForbidden ErrorCode = "Forbidden"
	// ErrorCodeNotFound indicates missing resources.
	ErrorCodeNotFound ErrorCode = "NotFound"
	// ErrorCodeConflict indicates conflicting resource state.
	ErrorCodeConflict ErrorCode = "Conflict"
	// ErrorCodeTooManyRequests indicates rate-limiting errors.
	ErrorCodeTooManyRequests ErrorCode = "TooManyRequests"
	// ErrorCodeInternal indicates an unexpected server error.
	ErrorCodeInternal ErrorCode = "InternalError"
)

// APIError is a typed application error for consistent error handling.
type APIError struct {
	Status  int
	Code    ErrorCode
	Message string
	err     error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.Message
}

// Unwrap exposes the underlying error for errors.Is/As checks.
func (e *APIError) Unwrap() error {
	return e.err
}

// NewError creates a new APIError without an underlying cause.
func NewError(status int, code ErrorCode, message string) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an underlying error with API metadata.
func WrapError(err error, status int, code ErrorCode, message string) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
		err:     err,
	}
}

// ErrorPayload is the JSON envelope for errors.
type ErrorPayload struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody represents the error details payload.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON sends a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

// WriteError sends a structured error response.
func WriteError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	WriteJSON(w, status, ErrorPayload{
		Error: ErrorBody{
			Code:    string(code),
			Message: message,
		},
	})
}

// WriteAPIError sends an error response derived from APIError or defaults to InternalError.
func WriteAPIError(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		WriteError(w, apiErr.Status, apiErr.Code, apiErr.Message)
		return
	}

	log.Printf("unhandled error: %v", err)
	WriteError(w, http.StatusInternalServerError, ErrorCodeInternal, "Internal server error")
}
