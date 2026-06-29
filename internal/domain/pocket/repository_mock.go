package pocket

import (
	"context"

	"github.com/stretchr/testify/mock"
	"pocket-app/internal/model"
)

type MockPocketRepository struct {
	mock.Mock
}

func (m *MockPocketRepository) Create(ctx context.Context, item *model.PocketItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockPocketRepository) Update(ctx context.Context, item *model.PocketItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockPocketRepository) FindByIDAndUserID(ctx context.Context, id, userID string) (*model.PocketItem, error) {
	args := m.Called(ctx, id, userID)
	
	var item *model.PocketItem
	if args.Get(0) != nil {
		item = args.Get(0).(*model.PocketItem)
	}
	
	return item, args.Error(1)
}

func (m *MockPocketRepository) List(ctx context.Context, userID string, query PocketListQuery) ([]model.PocketItem, int64, error) {
	args := m.Called(ctx, userID, query)
	
	var items []model.PocketItem
	if args.Get(0) != nil {
		items = args.Get(0).([]model.PocketItem)
	}
	var total int64
	if args.Get(1) != nil {
		total = args.Get(1).(int64)
	}
	
	return items, total, args.Error(2)
}
