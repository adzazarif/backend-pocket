package pocket

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"pocket-app/internal/model"
)

type PocketRepository interface {
	Create(ctx context.Context, item *model.PocketItem) error
	Update(ctx context.Context, item *model.PocketItem) error
	FindByIDAndUserID(ctx context.Context, id, userID string) (*model.PocketItem, error)
	List(ctx context.Context, userID string, query PocketListQuery) ([]model.PocketItem, int64, error)
}

type pocketRepository struct {
	db *gorm.DB
}

func NewPocketRepository(db *gorm.DB) PocketRepository {
	return &pocketRepository{db: db}
}

func (r *pocketRepository) Create(ctx context.Context, item *model.PocketItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *pocketRepository) Update(ctx context.Context, item *model.PocketItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *pocketRepository) FindByIDAndUserID(ctx context.Context, id, userID string) (*model.PocketItem, error) {
	var item model.PocketItem
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil so service can throw AppError
		}
		return nil, err
	}
	return &item, nil
}

func (r *pocketRepository) List(ctx context.Context, userID string, query PocketListQuery) ([]model.PocketItem, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.PocketItem{}).Where("user_id = ?", userID)

	// Status filter
	if query.Status == "archived" {
		db = db.Where("status = ?", "archived")
	} else {
		db = db.Where("status != ?", "archived")
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
	}

	// Type filter
	if query.Type != "" {
		db = db.Where("content_type = ?", query.Type)
	}

	// Favorite filter
	if query.Favorite == "true" {
		db = db.Where("is_favorite = ?", true)
	} else if query.Favorite == "false" {
		db = db.Where("is_favorite = ?", false)
	}

	// Search
	if query.Search != "" {
		// Using FULLTEXT index or fallback to LIKE for JSON
		searchStr := "%" + query.Search + "%"
		db = db.Where("(MATCH(title, url, description) AGAINST(? IN BOOLEAN MODE) OR JSON_SEARCH(tags, 'one', ?) IS NOT NULL OR title LIKE ?)", query.Search, searchStr, searchStr)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	switch query.Sort {
	case "createdAt:asc":
		db = db.Order("created_at asc")
	case "title:asc":
		db = db.Order("title asc")
	case "title:desc":
		db = db.Order("title desc")
	case "updatedAt:desc":
		db = db.Order("updated_at desc")
	default: // createdAt:desc
		db = db.Order("created_at desc")
	}

	// Pagination
	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit
	db = db.Offset(offset).Limit(limit)

	var items []model.PocketItem
	if err := db.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
