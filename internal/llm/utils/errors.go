package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents standard error codes
type ErrorCode string

const (
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrLLMError          ErrorCode = "LLM_ERROR"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrInternalError     ErrorCode = "INTERNAL_ERROR"
)

// APIError represents a structured error response
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// WriteAPIError writes a structured error response
func WriteAPIError(w http.ResponseWriter, code ErrorCode, message string, details string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errResp := APIError{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
			Details: details,
		},
	}
	
	json.NewEncoder(w).Encode(errResp)
}

// WriteInvalidRequestError writes an invalid request error
func WriteInvalidRequestError(w http.ResponseWriter, message string) {
	WriteAPIError(w, ErrInvalidRequest, message, "", http.StatusBadRequest)
}

// WriteInternalError writes an internal server error
func WriteInternalError(w http.ResponseWriter, message string, details string) {
	WriteAPIError(w, ErrInternalError, message, details, http.StatusInternalServerError)
}

// WriteLLMError writes an LLM service error
func WriteLLMError(w http.ResponseWriter, message string, details string) {
	WriteAPIError(w, ErrLLMError, message, details, http.StatusInternalServerError)
}

// WriteNotFoundError writes a not found error
func WriteNotFoundError(w http.ResponseWriter, message string) {
	WriteAPIError(w, ErrNotFound, message, "", http.StatusNotFound)
}

// WriteUnauthorizedError writes an unauthorized error
func WriteUnauthorizedError(w http.ResponseWriter, message string) {
	WriteAPIError(w, ErrUnauthorized, message, "", http.StatusUnauthorized)
}

// WriteRateLimitError writes a rate limit error
func WriteRateLimitError(w http.ResponseWriter, message string) {
	WriteAPIError(w, ErrRateLimitExceeded, message, "", http.StatusTooManyRequests)
}
