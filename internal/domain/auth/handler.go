package auth

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/response"
	"pocket-app/internal/pkg/validator"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "body", Message: "Invalid request body"},
		})
	}

	if errs := validator.Validate(req); errs != nil {
		return apperror.ValidationError(errs)
	}

	res, err := h.service.Login(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Login success")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Client handles token deletion
	return response.Success(c, nil, "Logout success")
}
