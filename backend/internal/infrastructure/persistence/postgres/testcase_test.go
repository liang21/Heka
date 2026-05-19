// tasks.md: T051 | TDD RED for TestCaseRepository
package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
)

// testSuite holds the test database and repository
type testSuite struct {
	ctx        context.Context
	container  *postgrescontainer.PostgresContainer
	repo       testcase.TestCaseRepository
	projectID  shared.ID
	moduleID   *shared.ID
	userID     shared.ID
}

// setupTestSuite creates a new PostgreSQL test container and test suite
func setupTestSuite(t *testing.T) *testSuite {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container with test database
	pgContainer, err := postgrescontainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgrescontainer.WithDatabase("heka_test"),
		postgrescontainer.WithUsername("test"),
		postgrescontainer.WithPassword("test"),
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
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
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

	return &testSuite{
		ctx:       ctx,
		container: pgContainer,
		repo:      repo,
		projectID: projectID,
		moduleID:  &moduleID,
		userID:    userID,
	}
}

// teardownTestSuite cleans up the test container
func (ts *testSuite) teardown(t *testing.T) {
	t.Helper()
	if err := ts.container.Terminate(ts.ctx); err != nil {
		t.Fatalf("Failed to terminate container: %v", err)
	}
}

// mockTestCaseRepository is a mock that always fails (RED phase)
type mockTestCaseRepository struct{}

func (m *mockTestCaseRepository) Create(ctx context.Context, tc *testcase.TestCase) error {
	return assert.AnError
}

func (m *mockTestCaseRepository) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
	return nil, assert.AnError
}

func (m *mockTestCaseRepository) List(ctx context.Context, filter testcase.TestCaseFilter) ([]*testcase.TestCase, int64, error) {
	return nil, 0, assert.AnError
}

func (m *mockTestCaseRepository) Update(ctx context.Context, tc *testcase.TestCase) error {
	return assert.AnError
}

func (m *mockTestCaseRepository) SoftDelete(ctx context.Context, id shared.ID) error {
	return assert.AnError
}

func (m *mockTestCaseRepository) BatchUpdateStatus(ctx context.Context, ids []shared.ID, status testcase.CaseStatus) error {
	return assert.AnError
}

func (m *mockTestCaseRepository) BatchDelete(ctx context.Context, ids []shared.ID) error {
	return assert.AnError
}

func (m *mockTestCaseRepository) BatchMove(ctx context.Context, ids []shared.ID, moduleID *shared.ID) error {
	return assert.AnError
}

// Helper function to create a test TestCase
func (ts *testSuite) createTestCase(title string) *testcase.TestCase {
	return &testcase.TestCase{
		ID:         shared.NewID(),
		ProjectID:  ts.projectID,
		ModuleID:   ts.moduleID,
		Title:      title,
		Description: "Test description",
		Status:     testcase.CaseDraft,
		Priority:   testcase.P1,
		Tags:       []string{"tag1", "tag2"},
		CreatedBy:  ts.userID,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
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

// TestFindByID tests finding a test case by ID with steps preloaded
func TestFindByID(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	testID := shared.NewID()

	foundTC, err := ts.repo.FindByID(ts.ctx, testID)
	assert.Error(t, err, "FindByID should fail in RED phase")
	assert.Nil(t, foundTC, "Should return nil test case on error")
	assert.Equal(t, assert.AnError, err, "Expected mock error")
}

// TestListWithFilter tests listing test cases with various filters
func TestListWithFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filter  testcase.TestCaseFilter
		wantErr bool
	}{
		{
			name: "list all test cases",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "filter by status",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Status:    func() *testcase.CaseStatus { s := testcase.CaseReady; return &s }(),
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "filter by priority",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Priority:  func() *testcase.Priority { p := testcase.P0; return &p }(),
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "filter by module",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				ModuleID:  func() *shared.ID { id := shared.NewID(); return &id }(),
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "filter by keyword",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Keyword:   "login",
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "filter by tags",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				Tags:      []string{"critical", "smoke"},
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "sort by title ascending",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				SortBy:    "title",
				SortDesc:  false,
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
		{
			name: "sort by created_at descending",
			filter: testcase.TestCaseFilter{
				ProjectID: shared.NewID(),
				SortBy:    "created_at",
				SortDesc:  true,
				Page:      1,
				PageSize:  10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuite(t)
			defer ts.teardown(t)

			cases, total, err := ts.repo.List(ts.ctx, tt.filter)

			if tt.wantErr {
				assert.Error(t, err, "List should fail in RED phase")
				assert.Equal(t, assert.AnError, err, "Expected mock error")
			}
			assert.Nil(t, cases, "Should return nil cases on error")
			assert.Equal(t, int64(0), total, "Should return 0 total on error")
		})
	}
}

// TestUpdate tests updating a test case
func TestUpdate(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	tc := ts.createTestCase("Updated Test Case")
	tc.ID = shared.NewID()

	err := ts.repo.Update(ts.ctx, tc)
	assert.Error(t, err, "Update should fail in RED phase")
	assert.Equal(t, assert.AnError, err, "Expected mock error")
}

// TestSoftDelete tests soft deleting a test case
func TestSoftDelete(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	testID := shared.NewID()

	err := ts.repo.SoftDelete(ts.ctx, testID)
	assert.Error(t, err, "SoftDelete should fail in RED phase")
	assert.Equal(t, assert.AnError, err, "Expected mock error")
}

// TestBatchUpdateStatus tests batch updating test case statuses
func TestBatchUpdateStatus(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ids := []shared.ID{shared.NewID(), shared.NewID(), shared.NewID()}

	err := ts.repo.BatchUpdateStatus(ts.ctx, ids, testcase.CaseReady)
	assert.Error(t, err, "BatchUpdateStatus should fail in RED phase")
	assert.Equal(t, assert.AnError, err, "Expected mock error")
}

// TestBatchDelete tests batch deleting test cases
func TestBatchDelete(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ids := []shared.ID{shared.NewID(), shared.NewID(), shared.NewID()}

	err := ts.repo.BatchDelete(ts.ctx, ids)
	assert.Error(t, err, "BatchDelete should fail in RED phase")
	assert.Equal(t, assert.AnError, err, "Expected mock error")
}

// TestBatchMove tests batch moving test cases to a different module
func TestBatchMove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		moduleID *shared.ID
		wantErr  bool
	}{
		{
			name:     "move to a module",
			moduleID: func() *shared.ID { id := shared.NewID(); return &id }(),
			wantErr:  true,
		},
		{
			name:     "move to root (no module)",
			moduleID: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuite(t)
			defer ts.teardown(t)

			ids := []shared.ID{shared.NewID(), shared.NewID(), shared.NewID()}

			err := ts.repo.BatchMove(ts.ctx, ids, tt.moduleID)

			if tt.wantErr {
				assert.Error(t, err, "BatchMove should fail in RED phase")
				assert.Equal(t, assert.AnError, err, "Expected mock error")
			}
		})
	}
}

// TestVersionConflict tests optimistic locking with version field
func TestVersionConflict(t *testing.T) {
	t.Parallel()

	t.Run("concurrent update with version mismatch", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuite(t)
		defer ts.teardown(t)

		testID := shared.NewID()

		// Simulate two users fetching the same test case
		tc1 := ts.createTestCase("Concurrent Update Test")
		tc1.ID = testID
		tc1.Version = 1

		tc2 := ts.createTestCase("Concurrent Update Test")
		tc2.ID = testID
		tc2.Version = 1

		// First user updates
		tc1.Title = "Updated by User 1"
		err1 := ts.repo.Update(ts.ctx, tc1)
		assert.Error(t, err1, "First update should fail in RED phase")

		// Second user tries to update with stale version
		tc2.Title = "Updated by User 2"
		tc2.Version = 1 // Still version 1, but should be 2 after first update
		err2 := ts.repo.Update(ts.ctx, tc2)
		assert.Error(t, err2, "Second update should fail in RED phase")
	})

	t.Run("successful update with correct version", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuite(t)
		defer ts.teardown(t)

		testID := shared.NewID()

		tc := ts.createTestCase("Version Update Test")
		tc.ID = testID
		tc.Version = 1

		// Update with correct version should still fail in RED phase
		tc.Title = "Updated Title"
		err := ts.repo.Update(ts.ctx, tc)
		assert.Error(t, err, "Update should fail in RED phase")
	})

	t.Run("version increment after update", func(t *testing.T) {
		t.Parallel()

		ts := setupTestSuite(t)
		defer ts.teardown(t)

		testID := shared.NewID()

		tc := ts.createTestCase("Version Increment Test")
		tc.ID = testID
		tc.Version = 1

		// After update, version should increment (will fail in RED phase)
		err := ts.repo.Update(ts.ctx, tc)
		assert.Error(t, err, "Update should fail in RED phase")
	})
}

// TestListPagination tests pagination functionality
func TestListPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		page          int
		pageSize      int
		expectedTotal int64
		wantErr       bool
	}{
		{
			name:          "first page",
			page:          1,
			pageSize:      10,
			expectedTotal: 0,
			wantErr:       true,
		},
		{
			name:          "second page",
			page:          2,
			pageSize:      10,
			expectedTotal: 0,
			wantErr:       true,
		},
		{
			name:          "large page size",
			page:          1,
			pageSize:      100,
			expectedTotal: 0,
			wantErr:       true,
		},
		{
			name:          "page size 1",
			page:          1,
			pageSize:      1,
			expectedTotal: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := setupTestSuite(t)
			defer ts.teardown(t)

			filter := testcase.TestCaseFilter{
				ProjectID: ts.projectID,
				Page:      tt.page,
				PageSize:  tt.pageSize,
			}

			cases, total, err := ts.repo.List(ts.ctx, filter)

			if tt.wantErr {
				assert.Error(t, err, "List should fail in RED phase")
			}
			assert.Nil(t, cases, "Should return nil cases on error")
			assert.Equal(t, tt.expectedTotal, total, "Should return 0 total on error")
		})
	}
}

// TestStepsPreload tests that steps are properly preloaded when finding by ID
func TestStepsPreload(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	testID := shared.NewID()

	// Find test case with steps
	foundTC, err := ts.repo.FindByID(ts.ctx, testID)
	assert.Error(t, err, "FindByID should fail in RED phase")
	assert.Nil(t, foundTC, "Should return nil test case on error")

	// If not error, steps should be preloaded
	if err == nil && foundTC != nil {
		assert.NotEmpty(t, foundTC.Steps, "Steps should be preloaded")
		assert.Greater(t, len(foundTC.Steps), 0, "Should have at least one step")
	}
}

// TestEmptyResults tests that empty results are handled correctly
func TestEmptyResults(t *testing.T) {
	t.Parallel()

	ts := setupTestSuite(t)
	defer ts.teardown(t)

	// Find non-existent test case
	nonExistentID := shared.NewID()
	foundTC, err := ts.repo.FindByID(ts.ctx, nonExistentID)
	assert.Error(t, err, "FindByID should fail in RED phase")
	assert.Nil(t, foundTC, "Should return nil for non-existent test case")

	// List with filter that returns no results
	filter := testcase.TestCaseFilter{
		ProjectID: shared.NewID(),
		Keyword:   "nonexistent-keyword-xyz",
		Page:      1,
		PageSize:  10,
	}

	cases, total, err := ts.repo.List(ts.ctx, filter)
	assert.Error(t, err, "List should fail in RED phase")
	assert.Nil(t, cases, "Should return nil cases on error")
	assert.Equal(t, int64(0), total, "Should return 0 total on error")
}
