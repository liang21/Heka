package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid credentials")
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

type TokenMaker interface {
	GenerateAccessToken(userID shared.ID, email string) (string, error)
	GenerateRefreshToken(userID shared.ID) (string, error)
	ParseAccessToken(token string) (*Claims, error)
	ParseRefreshToken(token string) (*Claims, error)
}

type JWTMaker struct {
	secret          string
	accessTokenTTL  TokenTTL
	refreshTokenTTL TokenTTL
}

type TokenTTL struct {
	Duration int
}

func NewJWTMaker(secret string, accessTokenTTL, refreshTokenTTL TokenTTL) *JWTMaker {
	return &JWTMaker{
		secret:          secret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (j *JWTMaker) GenerateAccessToken(userID shared.ID, email string) (string, error) {
	return GenerateToken(j.secret, userID, email, time.Duration(j.accessTokenTTL.Duration)*time.Second)
}

func (j *JWTMaker) GenerateRefreshToken(userID shared.ID) (string, error) {
	return GenerateRefreshToken(j.secret, userID, time.Duration(j.refreshTokenTTL.Duration)*time.Second)
}

func (j *JWTMaker) ParseAccessToken(token string) (*Claims, error) {
	return ParseToken(j.secret, token)
}

func (j *JWTMaker) ParseRefreshToken(token string) (*Claims, error) {
	return ParseToken(j.secret, token)
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id string) (*user.User, error)
	Create(ctx context.Context, u *user.User) error
}

type Service struct {
	userRepo   UserRepository
	hasher     PasswordHasher
	tokenMaker TokenMaker
}

func NewService(userRepo UserRepository, hasher PasswordHasher, tokenMaker TokenMaker) *Service {
	return &Service{
		userRepo:   userRepo,
		hasher:     hasher,
		tokenMaker: tokenMaker,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}

type UserDTO struct {
	ID        shared.ID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserDTO(u *user.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	u, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	if !s.hasher.CheckPassword(req.Password, u.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	accessToken, err := s.tokenMaker.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenMaker.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserDTO(u),
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error) {
	// Check if user already exists
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	newUser := &user.User{
		ID:           shared.NewID(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.tokenMaker.GenerateAccessToken(newUser.ID, newUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenMaker.GenerateRefreshToken(newUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserDTO(newUser),
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*UserDTO, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	dto := toUserDTO(u)
	return &dto, nil
}

func (s *Service) RefreshToken(ctx context.Context, token string) (*TokenResponse, error) {
	claims, err := s.tokenMaker.ParseRefreshToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	u, err := s.userRepo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	accessToken, err := s.tokenMaker.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenMaker.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserDTO(u),
	}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return nil
}
