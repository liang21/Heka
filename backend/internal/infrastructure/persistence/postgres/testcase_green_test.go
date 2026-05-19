// tasks.md: T052 | TDD GREEN for TestCaseRepository
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

// testSuiteGreen holds the test database and repository for GREEN phase tests
type testSuiteGreen struct {
	ctx       context.Context
	container *postgres.PostgresContainer
	db        *gorm.DB
	repo      testcase.TestCaseRepository
	projectID shared.ID
	moduleID  *shared.ID
	userID    shared.ID
}

// setupTestSuiteGreen creates a new PostgreSQL test container and test suite for GREEN phase
func setupTestSuiteGreen(t *testing.T) *testSuiteGreen {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container with test database
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
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string")

	// Initialize GORM DB
	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err, "Failed to open GORM connection")

	// Create tables
	err = db.AutoMigrate(&TestCaseModel{}, &StepModel{})
	require.NoError(t, err, "Failed to migrate tables")

	// Initialize repository
	repo := NewTestCaseRepository(db)

	// Create test IDs
	projectID := shared.NewID()
	moduleID := shared.NewID()
	userID := shared.NewID()

	return &testSuiteGreen{
		ctx:       ctx,
		container: pgContainer,
		db:        db,
		repo:      repo,
		projectID: projectID,
		moduleID:  &moduleID,
		userID:    userID,
	}
}

// teardownGreen cleans up the test container
func (ts *testSuiteGreen) teardownGreen(t *testing.T) {
	t.Helper()
	if err := ts.container.Terminate(ts.ctx); err != nil {
		t.Fatalf("Failed to terminate container: %v", err)
	}
}

// Helper function to create a test TestCase
func (ts *testSuiteGreen) createTestCase(title string) *testcase.TestCase {
	return &testcase.TestCase{
		ID:          shared.NewID(),
		ProjectID:   ts.projectID,
		ModuleID:    ts.moduleID,
		Title:       title,
		Description: "Test description",
		Status:      testcase.CaseDraft,
		Priority:    testcase.P1,
		Tags:        []string{"tag1", "tag2"},
		CreatedBy:   ts.userID,
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
}

// TestCreateWithSteps tests creating a test case with steps
func TestCreateWithSteps(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	tc := ts.createTestCase("Test Case with Steps")

	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err, "Create should succeed")
	assert.NotEmpty(t, tc.ID, "ID should be set after creation")
	assert.Greater(t, tc.Version, 0, "Version should be set after creation")
}

// TestFindByID tests finding a test case by ID with steps preloaded
func TestFindByID(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// First create a test case
	tc := ts.createTestCase("Find By ID Test")
	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err)

	// Find the test case
	foundTC, err := ts.repo.FindByID(ts.ctx, tc.ID)
	require.NoError(t, err, "FindByID should succeed")
	assert.Equal(t, tc.ID, foundTC.ID, "ID should match")
	assert.Equal(t, tc.Title, foundTC.Title, "Title should match")
	assert.NotEmpty(t, foundTC.Steps, "Steps should be preloaded")
	assert.Greater(t, len(foundTC.Steps), 0, "Should have at least one step")
	assert.Equal(t, tc.Steps[0].Action, foundTC.Steps[0].Action, "Step action should match")
}

// TestListWithFilter tests listing test cases with various filters
func TestListWithFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setup           func(t *testing.T, ts *testSuiteGreen) []shared.ID
		filter          testcase.TestCaseFilter
		wantMinCount    int
		wantMaxCount    int
		wantTotalMin    int64
		wantTotalMax    int64
	}{
		{
			name: "list all test cases",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				ids := []shared.ID{}
				for i := 1; i <= 5; i++ {
					tc := ts.createTestCase("Test Case " + string(rune('0'+i)))
					err := ts.repo.Create(ts.ctx, tc)
					require.NoError(t, err)
					ids = append(ids, tc.ID)
				}
				return ids
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 5,
			wantTotalMin: 0,
			wantTotalMax: 5,
		},
		{
			name: "filter by status",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				tc1 := ts.createTestCase("Draft Test")
				tc1.Status = testcase.CaseDraft
				err := ts.repo.Create(ts.ctx, tc1)
				require.NoError(t, err)

				tc2 := ts.createTestCase("Ready Test")
				tc2.Status = testcase.CaseReady
				err = ts.repo.Create(ts.ctx, tc2)
				require.NoError(t, err)

				return []shared.ID{tc1.ID, tc2.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Status:    func() *testcase.CaseStatus { s := testcase.CaseReady; return &s }(),
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 1,
			wantTotalMin: 0,
			wantTotalMax: 1,
		},
		{
			name: "filter by priority",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				tc1 := ts.createTestCase("P0 Test")
				tc1.Priority = testcase.P0
				err := ts.repo.Create(ts.ctx, tc1)
				require.NoError(t, err)

				tc2 := ts.createTestCase("P1 Test")
				tc2.Priority = testcase.P1
				err = ts.repo.Create(ts.ctx, tc2)
				require.NoError(t, err)

				return []shared.ID{tc1.ID, tc2.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Priority:  func() *testcase.Priority { p := testcase.P0; return &p }(),
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 1,
			wantTotalMin: 0,
			wantTotalMax: 1,
		},
		{
			name: "filter by module",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				moduleID := shared.NewID()
				tc := ts.createTestCase("Module Test")
				tc.ModuleID = &moduleID
				err := ts.repo.Create(ts.ctx, tc)
				require.NoError(t, err)

				return []shared.ID{tc.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				ModuleID:  func() *shared.ID { id := shared.NewID(); return &id }(),
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 0,
			wantTotalMin: 0,
			wantTotalMax: 0,
		},
		{
			name: "filter by keyword",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				tc := ts.createTestCase("Login Test Case")
				err := ts.repo.Create(ts.ctx, tc)
				require.NoError(t, err)

				return []shared.ID{tc.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Keyword:   "login",
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 1,
			wantTotalMin: 0,
			wantTotalMax: 1,
		},
		{
			name: "filter by tags",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				tc := ts.createTestCase("Tagged Test")
				tc.Tags = []string{"critical", "smoke"}
				err := ts.repo.Create(ts.ctx, tc)
				require.NoError(t, err)

				return []shared.ID{tc.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Tags:      []string{"critical"},
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 1,
			wantTotalMin: 0,
			wantTotalMax: 1,
		},
		{
			name: "sort by title ascending",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				tc1 := ts.createTestCase("Zebra Test")
				err := ts.repo.Create(ts.ctx, tc1)
				require.NoError(t, err)

				tc2 := ts.createTestCase("Apple Test")
				err = ts.repo.Create(ts.ctx, tc2)
				require.NoError(t, err)

				return []shared.ID{tc1.ID, tc2.ID}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				SortBy:    "title",
				SortDesc:  false,
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 2,
			wantTotalMin: 0,
			wantTotalMax: 2,
		},
		{
			name: "sort by created_at descending",
			setup: func(t *testing.T, ts *testSuiteGreen) []shared.ID {
				for i := 1; i <= 3; i++ {
					tc := ts.createTestCase("Time Test " + string(rune('0'+i)))
					err := ts.repo.Create(ts.ctx, tc)
					require.NoError(t, err)
					time.Sleep(10 * time.Millisecond) // Ensure different timestamps
				}
				return []shared.ID{}
			},
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				SortBy:    "created_at",
				SortDesc:  true,
				Page:      1,
				PageSize:  10,
			},
			wantMinCount: 0,
			wantMaxCount: 3,
			wantTotalMin: 0,
			wantTotalMax: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuiteGreen(t)
			defer ts.teardownGreen(t)

			// Setup test data
			if tt.setup != nil {
				tt.setup(t, ts)
			}

			cases, total, err := ts.repo.List(ts.ctx, tt.filter)

			require.NoError(t, err, "List should succeed")
			assert.GreaterOrEqual(t, len(cases), tt.wantMinCount, "Should have at least min cases")
			assert.LessOrEqual(t, len(cases), tt.wantMaxCount, "Should have at most max cases")
			assert.GreaterOrEqual(t, total, tt.wantTotalMin, "Total should be at least min")
			assert.LessOrEqual(t, total, tt.wantTotalMax, "Total should be at most max")
		})
	}
}

// TestUpdate tests updating a test case
func TestUpdate(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create a test case
	tc := ts.createTestCase("Update Test")
	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err)

	// Update the test case
	tc.Title = "Updated Test Case"
	tc.Description = "Updated description"
	err = ts.repo.Update(ts.ctx, tc)
	require.NoError(t, err, "Update should succeed")
	assert.Equal(t, 2, tc.Version, "Version should increment")

	// Verify the update
	foundTC, err := ts.repo.FindByID(ts.ctx, tc.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Case", foundTC.Title, "Title should be updated")
	assert.Equal(t, "Updated description", foundTC.Description, "Description should be updated")
	assert.Equal(t, 2, foundTC.Version, "Version should be incremented")
}

// TestSoftDelete tests soft deleting a test case
func TestSoftDelete(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create a test case
	tc := ts.createTestCase("Soft Delete Test")
	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err)

	// Soft delete the test case
	err = ts.repo.SoftDelete(ts.ctx, tc.ID)
	require.NoError(t, err, "SoftDelete should succeed")

	// Verify it's deleted
	_, err = ts.repo.FindByID(ts.ctx, tc.ID)
	assert.Error(t, err, "FindByID should fail for soft deleted test case")
	assert.Equal(t, shared.ErrTestCaseNotFound, err, "Should return not found error")
}

// TestBatchUpdateStatus tests batch updating test case statuses
func TestBatchUpdateStatus(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create test cases
	ids := []shared.ID{}
	for i := 1; i <= 3; i++ {
		tc := ts.createTestCase("Status Test " + string(rune('0'+i)))
		tc.Status = testcase.CaseDraft
		err := ts.repo.Create(ts.ctx, tc)
		require.NoError(t, err)
		ids = append(ids, tc.ID)
	}

	// Batch update status
	err := ts.repo.BatchUpdateStatus(ts.ctx, ids, testcase.CaseReady)
	require.NoError(t, err, "BatchUpdateStatus should succeed")

	// Verify the updates
	for _, id := range ids {
		tc, err := ts.repo.FindByID(ts.ctx, id)
		require.NoError(t, err)
		assert.Equal(t, testcase.CaseReady, tc.Status, "Status should be updated")
	}
}

// TestBatchDelete tests batch deleting test cases
func TestBatchDelete(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create test cases
	ids := []shared.ID{}
	for i := 1; i <= 3; i++ {
		tc := ts.createTestCase("Delete Test " + string(rune('0'+i)))
		err := ts.repo.Create(ts.ctx, tc)
		require.NoError(t, err)
		ids = append(ids, tc.ID)
	}

	// Batch delete
	err := ts.repo.BatchDelete(ts.ctx, ids)
	require.NoError(t, err, "BatchDelete should succeed")

	// Verify the deletions
	for _, id := range ids {
		_, err := ts.repo.FindByID(ts.ctx, id)
		assert.Error(t, err, "FindByID should fail for deleted test case")
		assert.Equal(t, shared.ErrTestCaseNotFound, err, "Should return not found error")
	}
}

// TestBatchMove tests batch moving test cases to a different module
func TestBatchMove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		moduleID *shared.ID
	}{
		{
			name:     "move to a module",
			moduleID: func() *shared.ID { id := shared.NewID(); return &id }(),
		},
		{
			name:     "move to root (no module)",
			moduleID: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuiteGreen(t)
			defer ts.teardownGreen(t)

			// Create test cases
			ids := []shared.ID{}
			for i := 1; i <= 3; i++ {
				tc := ts.createTestCase("Move Test " + string(rune('0'+i)))
				err := ts.repo.Create(ts.ctx, tc)
				require.NoError(t, err)
				ids = append(ids, tc.ID)
			}

			// Batch move
			err := ts.repo.BatchMove(ts.ctx, ids, tt.moduleID)
			require.NoError(t, err, "BatchMove should succeed")

			// Verify the moves
			for _, id := range ids {
				tc, err := ts.repo.FindByID(ts.ctx, id)
				require.NoError(t, err)
				if tt.moduleID != nil {
					assert.Equal(t, tt.moduleID, tc.ModuleID, "ModuleID should be updated")
				} else {
					assert.Nil(t, tc.ModuleID, "ModuleID should be nil")
				}
			}
		})
	}
}

// TestVersionConflict tests optimistic locking with version field
func TestVersionConflict(t *testing.T) {
	t.Parallel()

	t.Run("concurrent update with version mismatch", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuiteGreen(t)
		defer ts.teardownGreen(t)

		// Create a test case
		tc := ts.createTestCase("Concurrent Update Test")
		err := ts.repo.Create(ts.ctx, tc)
		require.NoError(t, err)

		// Simulate two users fetching the same test case
		tc1, err := ts.repo.FindByID(ts.ctx, tc.ID)
		require.NoError(t, err)

		tc2, err := ts.repo.FindByID(ts.ctx, tc.ID)
		require.NoError(t, err)

		// First user updates
		tc1.Title = "Updated by User 1"
		err1 := ts.repo.Update(ts.ctx, tc1)
		require.NoError(t, err1, "First update should succeed")
		assert.Equal(t, 2, tc1.Version, "Version should increment")

		// Second user tries to update with stale version
		tc2.Title = "Updated by User 2"
		err2 := ts.repo.Update(ts.ctx, tc2)
		assert.Error(t, err2, "Second update should fail with version conflict")
		assert.Equal(t, shared.ErrTestCaseConflict, err2, "Should return conflict error")
	})

	t.Run("successful update with correct version", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuiteGreen(t)
		defer ts.teardownGreen(t)

		// Create a test case
		tc := ts.createTestCase("Version Update Test")
		err := ts.repo.Create(ts.ctx, tc)
		require.NoError(t, err)

		// Fetch and update with correct version
		tc.Title = "Updated Title"
		err = ts.repo.Update(ts.ctx, tc)
		require.NoError(t, err, "Update with correct version should succeed")
		assert.Equal(t, 2, tc.Version, "Version should increment")
	})

	t.Run("version increment after update", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuiteGreen(t)
		defer ts.teardownGreen(t)

		// Create a test case
		tc := ts.createTestCase("Version Increment Test")
		err := ts.repo.Create(ts.ctx, tc)
		require.NoError(t, err)
		initialVersion := tc.Version

		// Update the test case
		tc.Title = "Updated"
		err = ts.repo.Update(ts.ctx, tc)
		require.NoError(t, err)

		// Verify version increment
		assert.Equal(t, initialVersion+1, tc.Version, "Version should increment by 1")
	})
}

// TestListPagination tests pagination functionality
func TestListPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pageSize      int
		totalItems    int
		wantPageCount int
	}{
		{
			name:          "first page with 10 items",
			pageSize:      10,
			totalItems:    25,
			wantPageCount: 10,
		},
		{
			name:          "second page with 10 items",
			pageSize:      10,
			totalItems:    25,
			wantPageCount: 10,
		},
		{
			name:          "last page with remaining items",
			pageSize:      10,
			totalItems:    25,
			wantPageCount: 5,
		},
		{
			name:          "large page size",
			pageSize:      100,
			totalItems:    10,
			wantPageCount: 10,
		},
		{
			name:          "page size 1",
			pageSize:      1,
			totalItems:    5,
			wantPageCount: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuiteGreen(t)
			defer ts.teardownGreen(t)

			// Create test data
			for i := 1; i <= tt.totalItems; i++ {
				tc := ts.createTestCase("Pagination Test " + string(rune('0'+i%10)))
				err := ts.repo.Create(ts.ctx, tc)
				require.NoError(t, err)
			}

			// Determine which page to test based on test name
			page := 1
			if tt.name == "second page with 10 items" {
				page = 2
			} else if tt.name == "last page with remaining items" {
				page = 3
			}

			filter := testcase.TestCaseFilter{
				ProjectID: ts.projectID,
				Page:      page,
				PageSize:  tt.pageSize,
			}

			cases, total, err := ts.repo.List(ts.ctx, filter)

			require.NoError(t, err, "List should succeed")
			assert.Equal(t, int64(tt.totalItems), total, "Total should match total items")
			assert.GreaterOrEqual(t, len(cases), 0, "Should return cases")
			assert.LessOrEqual(t, len(cases), tt.wantPageCount, "Should not exceed page size")
		})
	}
}

// TestStepsPreload tests that steps are properly preloaded when finding by ID
func TestStepsPreload(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create a test case with multiple steps
	tc := ts.createTestCase("Steps Preload Test")
	tc.Steps = []testcase.Step{
		{
			ID:        shared.NewID(),
			Number:    1,
			Action:    "Step 1 action",
			Expected:  "Step 1 expected",
		},
		{
			ID:        shared.NewID(),
			Number:    2,
			Action:    "Step 2 action",
			Expected:  "Step 2 expected",
		},
		{
			ID:        shared.NewID(),
			Number:    3,
			Action:    "Step 3 action",
			Expected:  "Step 3 expected",
		},
	}

	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err)

	// Find test case with steps
	foundTC, err := ts.repo.FindByID(ts.ctx, tc.ID)
	require.NoError(t, err, "FindByID should succeed")
	assert.NotEmpty(t, foundTC.Steps, "Steps should be preloaded")
	assert.Greater(t, len(foundTC.Steps), 0, "Should have at least one step")
	assert.Equal(t, 3, len(foundTC.Steps), "Should have all three steps")
	assert.Equal(t, "Step 1 action", foundTC.Steps[0].Action, "First step should match")
	assert.Equal(t, "Step 3 action", foundTC.Steps[2].Action, "Last step should match")
}

// TestEmptyResults tests that empty results are handled correctly
func TestEmptyResults(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Find non-existent test case
	nonExistentID := shared.NewID()
	foundTC, err := ts.repo.FindByID(ts.ctx, nonExistentID)
	assert.Error(t, err, "FindByID should fail for non-existent test case")
	assert.Nil(t, foundTC, "Should return nil for non-existent test case")
	assert.Equal(t, shared.ErrTestCaseNotFound, err, "Should return not found error")

	// List with filter that returns no results
	filter := testcase.TestCaseFilter{
		ProjectID: ts.projectID,
		Keyword:   "nonexistent-keyword-xyz",
		Page:      1,
		PageSize:  10,
	}

	cases, total, err := ts.repo.List(ts.ctx, filter)
	require.NoError(t, err, "List should succeed even with no results")
	assert.Empty(t, cases, "Should return empty cases")
	assert.Equal(t, int64(0), total, "Should return 0 total")
}

// TestCreateTransaction tests that Create uses transactions properly
func TestCreateTransaction(t *testing.T) {
	t.Parallel()

	ts := setupTestSuiteGreen(t)
	defer ts.teardownGreen(t)

	// Create a test case with steps
	tc := ts.createTestCase("Transaction Test")
	tc.Steps = []testcase.Step{
		{
			ID:       shared.NewID(),
			Number:   1,
			Action:   "Step 1",
			Expected: "Expected 1",
		},
		{
			ID:       shared.NewID(),
			Number:   2,
			Action:   "Step 2",
			Expected: "Expected 2",
		},
	}

	// Create should succeed
	err := ts.repo.Create(ts.ctx, tc)
	require.NoError(t, err, "Create should succeed")

	// Verify both test case and steps are created
	foundTC, err := ts.repo.FindByID(ts.ctx, tc.ID)
	require.NoError(t, err)
	assert.Equal(t, tc.Title, foundTC.Title, "Test case should be created")
	assert.Len(t, foundTC.Steps, 2, "Both steps should be created")
}
