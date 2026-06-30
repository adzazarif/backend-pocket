package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/domain/pocket"
	"pocket-app/internal/model"
)

func TestDashboardService_GetSummary(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		mockPocketRepo := new(pocket.MockPocketRepository)
		svc := NewDashboardService(mockRepo, mockPocketRepo)

		expectedRes := &DashboardSummaryResponse{
			TotalItems:    10,
			UnreadItems:   5,
			ReadingItems:  2,
			ReadItems:     3,
			ArchivedItems: 1,
			FavoriteItems: 4,
		}

		mockRepo.On("GetSummary", mock.Anything, "user1").Return(expectedRes, nil).Once()

		query := pocket.PocketListQuery{Limit: 5, Page: 1, Sort: "createdAt:desc"}
		mockPocketRepo.On("List", mock.Anything, "user1", query).Return([]model.PocketItem{}, int64(0), nil).Once()

		res, err := svc.GetSummary(context.Background(), "user1")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 10, res.TotalItems)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		mockPocketRepo := new(pocket.MockPocketRepository)
		svc := NewDashboardService(mockRepo, mockPocketRepo)

		mockRepo.On("GetSummary", mock.Anything, "user1").Return(nil, errors.New("db error")).Once()

		res, err := svc.GetSummary(context.Background(), "user1")

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})
}
