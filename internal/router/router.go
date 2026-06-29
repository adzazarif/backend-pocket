package router

import (
	"github.com/gofiber/fiber/v2"
	"pocket-app/internal/middleware"
)

func Setup(app *fiber.App, secret string) {
	// Setup API group
	api := app.Group("/api")

	// Apply logger middleware
	api.Use(middleware.Logger())

	// TODO: Add Auth routes (Public)
	// authGroup := api.Group("/auth")
	// authGroup.Post("/login", authHandler.Login)
	// authGroup.Post("/logout", authHandler.Logout)

	// TODO: Add Protected routes
	// protected := api.Group("", middleware.Auth(secret))
	
	// TODO: Dashboard
	// protected.Get("/dashboard", dashHandler.GetSummary)

	// TODO: Pocket
	// pockets := protected.Group("/pockets")
	// pockets.Get("", pocketHandler.List)
	// pockets.Post("", pocketHandler.Create)
	// pockets.Get("/:id", pocketHandler.GetDetail)
	// pockets.Put("/:id", pocketHandler.Update)
	// pockets.Delete("/:id", pocketHandler.Archive)
	// pockets.Patch("/:id/status", pocketHandler.UpdateStatus)
	// pockets.Patch("/:id/favorite", pocketHandler.ToggleFavorite)
}
