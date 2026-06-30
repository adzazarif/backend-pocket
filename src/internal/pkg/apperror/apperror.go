package apperror

import "net/http"

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Details    []FieldError
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NotFound(msg string) *AppError {
	return &AppError{
		Code:       "POCKET_NOT_FOUND", // We use a generic not found code based on the technical plan, but this might need adjustment per domain. The plan says "POCKET_NOT_FOUND" 404 for item not found.
		Message:    msg,
		HTTPStatus: http.StatusNotFound,
	}
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func ValidationError(details []FieldError) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation error",
		HTTPStatus: http.StatusUnprocessableEntity, // 422
		Details:    details,
	}
}

func Internal(msg string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    msg,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func InvalidCredential() *AppError {
	return &AppError{
		Code:       "INVALID_CREDENTIAL",
		Message:    "Email or password is incorrect",
		HTTPStatus: http.StatusUnauthorized,
	}
}
