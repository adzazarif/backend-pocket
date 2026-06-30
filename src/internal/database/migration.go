package database

import (
	"log"

	"gorm.io/gorm"
	"pocket-app/internal/model"
)

func Migrate(db *gorm.DB) error {
	// GORM AutoMigrate
	err := db.AutoMigrate(
		&model.User{},
		&model.PocketItem{},
	)
	if err != nil {
		return err
	}

	// Create composite indexes and fulltext index manually
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_pocket_items_status ON pocket_items (user_id, status);",
		"CREATE INDEX IF NOT EXISTS idx_pocket_items_content_type ON pocket_items (user_id, content_type);",
		"CREATE INDEX IF NOT EXISTS idx_pocket_items_is_favorite ON pocket_items (user_id, is_favorite);",
		"CREATE INDEX IF NOT EXISTS idx_pocket_items_created_at ON pocket_items (user_id, created_at DESC);",
	}

	for _, query := range indexes {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
			// Don't return error to allow continuing
		}
	}

	// FULLTEXT index
	fulltextQuery := "ALTER TABLE pocket_items ADD FULLTEXT INDEX ft_pocket_items_search (title, url, description);"
	if err := db.Exec(fulltextQuery).Error; err != nil {
		// Usually if it already exists it will return an error, we log it
		log.Printf("Warning (or info): failed to create fulltext index (might already exist): %v", err)
	}

	log.Println("Database migration completed")
	return nil
}
