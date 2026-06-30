package auth

import (
	"context"

	"pocket-app/internal/config"
	"pocket-app/internal/pkg/apperror"
	"pocket-app/internal/pkg/hash"
	"pocket-app/internal/pkg/jwt"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type authService struct {
	userRepo UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch user")
	}

	if user == nil {
		return nil, apperror.InvalidCredential()
	}

	if !hash.ComparePassword(req.Password, user.Password) {
		return nil, apperror.InvalidCredential()
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.Name, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		},
	}, nil
}
