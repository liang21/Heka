package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
)

// tasks.md: T100 | spec.md: §4.2 UserService TDD RED

// --- Mock Repository ---

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

// --- Tests ---

func TestUserService_Login_Success(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	hashedPassword := "$2a$12$hashedpassword"
	u := &user.User{
		ID:           shared.NewID(),
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
	}

	repo.On("FindByEmail", mock.Anything, "test@example.com").Return(u, nil)

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.True(t, resp.ExpiresAt > 0)
	repo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	u := &user.User{
		ID:           shared.NewID(),
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashedpassword",
	}

	repo.On("FindByEmail", mock.Anything, "test@example.com").Return(u, nil)

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	repo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, shared.ErrUserNotFound)

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	// AC: 登录失败不区分用户不存在和密码错误
	assert.Equal(t, shared.ErrAuthInvalidCredentials.HTTPStatus, err.(*shared.AppError).HTTPStatus)
}

func TestUserService_GetMe(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	userID := shared.NewID()
	u := &user.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	repo.On("FindByID", mock.Anything, userID.String()).Return(u, nil)

	resp, err := svc.GetMe(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, userID, resp.ID)
	assert.Equal(t, "Test User", resp.Name)
	assert.Equal(t, "test@example.com", resp.Email)
	repo.AssertExpectations(t)
}

func TestUserService_GetMe_NotFound(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	userID := shared.NewID()
	repo.On("FindByID", mock.Anything, userID.String()).Return(nil, shared.ErrUserNotFound)

	resp, err := svc.GetMe(context.Background(), userID)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestUserService_RefreshToken_Success(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	userID := shared.NewID()
	u := &user.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	repo.On("FindByID", mock.Anything, userID.String()).Return(u, nil)

	// Generate a valid refresh token first
	refreshToken, err := svc.GenerateRefreshToken(context.Background(), userID)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(context.Background(), refreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	repo.AssertExpectations(t)
}

func TestUserService_RefreshToken_Expired(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	// Very short TTL so token expires immediately
	svc := NewService(repo, "test-secret", 1*time.Nanosecond, 1*time.Nanosecond)

	resp, err := svc.RefreshToken(context.Background(), "invalid-token")

	assert.Nil(t, resp)
	assert.Error(t, err)
}

// Additional edge case tests

func TestUserService_Login_RepositoryError(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, shared.ErrSysInternal)

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, shared.ErrSysInternal, err)
}

func TestUserService_RefreshToken_MalformedToken(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	resp, err := svc.RefreshToken(context.Background(), "not-a-valid-jwt-token")

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestUserService_GetMe_UserNotFound(t *testing.T) {
	t.Parallel()
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret", 24*time.Hour, 7*24*time.Hour)

	userID := shared.NewID()
	repo.On("FindByID", mock.Anything, userID.String()).Return(nil, shared.ErrUserNotFound)

	resp, err := svc.GetMe(context.Background(), userID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, shared.ErrUserNotFound, err)
}
