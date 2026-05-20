package handler

import (
	"context"
	"encoding/json"
	"net/http"

	infraauth "github.com/liang21/heka/internal/infrastructure/auth"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T127 | spec.md: §4.2 Auth Handler Implementation

type AuthService interface {
	Login(ctx context.Context, req infraauth.LoginRequest) (*infraauth.TokenResponse, error)
	Register(ctx context.Context, req infraauth.RegisterRequest) (*infraauth.TokenResponse, error)
	GetMe(ctx context.Context, userID string) (*infraauth.UserDTO, error)
	RefreshToken(ctx context.Context, token string) (*infraauth.TokenResponse, error)
	Logout(ctx context.Context, token string) error
}

type AuthHandler struct {
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req infraauth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	token, err := h.svc.Login(r.Context(), req)
	if err != nil {
		response.Error(w, shared.ErrAuthInvalidCredentials)
		return
	}

	response.Success(w, token)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	userResp, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		response.Error(w, shared.ErrUserNotFound)
		return
	}

	response.Success(w, userResp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	token, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, shared.ErrAuthTokenExpired)
		return
	}

	response.Success(w, token)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		response.Error(w, shared.ErrSysInternal)
		return
	}

	response.Success(w, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req infraauth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	token, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if err.Error() == "user already exists" {
			response.Error(w, shared.NewAppError("AUTH-CF-001", "user already exists", http.StatusConflict))
			return
		}
		response.Error(w, shared.ErrSysInternal)
		return
	}

	response.Success(w, token)
}
