package pocket

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/response"
	"pocket-app/internal/middleware"
)

func setupTestApp(handler *PocketHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "test-user-id")
		return c.Next()
	})

	app.Get("/pockets", handler.List)
	app.Post("/pockets", handler.Create)
	app.Get("/pockets/:id", handler.GetDetail)
	app.Put("/pockets/:id", handler.Update)
	app.Delete("/pockets/:id", handler.Archive)
	app.Patch("/pockets/:id/status", handler.UpdateStatus)
	app.Patch("/pockets/:id/favorite", handler.ToggleFavorite)
	return app
}

func TestPocketHandler_Create(t *testing.T) {
	t.Run("Success Create", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		url := "https://example.com"
		reqBody := CreatePocketRequest{
			Title:       "Test Title",
			URL:         &url,
			ContentType: "article",
		}
		
		mockRes := &PocketResponse{ID: "1", Title: "Test Title", ContentType: "article"}
		mockSvc.On("Create", mock.Anything, "test-user-id", reqBody).Return(mockRes, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/pockets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		reqBody := CreatePocketRequest{Title: "T"} // Too short
		body, _ := json.Marshal(reqBody)
		
		req := httptest.NewRequest("POST", "/pockets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 422, resp.StatusCode)
	})
}

func TestPocketHandler_List(t *testing.T) {
	t.Run("Success List", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		expectedQuery := PocketListQuery{
			Search: "react",
			Status: "unread",
			Page:   1,
			Limit:  10,
		}
		
		mockRes := []PocketResponse{{ID: "1", Title: "Test"}}
		mockMeta := response.PaginationMeta{Page: 1, Limit: 10, Total: 1, TotalPage: 1}

		mockSvc.On("List", mock.Anything, "test-user-id", expectedQuery).Return(mockRes, mockMeta, nil).Once()

		req := httptest.NewRequest("GET", "/pockets?search=react&status=unread&page=1&limit=10", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestPocketHandler_GetDetail(t *testing.T) {
	t.Run("Success Get Detail", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		mockRes := &PocketResponse{ID: "1"}
		mockSvc.On("GetDetail", mock.Anything, "1", "test-user-id").Return(mockRes, nil).Once()

		req := httptest.NewRequest("GET", "/pockets/1", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		mockSvc.On("GetDetail", mock.Anything, "2", "test-user-id").Return(nil, apperror.NotFound("not found")).Once()

		req := httptest.NewRequest("GET", "/pockets/2", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestPocketHandler_Update(t *testing.T) {
	t.Run("Success Update", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		url := "https://example.com"
		reqBody := UpdatePocketRequest{
			Title:       "Test Title Updated",
			URL:         &url,
			ContentType: "article",
		}
		
		mockRes := &PocketResponse{ID: "1", Title: "Test Title Updated"}
		mockSvc.On("Update", mock.Anything, "1", "test-user-id", reqBody).Return(mockRes, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/pockets/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestPocketHandler_Archive(t *testing.T) {
	t.Run("Success Archive", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		mockRes := &PocketResponse{ID: "1"}
		mockSvc.On("Archive", mock.Anything, "1", "test-user-id").Return(mockRes, nil).Once()

		req := httptest.NewRequest("DELETE", "/pockets/1", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestPocketHandler_UpdateStatus(t *testing.T) {
	t.Run("Success Update Status", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		reqBody := UpdateStatusRequest{Status: "reading"}
		mockRes := &PocketResponse{ID: "1", Status: "reading"}
		
		mockSvc.On("UpdateStatus", mock.Anything, "1", "test-user-id", reqBody).Return(mockRes, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PATCH", "/pockets/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestPocketHandler_ToggleFavorite(t *testing.T) {
	t.Run("Success Toggle Favorite", func(t *testing.T) {
		mockSvc := new(MockPocketService)
		handler := NewPocketHandler(mockSvc)
		app := setupTestApp(handler)

		isFav := true
		reqBody := ToggleFavoriteRequest{IsFavorite: &isFav}
		mockRes := &PocketResponse{ID: "1", IsFavorite: true}
		
		mockSvc.On("ToggleFavorite", mock.Anything, "1", "test-user-id", reqBody).Return(mockRes, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PATCH", "/pockets/1/favorite", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}
