// tasks.md: T052 | TDD GREEN integration test for TestCaseRepository
// This is a minimal integration test to verify the implementation works
package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
)

// TestTestCaseRepositoryIntegration tests the repository implementation
func TestTestCaseRepositoryIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("heka_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	}()

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Initialize GORM DB
	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&TestCaseModel{}, &StepModel{})
	require.NoError(t, err)

	// Initialize repository
	repo := NewTestCaseRepository(db)

	// Test data
	projectID := shared.NewID()
	moduleID := shared.NewID()
	userID := shared.NewID()

	t.Run("Create", func(t *testing.T) {
		tc := &testcase.TestCase{
			ID:          shared.NewID(),
			ProjectID:   projectID,
			ModuleID:    &moduleID,
			Title:       "Integration Test Case",
			Description: "Test description",
			Status:      testcase.CaseDraft,
			Priority:    testcase.P1,
			Tags:        []string{"tag1", "tag2"},
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Steps: []testcase.Step{
				{
					ID:        shared.NewID(),
					Number:    1,
					Action:    "Test action",
					Expected:  "Test expected",
				},
			},
		}

		err := repo.Create(ctx, tc)
		require.NoError(t, err)
		assert.NotEmpty(t, tc.ID)
		assert.Equal(t, 1, tc.Version)
	})

	t.Run("FindByID", func(t *testing.T) {
		// First create a test case
		tc := &testcase.TestCase{
			ID:          shared.NewID(),
			ProjectID:   projectID,
			ModuleID:    &moduleID,
			Title:       "Find By ID Test",
			Description: "Test description",
			Status:      testcase.CaseDraft,
			Priority:    testcase.P1,
			Tags:        []string{"tag1"},
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Steps: []testcase.Step{
				{
					ID:        shared.NewID(),
					Number:    1,
					Action:    "Step 1",
					Expected:  "Expected 1",
				},
				{
					ID:        shared.NewID(),
					Number:    2,
					Action:    "Step 2",
					Expected:  "Expected 2",
				},
			},
		}

		err := repo.Create(ctx, tc)
		require.NoError(t, err)

		// Find by ID
		foundTC, err := repo.FindByID(ctx, tc.ID)
		require.NoError(t, err)
		assert.Equal(t, tc.ID, foundTC.ID)
		assert.Equal(t, tc.Title, foundTC.Title)
		assert.Len(t, foundTC.Steps, 2)
		assert.Equal(t, "Step 1", foundTC.Steps[0].Action)
	})

	t.Run("List", func(t *testing.T) {
		// Create multiple test cases
		for i := 1; i <= 5; i++ {
			tc := &testcase.TestCase{
				ID:          shared.NewID(),
				ProjectID:   projectID,
				ModuleID:    &moduleID,
				Title:       "List Test Case",
				Description: "Test description",
				Status:      testcase.CaseDraft,
				Priority:    testcase.P1,
				Tags:        []string{"list"},
				CreatedBy:   userID,
				Version:     1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			err := repo.Create(ctx, tc)
			require.NoError(t, err)
		}

		// List test cases
		filter := testcase.TestCaseFilter{
			ProjectID: projectID,
			Page:      1,
			PageSize:  10,
		}

		cases, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 5)
		assert.GreaterOrEqual(t, total, int64(5))
	})

	t.Run("Update", func(t *testing.T) {
		// Create a test case
		tc := &testcase.TestCase{
			ID:          shared.NewID(),
			ProjectID:   projectID,
			ModuleID:    &moduleID,
			Title:       "Update Test",
			Description: "Original description",
			Status:      testcase.CaseDraft,
			Priority:    testcase.P1,
			Tags:        []string{"update"},
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, tc)
		require.NoError(t, err)

		// Update the test case
		tc.Title = "Updated Test"
		tc.Description = "Updated description"
		err = repo.Update(ctx, tc)
		require.NoError(t, err)
		assert.Equal(t, 2, tc.Version)

		// Verify update
		foundTC, err := repo.FindByID(ctx, tc.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test", foundTC.Title)
		assert.Equal(t, "Updated description", foundTC.Description)
		assert.Equal(t, 2, foundTC.Version)
	})

	t.Run("OptimisticLocking", func(t *testing.T) {
		// Create a test case
		tc := &testcase.TestCase{
			ID:          shared.NewID(),
			ProjectID:   projectID,
			ModuleID:    &moduleID,
			Title:       "Optimistic Lock Test",
			Description: "Test",
			Status:      testcase.CaseDraft,
			Priority:    testcase.P1,
			Tags:        []string{"lock"},
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, tc)
		require.NoError(t, err)

		// Fetch two copies
		tc1, err := repo.FindByID(ctx, tc.ID)
		require.NoError(t, err)

		tc2, err := repo.FindByID(ctx, tc.ID)
		require.NoError(t, err)

		// First update succeeds
		tc1.Title = "Updated by User 1"
		err = repo.Update(ctx, tc1)
		require.NoError(t, err)
		assert.Equal(t, 2, tc1.Version)

		// Second update with stale version fails
		tc2.Title = "Updated by User 2"
		err = repo.Update(ctx, tc2)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrTestCaseConflict, err)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		// Create a test case
		tc := &testcase.TestCase{
			ID:          shared.NewID(),
			ProjectID:   projectID,
			ModuleID:    &moduleID,
			Title:       "Soft Delete Test",
			Description: "Test",
			Status:      testcase.CaseDraft,
			Priority:    testcase.P1,
			Tags:        []string{"delete"},
			CreatedBy:   userID,
			Version:     1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, tc)
		require.NoError(t, err)

		// Soft delete
		err = repo.SoftDelete(ctx, tc.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = repo.FindByID(ctx, tc.ID)
		assert.Error(t, err)
		assert.Equal(t, shared.ErrTestCaseNotFound, err)
	})

	t.Run("BatchOperations", func(t *testing.T) {
		// Create test cases
		ids := []shared.ID{}
		for i := 1; i <= 3; i++ {
			tc := &testcase.TestCase{
				ID:          shared.NewID(),
				ProjectID:   projectID,
				ModuleID:    &moduleID,
				Title:       "Batch Test",
				Description: "Test",
				Status:      testcase.CaseDraft,
				Priority:    testcase.P1,
				Tags:        []string{"batch"},
				CreatedBy:   userID,
				Version:     1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			err := repo.Create(ctx, tc)
			require.NoError(t, err)
			ids = append(ids, tc.ID)
		}

		// Batch update status
		err := repo.BatchUpdateStatus(ctx, ids, testcase.CaseReady)
		require.NoError(t, err)

		// Verify status updates
		for _, id := range ids {
			tc, err := repo.FindByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, testcase.CaseReady, tc.Status)
		}

		// Batch move
		newModuleID := shared.NewID()
		err = repo.BatchMove(ctx, ids, &newModuleID)
		require.NoError(t, err)

		// Verify move
		for _, id := range ids {
			tc, err := repo.FindByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, &newModuleID, tc.ModuleID)
		}

		// Batch delete
		err = repo.BatchDelete(ctx, ids)
		require.NoError(t, err)

		// Verify deletion
		for _, id := range ids {
			_, err := repo.FindByID(ctx, id)
			assert.Error(t, err)
			assert.Equal(t, shared.ErrTestCaseNotFound, err)
		}
	})
}
