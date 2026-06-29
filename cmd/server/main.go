package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"pocket-app/internal/config"
	"pocket-app/internal/database"
	"pocket-app/internal/middleware"
	"pocket-app/internal/router"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Connect to database
	db := database.Connect(cfg)
	
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	// 3. Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// 4. Middleware
	app.Use(cors.New())

	// 5. Setup dependencies
	// TODO: Initialize repositories, services, and handlers

	// 6. Setup router
	router.Setup(app, cfg.JWTSecret)

	// 7. Start server
	log.Printf("Server is starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
