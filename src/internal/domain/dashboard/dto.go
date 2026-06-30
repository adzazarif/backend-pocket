package dashboard

import "pocket-app/internal/domain/pocket"

type DashboardSummaryResponse struct {
	TotalItems    int                     `json:"totalItems"`
	UnreadItems   int                     `json:"unreadItems"`
	ReadingItems  int                     `json:"readingItems"`
	ReadItems     int                     `json:"readItems"`
	ArchivedItems int                     `json:"archivedItems"`
	FavoriteItems int                     `json:"favoriteItems"`
	RecentlyAdded []pocket.PocketResponse `json:"recentlyAdded"`
}
