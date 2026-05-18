package errors

import (
	"errors"
	"net/http"
)

// Domain errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTaskNotFound      = errors.New("task not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// NewAppError creates a new AppError
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Predefined errors
func NewUserNotFoundError() *AppError {
	return NewAppError(ErrUserNotFound, "User not found", http.StatusNotFound)
}

func NewUserAlreadyExistsError() *AppError {
	return NewAppError(ErrUserAlreadyExists, "Não foi possível criar a conta. Verifique os dados informados.", http.StatusBadRequest)
}

func NewInvalidCredentialsError() *AppError {
	return NewAppError(ErrInvalidCredentials, "Invalid credentials", http.StatusUnauthorized)
}

func NewTaskNotFoundError() *AppError {
	return NewAppError(ErrTaskNotFound, "Task not found", http.StatusNotFound)
}

func NewUnauthorizedError() *AppError {
	return NewAppError(ErrUnauthorized, "Unauthorized", http.StatusUnauthorized)
}

func NewForbiddenError() *AppError {
	return NewAppError(ErrForbidden, "You don't have permission to access this resource", http.StatusForbidden)
}

func NewInvalidInputError(message string) *AppError {
	return NewAppError(ErrInvalidInput, message, http.StatusBadRequest)
}

func NewInternalServerError(err error) *AppError {
	return NewAppError(err, "Internal server error", http.StatusInternalServerError)
}

