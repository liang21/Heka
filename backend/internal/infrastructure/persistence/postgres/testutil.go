package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection using testcontainers
// This is a shared utility function for all repository tests in the postgres package
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// TODO: Implement testcontainers-go PostgreSQL setup
	// This requires:
	// 1. Adding testcontainers-go to go.mod: go get github.com/testcontainers/testcontainers-go
	// 2. Setting up PostgreSQL container with proper configuration
	// 3. Running database migrations to create schema
	// 4. Returning configured *gorm.DB instance

	// For now, return nil to make tests RED (fail) as per TDD principles
	assert.Fail(t, "setupTestDB: testcontainers PostgreSQL not implemented - TDD RED")
	return nil
}

// setupTestDBWithCtx creates a test database connection with context
// This is a convenience function for tests that need both context and database
func setupTestDBWithCtx(t *testing.T) (context.Context, *gorm.DB) {
	t.Helper()
	ctx := context.Background()
	db := setupTestDB(t)
	return ctx, db
}

// teardownTestDB cleans up the test database connection
// This is a shared utility function for all repository tests in the postgres package
func teardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	// TODO: Implement testcontainers cleanup
	// This should:
	// 1. Close the database connection
	// 2. Stop and remove the testcontainers PostgreSQL container
	// 3. Clean up any temporary resources

	// For now, this is a no-op since setupTestDB returns nil
	if db == nil {
		return
	}

	// Close the database connection if it exists
	sqlDB, err := db.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

// setupTestDBLegacy is the original setupTestDB function that takes context
// Kept for backward compatibility with existing tests
func setupTestDBLegacy(ctx context.Context) (*gorm.DB, error) {
	// TODO: Implement testcontainers-go PostgreSQL setup
	// This requires:
	// 1. Adding testcontainers-go to go.mod: go get github.com/testcontainers/testcontainers-go
	// 2. Setting up PostgreSQL container with proper configuration
	// 3. Running database migrations to create schema
	// 4. Returning configured *gorm.DB instance

	// For now, return error to make tests RED (fail) as per TDD principles
	return nil, assert.AnError
}
