package middleware

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/response"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	log.Printf("Error: %v\n", err)

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.Error(c, fiberErr.Code, "HTTP_ERROR", fiberErr.Message, nil)
	}

	// Default 500 error
	return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}
