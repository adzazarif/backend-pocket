package router

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/domain/auth"
	"pocket-app/internal/domain/dashboard"
	"pocket-app/internal/domain/pocket"
	"pocket-app/internal/middleware"
)

func Setup(app *fiber.App, secret string, authHandler *auth.AuthHandler, dashHandler *dashboard.DashboardHandler, pocketHandler *pocket.PocketHandler) {
	// Setup API group
	api := app.Group("/api")

	// Apply logger middleware
	api.Use(middleware.Logger())

	// Auth routes (Public)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/logout", middleware.Auth(secret), authHandler.Logout)

	// Protected routes
	protected := api.Group("", middleware.Auth(secret))
	
	// Dashboard
	protected.Get("/dashboard", dashHandler.GetSummary)

	// Pocket
	pockets := protected.Group("/pockets")
	pockets.Get("", pocketHandler.List)
	pockets.Post("", pocketHandler.Create)
	pockets.Get("/:id", pocketHandler.GetDetail)
	pockets.Put("/:id", pocketHandler.Update)
	pockets.Delete("/:id", pocketHandler.Archive)
	pockets.Patch("/:id/status", pocketHandler.UpdateStatus)
	pockets.Patch("/:id/favorite", pocketHandler.ToggleFavorite)
}
