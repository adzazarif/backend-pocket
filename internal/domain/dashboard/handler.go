package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/pkg/response"
)

type DashboardHandler struct {
	service DashboardService
}

func NewDashboardHandler(service DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetSummary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	res, err := h.service.GetSummary(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.Success(c, res, "Dashboard loaded")
}
