package models

import (
	"encoding/json"
	"net/http"
)

// Error codes
const (
	ErrCodeBadRequest     = "bad_request"
	ErrCodeNotFound       = "not_found"
	ErrCodeConflict       = "conflict"
	ErrCodeInternalError  = "internal_error"
	ErrCodeValidation     = "validation_error"
	ErrCodeUnauthorized   = "unauthorized"
	ErrCodeForbidden      = "forbidden"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

// WithDetails adds additional details to the error response
func (e *ErrorResponse) WithDetails(details map[string]interface{}) *ErrorResponse {
	e.Error.Details = details
	return e
}

// WriteJSON writes the error response as JSON to the HTTP response writer
func (e *ErrorResponse) WriteJSON(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(e)
}

// Helper functions for common error responses

// BadRequest creates a 400 Bad Request error response
func BadRequest(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeBadRequest, message)
}

// NotFound creates a 404 Not Found error response
func NotFound(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeNotFound, message)
}

// Conflict creates a 409 Conflict error response
func Conflict(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeConflict, message)
}

// InternalError creates a 500 Internal Server Error response
func InternalError(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeInternalError, message)
}

// ValidationError creates a 400 Bad Request error response for validation failures
func ValidationError(message string, details map[string]interface{}) *ErrorResponse {
	return NewErrorResponse(ErrCodeValidation, message).WithDetails(details)
}

// Unauthorized creates a 401 Unauthorized error response
func Unauthorized(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeUnauthorized, message)
}

// Forbidden creates a 403 Forbidden error response
func Forbidden(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeForbidden, message)
}

