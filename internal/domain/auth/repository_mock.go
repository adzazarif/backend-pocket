package auth

import (
	"context"

	"github.com/stretchr/testify/mock"
	"pocket-app/internal/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	
	var user *model.User
	if args.Get(0) != nil {
		user = args.Get(0).(*model.User)
	}
	
	return user, args.Error(1)
}
