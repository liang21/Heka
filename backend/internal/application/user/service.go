package user

import (
	"context"
	"time"

	"github.com/liang21/heka/internal/infrastructure/auth"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
)

// tasks.md: T101 | spec.md: §4.2 UserService Implementation

type Service struct {
	repo           user.UserRepository
	jwtSecret      string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func NewService(repo user.UserRepository, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration) *Service {
	return &Service{
		repo:           repo,
		jwtSecret:      jwtSecret,
		accessTokenTTL: accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == shared.ErrUserNotFound {
			return nil, shared.ErrAuthInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		return nil, shared.ErrAuthInvalidCredentials
	}

	accessToken, err := auth.GenerateToken(s.jwtSecret, u.ID, u.Email, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(s.jwtSecret, u.ID, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.accessTokenTTL).Unix(),
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID shared.ID) (*UserResponse, error) {
	// Note: Repository interface uses string for ID - conversion required
	u, err := s.repo.FindByID(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := auth.ParseToken(s.jwtSecret, refreshToken)
	if err != nil {
		return nil, shared.ErrAuthTokenExpired
	}

	u, err := s.repo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, err
	}

	accessToken, err := auth.GenerateToken(s.jwtSecret, u.ID, u.Email, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := auth.GenerateRefreshToken(s.jwtSecret, u.ID, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.accessTokenTTL).Unix(),
	}, nil
}

func (s *Service) GenerateRefreshToken(ctx context.Context, userID shared.ID) (string, error) {
	return auth.GenerateRefreshToken(s.jwtSecret, userID, s.refreshTokenTTL)
}
