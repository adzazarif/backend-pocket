package pocket

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/response"
	"pocket-app/internal/pkg/validator"
)

type PocketHandler struct {
	service PocketService
}

func NewPocketHandler(service PocketService) *PocketHandler {
	return &PocketHandler{service: service}
}

func (h *PocketHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req CreatePocketRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "body", Message: "Invalid request body"},
		})
	}

	if errs := validator.Validate(req); errs != nil {
		return apperror.ValidationError(errs)
	}

	res, err := h.service.Create(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return response.Created(c, res, "Pocket item created successfully")
}

func (h *PocketHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	var req UpdatePocketRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "body", Message: "Invalid request body"},
		})
	}

	if errs := validator.Validate(req); errs != nil {
		return apperror.ValidationError(errs)
	}

	res, err := h.service.Update(c.Context(), id, userID, req)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Pocket item updated successfully")
}

func (h *PocketHandler) GetDetail(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	res, err := h.service.GetDetail(c.Context(), id, userID)
	if err != nil {
		return err
	}

	return response.Success(c, res, "")
}

func (h *PocketHandler) Archive(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	res, err := h.service.Archive(c.Context(), id, userID)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Pocket item archived successfully")
}

func (h *PocketHandler) UpdateStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "body", Message: "Invalid request body"},
		})
	}

	if errs := validator.Validate(req); errs != nil {
		return apperror.ValidationError(errs)
	}

	res, err := h.service.UpdateStatus(c.Context(), id, userID, req)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Status updated successfully")
}

func (h *PocketHandler) ToggleFavorite(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	var req ToggleFavoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "body", Message: "Invalid request body"},
		})
	}

	if errs := validator.Validate(req); errs != nil {
		return apperror.ValidationError(errs)
	}

	res, err := h.service.ToggleFavorite(c.Context(), id, userID, req)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Favorite updated successfully")
}

func (h *PocketHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var query PocketListQuery
	if err := c.QueryParser(&query); err != nil {
		return apperror.ValidationError([]apperror.FieldError{
			{Field: "query", Message: "Invalid query parameters"},
		})
	}

	items, meta, err := h.service.List(c.Context(), userID, query)
	if err != nil {
		return err
	}

	return response.Paginated(c, items, meta)
}
