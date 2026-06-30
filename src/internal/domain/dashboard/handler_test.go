package dashboard

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/middleware"
)

func setupTestApp(handler *DashboardHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	
	// Mock auth middleware directly setting locals
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "test-user-id")
		return c.Next()
	})

	app.Get("/dashboard", handler.GetSummary)
	return app
}

func TestDashboardHandler_GetSummary(t *testing.T) {
	mockSvc := new(MockDashboardService)
	handler := NewDashboardHandler(mockSvc)
	app := setupTestApp(handler)

	t.Run("Success Get Summary", func(t *testing.T) {
		mockRes := &DashboardSummaryResponse{
			TotalItems: 10,
			UnreadItems: 5,
		}

		mockSvc.On("GetSummary", mock.Anything, "test-user-id").Return(mockRes, nil).Once()

		req := httptest.NewRequest("GET", "/dashboard", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}
