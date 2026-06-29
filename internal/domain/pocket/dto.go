package pocket

import "time"

type CreatePocketRequest struct {
	Title       string   `json:"title" validate:"required,min=3,max=120"`
	URL         *string  `json:"url"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	ContentType string   `json:"contentType" validate:"required,oneof=article video document note"`
	Tags        []string `json:"tags" validate:"omitempty,max=10,dive,max=24"`
}

type UpdatePocketRequest struct {
	Title       string   `json:"title" validate:"required,min=3,max=120"`
	URL         *string  `json:"url"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	ContentType string   `json:"contentType" validate:"required,oneof=article video document note"`
	Tags        []string `json:"tags" validate:"omitempty,max=10,dive,max=24"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=unread reading read"`
}

type ToggleFavoriteRequest struct {
	IsFavorite *bool `json:"isFavorite" validate:"required"`
}

type PocketResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         *string   `json:"url"`
	Description *string   `json:"description"`
	ContentType string    `json:"contentType"`
	Status      string    `json:"status"`
	IsFavorite  bool      `json:"isFavorite"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PocketListQuery struct {
	Search   string `query:"search"`
	Status   string `query:"status"`
	Type     string `query:"type"`
	Favorite string `query:"favorite"` // bool as string to handle empty easily
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Sort     string `query:"sort"`
}
