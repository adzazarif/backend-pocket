package pocket

import (
	"context"

	"github.com/stretchr/testify/mock"
	"pocket-app/internal/pkg/response"
)

type MockPocketService struct {
	mock.Mock
}

func (m *MockPocketService) Create(ctx context.Context, userID string, req CreatePocketRequest) (*PocketResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) Update(ctx context.Context, id, userID string, req UpdatePocketRequest) (*PocketResponse, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) GetDetail(ctx context.Context, id, userID string) (*PocketResponse, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) Archive(ctx context.Context, id, userID string) (*PocketResponse, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) UpdateStatus(ctx context.Context, id, userID string, req UpdateStatusRequest) (*PocketResponse, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) ToggleFavorite(ctx context.Context, id, userID string, req ToggleFavoriteRequest) (*PocketResponse, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PocketResponse), args.Error(1)
}

func (m *MockPocketService) List(ctx context.Context, userID string, query PocketListQuery) ([]PocketResponse, response.PaginationMeta, error) {
	args := m.Called(ctx, userID, query)
	return args.Get(0).([]PocketResponse), args.Get(1).(response.PaginationMeta), args.Error(2)
}
