package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/middleware"
)

func setupTestApp(handler *AuthHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	app.Post("/login", handler.Login)
	app.Post("/logout", handler.Logout)
	return app
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("Success Login", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := NewAuthHandler(mockSvc)
		app := setupTestApp(handler)

		reqBody := LoginRequest{Email: "test@example.com", Password: "password123"}
		mockRes := &LoginResponse{Token: "token123", User: UserResponse{ID: "1", Email: reqBody.Email}}

		mockSvc.On("Login", mock.Anything, reqBody).Return(mockRes, nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation Error - Invalid Email", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := NewAuthHandler(mockSvc)
		app := setupTestApp(handler)

		reqBody := map[string]string{"email": "not-email", "password": "password123"}
		body, _ := json.Marshal(reqBody)
		
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 422, resp.StatusCode)
	})

	t.Run("Invalid Credential", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := NewAuthHandler(mockSvc)
		app := setupTestApp(handler)

		reqBody := LoginRequest{Email: "test@example.com", Password: "wrongpassword"}

		mockSvc.On("Login", mock.Anything, reqBody).Return(nil, apperror.InvalidCredential()).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("Success Logout", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := NewAuthHandler(mockSvc)
		app := setupTestApp(handler)

		req := httptest.NewRequest("POST", "/logout", nil)
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
