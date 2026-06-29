package pocket

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"pocket-app/internal/model"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/response"
)

type PocketService interface {
	Create(ctx context.Context, userID string, req CreatePocketRequest) (*PocketResponse, error)
	Update(ctx context.Context, id, userID string, req UpdatePocketRequest) (*PocketResponse, error)
	GetDetail(ctx context.Context, id, userID string) (*PocketResponse, error)
	Archive(ctx context.Context, id, userID string) (*PocketResponse, error) // Returns minimal response
	UpdateStatus(ctx context.Context, id, userID string, req UpdateStatusRequest) (*PocketResponse, error)
	ToggleFavorite(ctx context.Context, id, userID string, req ToggleFavoriteRequest) (*PocketResponse, error)
	List(ctx context.Context, userID string, query PocketListQuery) ([]PocketResponse, response.PaginationMeta, error)
}

type pocketService struct {
	repo PocketRepository
}

func NewPocketService(repo PocketRepository) PocketService {
	return &pocketService{repo: repo}
}

func (s *pocketService) Create(ctx context.Context, userID string, req CreatePocketRequest) (*PocketResponse, error) {
	if req.ContentType != "note" && (req.URL == nil || *req.URL == "") {
		return nil, apperror.ValidationError([]apperror.FieldError{
			{Field: "url", Message: "URL is required"},
		})
	}

	req.Tags = deduplicateTags(req.Tags)

	item := &model.PocketItem{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       req.Title,
		URL:         req.URL,
		Description: req.Description,
		ContentType: req.ContentType,
		Status:      "unread",
		IsFavorite:  false,
		Tags:        req.Tags,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperror.Internal("Failed to create pocket item")
	}

	return toPocketResponse(item), nil
}

func (s *pocketService) Update(ctx context.Context, id, userID string, req UpdatePocketRequest) (*PocketResponse, error) {
	if req.ContentType != "note" && (req.URL == nil || *req.URL == "") {
		return nil, apperror.ValidationError([]apperror.FieldError{
			{Field: "url", Message: "URL is required"},
		})
	}

	item, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch pocket item")
	}
	if item == nil {
		return nil, apperror.NotFound("Pocket item not found")
	}

	req.Tags = deduplicateTags(req.Tags)

	item.Title = req.Title
	item.URL = req.URL
	item.Description = req.Description
	item.ContentType = req.ContentType
	item.Tags = req.Tags
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal("Failed to update pocket item")
	}

	return toPocketResponse(item), nil
}

func (s *pocketService) GetDetail(ctx context.Context, id, userID string) (*PocketResponse, error) {
	item, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch pocket item")
	}
	if item == nil {
		return nil, apperror.NotFound("Pocket item not found")
	}

	return toPocketResponse(item), nil
}

func (s *pocketService) Archive(ctx context.Context, id, userID string) (*PocketResponse, error) {
	item, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch pocket item")
	}
	if item == nil {
		return nil, apperror.NotFound("Pocket item not found")
	}

	item.Status = "archived"
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal("Failed to archive pocket item")
	}

	return toPocketResponse(item), nil
}

func (s *pocketService) UpdateStatus(ctx context.Context, id, userID string, req UpdateStatusRequest) (*PocketResponse, error) {
	item, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch pocket item")
	}
	if item == nil {
		return nil, apperror.NotFound("Pocket item not found")
	}

	item.Status = req.Status
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal("Failed to update status")
	}

	res := toPocketResponse(item)
	return res, nil
}

func (s *pocketService) ToggleFavorite(ctx context.Context, id, userID string, req ToggleFavoriteRequest) (*PocketResponse, error) {
	item, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch pocket item")
	}
	if item == nil {
		return nil, apperror.NotFound("Pocket item not found")
	}

	item.IsFavorite = *req.IsFavorite
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal("Failed to update favorite")
	}

	res := toPocketResponse(item)
	return res, nil
}

func (s *pocketService) List(ctx context.Context, userID string, query PocketListQuery) ([]PocketResponse, response.PaginationMeta, error) {
	items, total, err := s.repo.List(ctx, userID, query)
	if err != nil {
		return nil, response.PaginationMeta{}, apperror.Internal("Failed to fetch pocket items")
	}

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

	totalPage := 0
	if total > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(limit)))
	}

	meta := response.PaginationMeta{
		Page:      page,
		Limit:     limit,
		Total:     int(total),
		TotalPage: totalPage,
	}

	responses := make([]PocketResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, *toPocketResponse(&item))
	}

	return responses, meta, nil
}

func toPocketResponse(item *model.PocketItem) *PocketResponse {
	return &PocketResponse{
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
	}
}

func deduplicateTags(tags []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, tag := range tags {
		if _, ok := seen[tag]; !ok && tag != "" {
			seen[tag] = struct{}{}
			result = append(result, tag)
		}
	}
	return result
}
