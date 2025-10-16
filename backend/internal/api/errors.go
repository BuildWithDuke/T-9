package api

import (
	"net/http"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeInvalidInput  ErrorType = "invalid_input"
	ErrorTypeGameLogic     ErrorType = "game_logic"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeUnauthorized  ErrorType = "unauthorized"
)

// APIError represents a structured API error
type APIError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Code    int       `json:"code"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(errorType ErrorType, message string, code int) *APIError {
	return &APIError{
		Type:    errorType,
		Message: message,
		Code:    code,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *APIError {
	return NewAPIError(ErrorTypeNotFound, resource+" not found", http.StatusNotFound)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string) *APIError {
	return NewAPIError(ErrorTypeInvalidInput, message, http.StatusBadRequest)
}

// NewGameLogicError creates a game logic error
func NewGameLogicError(message string) *APIError {
	return NewAPIError(ErrorTypeGameLogic, message, http.StatusBadRequest)
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *APIError {
	return NewAPIError(ErrorTypeInternal, message, http.StatusInternalServerError)
}

// WithDetails adds details to an API error
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}