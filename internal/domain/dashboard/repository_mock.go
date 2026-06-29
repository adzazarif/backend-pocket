package dashboard

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetSummary(ctx context.Context, userID string) (*DashboardSummaryResponse, error) {
	args := m.Called(ctx, userID)
	
	var res *DashboardSummaryResponse
	if args.Get(0) != nil {
		res = args.Get(0).(*DashboardSummaryResponse)
	}
	
	return res, args.Error(1)
}
