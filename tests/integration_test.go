package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"pocket-app/internal/config"
	"pocket-app/internal/domain/auth"
	"pocket-app/internal/domain/dashboard"
	"pocket-app/internal/domain/pocket"
	"pocket-app/internal/middleware"
	"pocket-app/internal/model"
	"pocket-app/internal/pkg/hash"
	"pocket-app/internal/router"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to in-memory database")
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.PocketItem{})
	if err != nil {
		panic("Failed to migrate database")
	}

	return db
}

func setupIntegrationApp(db *gorm.DB) *fiber.App {
	cfg := &config.Config{
		JWTSecret:      "integration-secret",
		JWTExpiryHours: 24,
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Init repositories
	userRepo := auth.NewUserRepository(db)
	pocketRepo := pocket.NewPocketRepository(db)
	dashboardRepo := dashboard.NewDashboardRepository(db)

	// Init services
	authService := auth.NewAuthService(userRepo, cfg)
	pocketService := pocket.NewPocketService(pocketRepo)
	dashboardService := dashboard.NewDashboardService(dashboardRepo, pocketRepo)

	// Init handlers
	authHandler := auth.NewAuthHandler(authService)
	pocketHandler := pocket.NewPocketHandler(pocketService)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardService)

	// Setup routes
	router.Setup(app, cfg.JWTSecret, authHandler, dashboardHandler, pocketHandler)

	return app
}

func TestAPIIntegration(t *testing.T) {
	db := setupTestDB()
	app := setupIntegrationApp(db)

	// 1. Seed User
	hashedPassword, _ := hash.HashPassword("password123")
	user := model.User{
		ID:       "test-user-1",
		Name:     "Test Integration User",
		Email:    "integration@example.com",
		Password: hashedPassword,
	}
	db.Create(&user)

	var token string

	t.Run("1. Login Success", func(t *testing.T) {
		reqBody := auth.LoginRequest{
			Email:    "integration@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var resBody struct {
			Data auth.LoginResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&resBody)
		token = resBody.Data.Token
		assert.NotEmpty(t, token)
	})

	t.Run("2. Create Pocket Item Success", func(t *testing.T) {
		url := "https://example.com/integration"
		reqBody := pocket.CreatePocketRequest{
			Title:       "Integration Test Item",
			URL:         &url,
			ContentType: "article",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/pockets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("3. Validation Error (Create Pocket without Title)", func(t *testing.T) {
		reqBody := pocket.CreatePocketRequest{
			ContentType: "article",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/pockets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode) // 422
	})

	t.Run("4. Check Dashboard Summary", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var resBody struct {
			Data dashboard.DashboardSummaryResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&resBody)
		
		assert.Equal(t, 1, resBody.Data.TotalItems)
		assert.Equal(t, 1, resBody.Data.UnreadItems) // Default status is unread
	})

	t.Run("5. Error Handling - Unauthorized (No Token)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/dashboard", nil)
		
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) // 401
	})
}
