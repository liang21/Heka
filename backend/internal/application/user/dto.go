package user

import (
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T099 | spec.md: §4.2 认证 DTO

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID    shared.ID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}
