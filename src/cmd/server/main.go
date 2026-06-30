package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"pocket-app/internal/config"
	"pocket-app/internal/database"
	"pocket-app/internal/domain/auth"
	"pocket-app/internal/domain/dashboard"
	"pocket-app/internal/domain/pocket"
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
	userRepo := auth.NewUserRepository(db)
	pocketRepo := pocket.NewPocketRepository(db)
	dashRepo := dashboard.NewDashboardRepository(db)

	authSvc := auth.NewAuthService(userRepo, cfg)
	pocketSvc := pocket.NewPocketService(pocketRepo)
	dashSvc := dashboard.NewDashboardService(dashRepo, pocketRepo)

	authHandler := auth.NewAuthHandler(authSvc)
	pocketHandler := pocket.NewPocketHandler(pocketSvc)
	dashHandler := dashboard.NewDashboardHandler(dashSvc)

	// 6. Setup router
	router.Setup(app, cfg.JWTSecret, authHandler, dashHandler, pocketHandler)

	// 7. Start server
	log.Printf("Server is starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
