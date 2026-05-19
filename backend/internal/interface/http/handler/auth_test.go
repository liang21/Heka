package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/user"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T126 | spec.md: §4.2 Auth Handler TDD RED

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, req user.LoginRequest) (*user.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*user.TokenResponse), args.Error(1)
}

func (m *mockAuthService) GetMe(ctx context.Context, userID shared.ID) (*user.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*user.TokenResponse, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*user.TokenResponse), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()
	svc := new(mockAuthService)
	h := NewAuthHandler(svc)

	svc.On("Login", mock.Anything, mock.Anything).Return(&user.TokenResponse{
		AccessToken: "test-token", ExpiresAt: 1234567890,
	}, nil)

	body := map[string]string{"email": "test@example.com", "password": "password"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_GetMe(t *testing.T) {
	t.Parallel()
	svc := new(mockAuthService)
	h := NewAuthHandler(svc)

	userID := shared.NewID()
	svc.On("GetMe", mock.Anything, userID).Return(&user.UserResponse{
		ID: userID, Name: "Test", Email: "test@example.com",
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetMe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
