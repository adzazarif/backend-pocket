package pocket

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/model"
	"pocket-app/internal/pkg/apperror"
)

func TestPocketService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		url := "https://example.com"
		req := CreatePocketRequest{
			Title:       "Test",
			URL:         &url,
			ContentType: "article",
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(item *model.PocketItem) bool {
			return item.Title == req.Title && item.UserID == "user1" && item.Status == "unread"
		})).Return(nil).Once()

		res, err := svc.Create(context.Background(), "user1", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Test", res.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		url := "https://example.com"
		req := CreatePocketRequest{
			Title:       "Test",
			URL:         &url,
			ContentType: "article",
		}

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

		res, err := svc.Create(context.Background(), "user1", req)

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		query := PocketListQuery{Page: 1, Limit: 10}
		items := []model.PocketItem{{ID: "1", Title: "Test"}}
		
		mockRepo.On("List", mock.Anything, "user1", query).Return(items, int64(1), nil).Once()

		res, meta, err := svc.List(context.Background(), "user1", query)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 1, meta.Total)
		assert.Equal(t, 1, meta.TotalPage)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_GetDetail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		item := &model.PocketItem{ID: "1", Title: "Test"}
		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(item, nil).Once()

		res, err := svc.GetDetail(context.Background(), "1", "user1")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(nil, nil).Once()

		res, err := svc.GetDetail(context.Background(), "1", "user1")

		assert.Error(t, err)
		assert.Nil(t, res)
		
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, "POCKET_NOT_FOUND", appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		item := &model.PocketItem{ID: "1", Title: "Old Title"}
		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(item, nil).Once()
		
		url := "https://example.com"
		req := UpdatePocketRequest{Title: "New Title", URL: &url}
		
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(i *model.PocketItem) bool {
			return i.Title == "New Title"
		})).Return(nil).Once()

		res, err := svc.Update(context.Background(), "1", "user1", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "New Title", res.Title)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_Archive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		item := &model.PocketItem{ID: "1", Status: "unread"}
		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(item, nil).Once()
		
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(i *model.PocketItem) bool {
			return i.Status == "archived"
		})).Return(nil).Once()

		res, err := svc.Archive(context.Background(), "1", "user1")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "archived", res.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_UpdateStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		item := &model.PocketItem{ID: "1", Status: "unread"}
		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(item, nil).Once()
		
		req := UpdateStatusRequest{Status: "reading"}
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(i *model.PocketItem) bool {
			return i.Status == "reading"
		})).Return(nil).Once()

		res, err := svc.UpdateStatus(context.Background(), "1", "user1", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "reading", res.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestPocketService_ToggleFavorite(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPocketRepository)
		svc := NewPocketService(mockRepo)

		item := &model.PocketItem{ID: "1", IsFavorite: false}
		mockRepo.On("FindByIDAndUserID", mock.Anything, "1", "user1").Return(item, nil).Once()
		
		isFav := true
		req := ToggleFavoriteRequest{IsFavorite: &isFav}
		
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(i *model.PocketItem) bool {
			return i.IsFavorite == true
		})).Return(nil).Once()

		res, err := svc.ToggleFavorite(context.Background(), "1", "user1", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, res.IsFavorite)
		mockRepo.AssertExpectations(t)
	})
}
