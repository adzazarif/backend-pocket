package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pocket-app/internal/config"
	"pocket-app/internal/model"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/hash"
)

func TestAuthService_Login(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:      "secret",
		JWTExpiryHours: 24,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := NewAuthService(mockRepo, cfg)

		hashedPassword, _ := hash.HashPassword("password123")
		user := &model.User{
			ID:       "1",
			Email:    "test@example.com",
			Password: hashedPassword,
			Name:     "Test User",
		}

		req := LoginRequest{Email: "test@example.com", Password: "password123"}
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil).Once()

		res, err := svc.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Token)
		assert.Equal(t, "1", res.User.ID)
		assert.Equal(t, "test@example.com", res.User.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := NewAuthService(mockRepo, cfg)

		req := LoginRequest{Email: "test@example.com", Password: "password123"}
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil).Once()

		res, err := svc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, res)
		
		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INVALID_CREDENTIAL", appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := NewAuthService(mockRepo, cfg)

		hashedPassword, _ := hash.HashPassword("password123")
		user := &model.User{
			ID:       "1",
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		req := LoginRequest{Email: "test@example.com", Password: "wrongpassword"}
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil).Once()

		res, err := svc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, res)

		var appErr *apperror.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INVALID_CREDENTIAL", appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := NewAuthService(mockRepo, cfg)

		req := LoginRequest{Email: "test@example.com", Password: "password123"}
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("db error")).Once()

		res, err := svc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})
}
