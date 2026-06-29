package dashboard

import (
	"context"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error) {
	query := `
		SELECT
			COUNT(*) AS total_items,
			COUNT(CASE WHEN status = 'unread' THEN 1 END) AS unread_items,
			COUNT(CASE WHEN status = 'reading' THEN 1 END) AS reading_items,
			COUNT(CASE WHEN status = 'read' THEN 1 END) AS read_items,
			COUNT(CASE WHEN status = 'archived' THEN 1 END) AS archived_items,
			COUNT(CASE WHEN is_favorite = 1 THEN 1 END) AS favorite_items
		FROM pocket_items
		WHERE user_id = ?;
	`

	var summary struct {
		TotalItems    int
		UnreadItems   int
		ReadingItems  int
		ReadItems     int
		ArchivedItems int
		FavoriteItems int
	}

	err := r.db.WithContext(ctx).Raw(query, userID).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	res := &DashboardSummaryResponse{
		TotalItems:    summary.TotalItems,
		UnreadItems:   summary.UnreadItems,
		ReadingItems:  summary.ReadingItems,
		ReadItems:     summary.ReadItems,
		ArchivedItems: summary.ArchivedItems,
		FavoriteItems: summary.FavoriteItems,
	}

	return res, nil
}
