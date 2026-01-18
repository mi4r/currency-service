package httputil

import (
	"net/http"
)

// ErrorResponse represents an error response structure.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Error writes an error JSON response.
func Error(w http.ResponseWriter, status int, message, code string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// BadRequest writes a 400 Bad Request error response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message, "BAD_REQUEST")
}

// Unauthorized writes a 401 Unauthorized error response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

// Forbidden writes a 403 Forbidden error response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, "FORBIDDEN")
}

// NotFound writes a 404 Not Found error response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, "NOT_FOUND")
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message, "INTERNAL_ERROR")
}

// ValidationError writes a 422 Unprocessable Entity error response.
func ValidationError(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, message, "VALIDATION_ERROR")
}
