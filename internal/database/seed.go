package database

import (
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pocket-app/internal/config"
	"pocket-app/internal/model"
)

func Seed(db *gorm.DB, cfg *config.Config) error {
	var count int64
	db.Model(&model.User{}).Where("email = ?", cfg.SeedUserEmail).Count(&count)
	if count > 0 {
		log.Println("Database already seeded")
		return nil // already seeded
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(cfg.SeedUserPass), bcrypt.DefaultCost)
	user := model.User{
		ID:       uuid.New().String(),
		Name:     cfg.SeedUserName,
		Email:    cfg.SeedUserEmail,
		Password: string(hashedPassword),
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	url1 := "https://example.com/react-performance"
	desc1 := "A guide about React rendering optimization"
	url2 := "https://example.com/typescript-generics"
	desc2 := "Deep dive into TypeScript generics"
	desc3 := "Personal notes about scalable frontend architecture"

	items := []model.PocketItem{
		{
			ID: uuid.New().String(), UserID: user.ID,
			Title: "React Performance Guide", URL: &url1, Description: &desc1,
			ContentType: "article", Status: "unread", IsFavorite: true,
			Tags: model.StringArray{"frontend", "react"},
		},
		{
			ID: uuid.New().String(), UserID: user.ID,
			Title: "Understanding TypeScript Generics", URL: &url2, Description: &desc2,
			ContentType: "article", Status: "reading", IsFavorite: false,
			Tags: model.StringArray{"typescript", "frontend"},
		},
		{
			ID: uuid.New().String(), UserID: user.ID,
			Title: "Frontend System Design Notes", Description: &desc3,
			ContentType: "note", Status: "read", IsFavorite: true,
			Tags: model.StringArray{"architecture", "frontend"},
		},
	}

	if err := db.Create(&items).Error; err != nil {
		return err
	}

	log.Println("Database seeded successfully")
	return nil
}
