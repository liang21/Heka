// tasks.md: T043 | TDD RED phase - UserRepository implementation tests
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// createUserRepo creates a new UserRepository instance for testing
func createUserRepo(db *gorm.DB) *UserRepository {
	// This will fail to compile until UserRepository is implemented
	return &UserRepository{db: db}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	testUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, testUser)
	assert.NoError(t, err, "Create should succeed")

	// Verify user was actually created
	found, err := repo.FindByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, found.ID)
	assert.Equal(t, testUser.Email, found.Email)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	existingUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Existing User",
		Email:        "duplicate@example.com",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, existingUser)
	require.NoError(t, err)

	duplicateUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Duplicate User",
		Email:        "duplicate@example.com", // Same email
		PasswordHash: "hash2",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(ctx, duplicateUser)
	assert.Error(t, err, "Create with duplicate email should fail")
}

func TestFindByEmail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	testUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Email Search User",
		Email:        "findme@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	found, err := repo.FindByEmail(ctx, "findme@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testUser.ID, found.ID)
	assert.Equal(t, testUser.Email, found.Email)
}

func TestFindByEmail_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	found, err := repo.FindByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestFindByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	testUser := &user.User{
		ID:           shared.NewID(),
		Name:         "ID Search User",
		Email:        "idsearch@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testUser.ID, found.ID)
	assert.Equal(t, testUser.Name, found.Name)
	assert.Equal(t, testUser.Email, found.Email)
}

func TestFindByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, shared.NewID())
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	testUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Original Name",
		Email:        "update@example.com",
		PasswordHash: "original_hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Update user
	testUser.Name = "Updated Name"
	testUser.PasswordHash = "new_hash"
	testUser.UpdatedAt = time.Now()

	err = repo.Update(ctx, testUser)
	assert.NoError(t, err, "Update should succeed")

	// Verify changes
	updated, err := repo.FindByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "new_hash", updated.PasswordHash)
}

func TestUpdateUser_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	nonExistentUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Ghost User",
		Email:        "ghost@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Update(ctx, nonExistentUser)
	assert.Error(t, err, "Update non-existent user should fail")
}

func TestSoftDelete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	testUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Delete Me User",
		Email:        "delete@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Soft delete the user
	err = repo.SoftDelete(ctx, testUser.ID)
	assert.NoError(t, err, "SoftDelete should succeed")

	// Verify user is soft deleted (not found by normal queries)
	found, err := repo.FindByID(ctx, testUser.ID)
	assert.Error(t, err, "Soft deleted user should not be found")
	assert.Nil(t, found)

	foundByEmail, err := repo.FindByEmail(ctx, "delete@example.com")
	assert.Error(t, err, "Soft deleted user should not be found by email")
	assert.Nil(t, foundByEmail)
}

func TestSoftDelete_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	err := repo.SoftDelete(ctx, shared.NewID())
	assert.Error(t, err, "SoftDelete non-existent user should fail")
}

func TestUserRepository_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	// Test concurrent creates
	const numUsers = 10
	users := make([]*user.User, numUsers)

	for i := 0; i < numUsers; i++ {
		users[i] = &user.User{
			ID:           shared.NewID(),
			Name:         fmt.Sprintf("Concurrent User %d", i),
			Email:        fmt.Sprintf("concurrent%d@example.com", i),
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	}

	// Create all users
	for _, u := range users {
		err := repo.Create(ctx, u)
		assert.NoError(t, err)
	}

	// Verify all users can be found
	for _, u := range users {
		found, err := repo.FindByID(ctx, u.ID)
		assert.NoError(t, err)
		assert.Equal(t, u.ID, found.ID)
	}
}

func TestUserRepository_EmptyEmail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	invalidUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Invalid User",
		Email:        "", // Empty email
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, invalidUser)
	assert.Error(t, err, "Create with empty email should fail")
}

func TestUserRepository_InvalidEmailFormat(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := createUserRepo(db)
	ctx := context.Background()

	invalidUser := &user.User{
		ID:           shared.NewID(),
		Name:         "Invalid Email User",
		Email:        "not-an-email", // Invalid format
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, invalidUser)
	assert.Error(t, err, "Create with invalid email format should fail")
}
