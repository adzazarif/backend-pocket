package dashboard

import (
	"context"

	"pocket-app/internal/domain/pocket"
	"pocket-app/internal/pkg/apperror"
)

type DashboardService interface {
	GetSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error)
}

type dashboardService struct {
	dashRepo   DashboardRepository
	pocketRepo pocket.PocketRepository
}

func NewDashboardService(dashRepo DashboardRepository, pocketRepo pocket.PocketRepository) DashboardService {
	return &dashboardService{dashRepo: dashRepo, pocketRepo: pocketRepo}
}

func (s *dashboardService) GetSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error) {
	summary, err := s.dashRepo.GetSummary(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch dashboard summary")
	}

	query := pocket.PocketListQuery{
		Limit: 5,
		Page:  1,
		Sort:  "createdAt:desc",
	}
	recentItems, _, err := s.pocketRepo.List(ctx, userID, query)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch recent items")
	}

	var recentlyAdded []pocket.PocketResponse
	for _, item := range recentItems {
		recentlyAdded = append(recentlyAdded, pocket.PocketResponse{
			ID:          item.ID,
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Description,
			ContentType: item.ContentType,
			Status:      item.Status,
			IsFavorite:  item.IsFavorite,
			Tags:        item.Tags,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	if recentlyAdded == nil {
		recentlyAdded = []pocket.PocketResponse{}
	}

	summary.RecentlyAdded = recentlyAdded
	return summary, nil
}
