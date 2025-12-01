package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// General errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Authentication errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"
	ErrCodeAccountDisabled    ErrorCode = "ACCOUNT_DISABLED"

	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeRecordNotFound     ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDuplicateEntry     ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"

	// Module errors
	ErrCodeModuleNotFound    ErrorCode = "MODULE_NOT_FOUND"
	ErrCodeModuleDisabled    ErrorCode = "MODULE_DISABLED"
	ErrCodeModuleInstallFail ErrorCode = "MODULE_INSTALL_FAIL"
	
	// Permission errors
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrCodeInvalidRole             ErrorCode = "INVALID_ROLE"
)

// AppError represents application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithError adds underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common error constructors
func NewBadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func NewUnauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NewNotFound(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound)
}

func NewConflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func NewInternal(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

func NewValidationError(message string, details map[string]interface{}) *AppError {
	return New(ErrCodeValidation, message, http.StatusUnprocessableEntity).WithDetails(details)
}

// IsAppError checks if error is AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError tries to extract AppError from error
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
