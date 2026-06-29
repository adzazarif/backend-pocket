package response

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/apperror"
)

type PaginationMeta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"totalPage"`
}

type singleResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type paginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type errorResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Details []apperror.FieldError `json:"details,omitempty"`
}

func Success(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(singleResponse{
		Data:    data,
		Message: message,
	})
}

func Created(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusCreated).JSON(singleResponse{
		Data:    data,
		Message: message,
	})
}

func Paginated(c *fiber.Ctx, data interface{}, meta PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(paginatedResponse{
		Data: data,
		Meta: meta,
	})
}

// Error generates an error response. Note: usually called by the global error handler.
func Error(c *fiber.Ctx, status int, code string, message string, details []apperror.FieldError) error {
	return c.Status(status).JSON(errorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
