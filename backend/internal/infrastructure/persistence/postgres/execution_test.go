// tasks.md: T057 | TDD RED phase - ExecutionRepository implementation tests
package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupExecutionTestDB creates a test database connection using testcontainers-go
// This function should be replaced with actual testcontainers setup once added to go.mod
func setupExecutionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// TODO: Replace with testcontainers-go setup
	// pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
	//     postgres.WithDatabase("testdb"),
	//     postgres.WithUsername("test"),
	//     postgres.WithPassword("test"),
	// )

	// For now, using environment variable or hardcoded test DSN
	dsn := "host=localhost port=5432 user=test_user password=test_pass dbname=test_db sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
	}

	// Auto-migrate the TestExecution and ExecutionResult schemas
	err = db.AutoMigrate(&execution.TestExecution{}, &execution.ExecutionResult{})
	if err != nil {
		t.Fatalf("Failed to migrate Execution schemas: %v", err)
	}

	// Clean tables before each test
	db.Exec("DELETE FROM execution_results")
	db.Exec("DELETE FROM test_executions")

	return db
}

// createExecutionRepo creates a new ExecutionRepository instance for testing
func createExecutionRepo(db *gorm.DB) execution.ExecutionRepository {
	return NewExecutionRepository(db)
}

// createTestExecution creates a test execution with default values
func createTestExecution(planID shared.ID, name string) *execution.TestExecution {
	now := time.Now()
	return &execution.TestExecution{
		ID:         shared.NewID(),
		PlanID:     planID,
		Name:       name,
		Status:     execution.ExecInProgress,
		ExecutorID: shared.NewID(),
		StartedAt:  now,
		Notes:      "Test execution notes",
	}
}

// createTestResult creates a test execution result with default values
func createTestResult(executionID, testCaseID shared.ID, status execution.ResultStatus) *execution.ExecutionResult {
	return &execution.ExecutionResult{
		ID:          shared.NewID(),
		ExecutionID: executionID,
		TestCaseID:  testCaseID,
		ExecutorID:  shared.NewID(),
		Status:      status,
		BugID:       "",
		BugURL:      "",
		Notes:       "Test result notes",
		ExecutedAt:  time.Now(),
	}
}

func TestExecutionRepository_ExecutionRepository_CreateExecution(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Test Execution 1")

	err := repo.Create(ctx, testExec)
	assert.NoError(t, err, "Create should succeed with valid execution")
	assert.NotEmpty(t, testExec.ID, "Execution ID should be set")

	// Verify the execution was created
	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, testExec.ID, found.ID)
	assert.Equal(t, testExec.PlanID, found.PlanID)
	assert.Equal(t, testExec.Name, found.Name)
	assert.Equal(t, execution.ExecInProgress, found.Status)
}

func TestExecutionRepository_CreateExecution_WithAllFields(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	executorID := shared.NewID()
	now := time.Now()

	testExec := &execution.TestExecution{
		ID:         shared.NewID(),
		PlanID:     planID,
		Name:       "Complete Test Execution",
		Status:     execution.ExecInProgress,
		ExecutorID: executorID,
		StartedAt:  now,
		PausedAt:   nil,
		CompletedAt: nil,
		Notes:      "This is a complete test execution with all fields populated",
	}

	err := repo.Create(ctx, testExec)
	assert.NoError(t, err, "Create should succeed with all fields populated")

	// Verify all fields were persisted
	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, testExec.Name, found.Name)
	assert.Equal(t, testExec.ExecutorID, found.ExecutorID)
	assert.Equal(t, testExec.Notes, found.Notes)
	assert.WithinDuration(t, testExec.StartedAt, found.StartedAt, time.Second)
}

func TestExecutionRepository_CreateExecution_CompletedStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	now := time.Now()
	completedAt := now.Add(1 * time.Hour)

	testExec := &execution.TestExecution{
		ID:          shared.NewID(),
		PlanID:      planID,
		Name:        "Completed Execution",
		Status:      execution.ExecCompleted,
		ExecutorID:  shared.NewID(),
		StartedAt:   now,
		CompletedAt: &completedAt,
		Notes:       "Execution completed successfully",
	}

	err := repo.Create(ctx, testExec)
	assert.NoError(t, err, "Create should succeed with completed status")

	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecCompleted, found.Status)
	assert.NotNil(t, found.CompletedAt)
	assert.WithinDuration(t, completedAt, *found.CompletedAt, time.Second)
}

func TestExecutionRepository_CreateExecution_PausedStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	now := time.Now()
	pausedAt := now.Add(30 * time.Minute)

	testExec := &execution.TestExecution{
		ID:         shared.NewID(),
		PlanID:     planID,
		Name:       "Paused Execution",
		Status:     execution.ExecPaused,
		ExecutorID: shared.NewID(),
		StartedAt:  now,
		PausedAt:   &pausedAt,
		Notes:      "Execution paused for review",
	}

	err := repo.Create(ctx, testExec)
	assert.NoError(t, err, "Create should succeed with paused status")

	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecPaused, found.Status)
	assert.NotNil(t, found.PausedAt)
	assert.WithinDuration(t, pausedAt, *found.PausedAt, time.Second)
}

func TestExecutionRepository_CreateExecution_CancelledStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	now := time.Now()
	cancelledAt := now.Add(15 * time.Minute)

	testExec := &execution.TestExecution{
		ID:          shared.NewID(),
		PlanID:      planID,
		Name:        "Cancelled Execution",
		Status:      execution.ExecCancelled,
		ExecutorID:  shared.NewID(),
		StartedAt:   now,
		CompletedAt: &cancelledAt,
		Notes:       "Execution cancelled due to blocking issues",
	}

	err := repo.Create(ctx, testExec)
	assert.NoError(t, err, "Create should succeed with cancelled status")

	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecCancelled, found.Status)
}

func TestExecutionRepository_ExecutionRepository_FindByID(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Find By ID Test")

	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err, "FindByID should succeed")
	assert.NotNil(t, found)
	assert.Equal(t, testExec.ID, found.ID)
	assert.Equal(t, testExec.Name, found.Name)
	assert.Equal(t, testExec.PlanID, found.PlanID)
}

func TestExecutionRepository_ExecutionRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	nonExistentID := shared.NewID()

	found, err := repo.FindByID(ctx, nonExistentID)
	assert.Error(t, err, "FindByID should return error for non-existent ID")
	assert.Nil(t, found)
}

func TestExecutionRepository_ExecutionRepository_SubmitResult(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	// Create test execution
	planID := shared.NewID()
	testExec := createTestExecution(planID, "Submit Result Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	// Submit a result
	testCaseID := shared.NewID()
	result := createTestResult(testExec.ID, testCaseID, execution.ResultPassed)

	err = repo.SubmitResult(ctx, result)
	assert.NoError(t, err, "SubmitResult should succeed")
	assert.NotEmpty(t, result.ID, "Result ID should be set")

	// Verify the result was submitted
	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, summary.Total)
	assert.Equal(t, 1, summary.Passed)
	assert.Equal(t, 0, summary.Failed)
}

func TestExecutionRepository_SubmitResult_FailedStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Failed Result Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	testCaseID := shared.NewID()
	result := createTestResult(testExec.ID, testCaseID, execution.ResultFailed)
	result.BugID = "BUG-123"
	result.BugURL = "https://bugtracker.example.com/BUG-123"
	result.Notes = "Test failed due to assertion error"

	err = repo.SubmitResult(ctx, result)
	assert.NoError(t, err, "SubmitResult should succeed for failed status")

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, summary.Failed)
}

func TestExecutionRepository_SubmitResult_BlockedStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Blocked Result Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	testCaseID := shared.NewID()
	result := createTestResult(testExec.ID, testCaseID, execution.ResultBlocked)
	result.Notes = "Test blocked due to environment unavailability"

	err = repo.SubmitResult(ctx, result)
	assert.NoError(t, err, "SubmitResult should succeed for blocked status")

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, summary.Blocked)
}

func TestExecutionRepository_SubmitResult_SkippedStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Skipped Result Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	testCaseID := shared.NewID()
	result := createTestResult(testExec.ID, testCaseID, execution.ResultSkipped)
	result.Notes = "Test skipped - feature not implemented yet"

	err = repo.SubmitResult(ctx, result)
	assert.NoError(t, err, "SubmitResult should succeed for skipped status")

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, summary.Skipped)
}

func TestExecutionRepository_SubmitResult_InvalidExecutionID(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	nonExistentExecutionID := shared.NewID()
	testCaseID := shared.NewID()
	result := createTestResult(nonExistentExecutionID, testCaseID, execution.ResultPassed)

	err := repo.SubmitResult(ctx, result)
	assert.Error(t, err, "SubmitResult should fail with invalid execution ID")
}

func TestExecutionRepository_SubmitResult_DuplicateTestCase(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Duplicate Test Case Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	testCaseID := shared.NewID()

	// Submit first result
	result1 := createTestResult(testExec.ID, testCaseID, execution.ResultPassed)
	err = repo.SubmitResult(ctx, result1)
	require.NoError(t, err)

	// Try to submit duplicate result for same test case
	result2 := createTestResult(testExec.ID, testCaseID, execution.ResultFailed)
	result2.Notes = "Retrying the same test case"

	err = repo.SubmitResult(ctx, result2)
	// This should either fail (unique constraint) or update (upsert behavior)
	// The test documents the expected behavior
	assert.Error(t, err, "SubmitResult should handle duplicate test case submissions")
}

func TestExecutionRepository_BatchSubmitResults(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Batch Submit Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	// Create multiple results
	const numResults = 5
	results := make([]*execution.ExecutionResult, numResults)
	for i := 0; i < numResults; i++ {
		testCaseID := shared.NewID()
		status := execution.ResultPassed
		if i%2 == 0 {
			status = execution.ResultFailed
		}
		results[i] = createTestResult(testExec.ID, testCaseID, status)
	}

	err = repo.BatchSubmitResults(ctx, results)
	assert.NoError(t, err, "BatchSubmitResults should succeed")

	// Verify all results were submitted
	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, numResults, summary.Total)
	assert.Equal(t, 3, summary.Passed)  // Odd indices
	assert.Equal(t, 2, summary.Failed)  // Even indices
}

func TestExecutionRepository_BatchSubmitResults_EmptyBatch(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Empty Batch Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	emptyBatch := []*execution.ExecutionResult{}

	err = repo.BatchSubmitResults(ctx, emptyBatch)
	assert.NoError(t, err, "BatchSubmitResults should handle empty batch gracefully")

	// Verify no results were submitted
	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, summary.Total)
}

func TestExecutionRepository_BatchSubmitResults_LargeBatch(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Large Batch Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	// Create a large batch of results
	const numResults = 100
	results := make([]*execution.ExecutionResult, numResults)
	for i := 0; i < numResults; i++ {
		testCaseID := shared.NewID()
		status := execution.ResultPassed
		switch i % 4 {
		case 0:
			status = execution.ResultPassed
		case 1:
			status = execution.ResultFailed
		case 2:
			status = execution.ResultBlocked
		case 3:
			status = execution.ResultSkipped
		}
		results[i] = createTestResult(testExec.ID, testCaseID, status)
	}

	err = repo.BatchSubmitResults(ctx, results)
	assert.NoError(t, err, "BatchSubmitResults should handle large batches")

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, numResults, summary.Total)
	assert.Equal(t, 25, summary.Passed)
	assert.Equal(t, 25, summary.Failed)
	assert.Equal(t, 25, summary.Blocked)
	assert.Equal(t, 25, summary.Skipped)
}

func TestExecutionRepository_BatchSubmitResults_WithDuplicateTestCases(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Batch Duplicate Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	testCaseID := shared.NewID()

	// Create batch with duplicate test case IDs
	results := []*execution.ExecutionResult{
		createTestResult(testExec.ID, testCaseID, execution.ResultPassed),
		createTestResult(testExec.ID, testCaseID, execution.ResultFailed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
	}

	err = repo.BatchSubmitResults(ctx, results)
	// Should either fail (unique constraint) or handle duplicates (upsert/ignore)
	// The test documents the expected behavior
	assert.Error(t, err, "BatchSubmitResults should handle duplicate test cases")
}

func TestExecutionRepository_GetSummary(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "Summary Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	// Submit results with different statuses
	results := []*execution.ExecutionResult{
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultFailed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultFailed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultBlocked),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultSkipped),
	}

	err = repo.BatchSubmitResults(ctx, results)
	require.NoError(t, err)

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err, "GetSummary should succeed")
	assert.Equal(t, 7, summary.Total)
	assert.Equal(t, 3, summary.Passed)
	assert.Equal(t, 2, summary.Failed)
	assert.Equal(t, 1, summary.Blocked)
	assert.Equal(t, 1, summary.Skipped)
}

func TestExecutionRepository_GetSummary_NoResults(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "No Results Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err, "GetSummary should succeed with no results")
	assert.Equal(t, 0, summary.Total)
	assert.Equal(t, 0, summary.Passed)
	assert.Equal(t, 0, summary.Failed)
	assert.Equal(t, 0, summary.Blocked)
	assert.Equal(t, 0, summary.Skipped)
}

func TestExecutionRepository_GetSummary_OnlyPassed(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "All Passed Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	results := []*execution.ExecutionResult{
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultPassed),
	}

	err = repo.BatchSubmitResults(ctx, results)
	require.NoError(t, err)

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, summary.Total)
	assert.Equal(t, 3, summary.Passed)
	assert.Equal(t, 0, summary.Failed)
	assert.Equal(t, 0, summary.Blocked)
	assert.Equal(t, 0, summary.Skipped)
}

func TestExecutionRepository_GetSummary_OnlyFailed(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	testExec := createTestExecution(planID, "All Failed Test")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	results := []*execution.ExecutionResult{
		createTestResult(testExec.ID, shared.NewID(), execution.ResultFailed),
		createTestResult(testExec.ID, shared.NewID(), execution.ResultFailed),
	}

	err = repo.BatchSubmitResults(ctx, results)
	require.NoError(t, err)

	summary, err := repo.GetSummary(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, summary.Total)
	assert.Equal(t, 0, summary.Passed)
	assert.Equal(t, 2, summary.Failed)
	assert.Equal(t, 0, summary.Blocked)
	assert.Equal(t, 0, summary.Skipped)
}

func TestExecutionRepository_GetSummary_InvalidExecutionID(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	nonExistentID := shared.NewID()

	summary, err := repo.GetSummary(ctx, nonExistentID)
	assert.Error(t, err, "GetSummary should fail with invalid execution ID")
	assert.Nil(t, summary)
}

func TestExecutionRepository_ConcurrentCreate(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()

	// Try to create multiple executions for the same plan concurrently
	const numGoroutines = 10
	executions := make([]*execution.TestExecution, numGoroutines)
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			exec := createTestExecution(planID, fmt.Sprintf("Concurrent Execution %d", idx))
			err := repo.Create(ctx, exec)
			if err != nil {
				errChan <- err
			}
			executions[idx] = exec
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Count successful creations and errors
	successCount := 0
	errorCount := 0
	for range errChan {
		errorCount++
	}
	for _, exec := range executions {
		if exec != nil && !exec.ID.IsEmpty() {
			successCount++
		}
	}

	// Only one execution should succeed due to partial unique index
	// (plan_id, status) where status = 'in_progress'
	assert.Equal(t, 1, successCount, "Only one in_progress execution should exist per plan")
	assert.Equal(t, numGoroutines-1, errorCount, "All other concurrent creates should fail")

	// Verify the single execution exists
	found, err := repo.FindByID(ctx, executions[0].ID)
	if err == nil {
		assert.Equal(t, execution.ExecInProgress, found.Status)
		assert.Equal(t, planID, found.PlanID)
	}
}

func TestExecutionRepository_ConcurrentCreate_DifferentPlans(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	// Create executions for different plans concurrently
	const numPlans = 5
	executions := make([]*execution.TestExecution, numPlans)
	var wg sync.WaitGroup

	for i := 0; i < numPlans; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			planID := shared.NewID()
			exec := createTestExecution(planID, fmt.Sprintf("Plan %d Execution", idx))
			err := repo.Create(ctx, exec)
			if err != nil {
				t.Errorf("Create failed for plan %d: %v", idx, err)
			}
			executions[idx] = exec
		}(i)
	}

	wg.Wait()

	// All executions for different plans should succeed
	successCount := 0
	for _, exec := range executions {
		if exec != nil && !exec.ID.IsEmpty() {
			successCount++
		}
	}
	assert.Equal(t, numPlans, successCount, "All executions for different plans should succeed")
}

func TestExecutionRepository_ConcurrentCreate_NonInProgressStatus(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()

	// Create first execution with in_progress status
	exec1 := createTestExecution(planID, "First Execution")
	err := repo.Create(ctx, exec1)
	require.NoError(t, err)

	// Create second execution with completed status (should not conflict)
	now := time.Now()
	completedAt := now.Add(1 * time.Hour)
	exec2 := &execution.TestExecution{
		ID:          shared.NewID(),
		PlanID:      planID,
		Name:        "Second Execution",
		Status:      execution.ExecCompleted,
		ExecutorID:  shared.NewID(),
		StartedAt:   now,
		CompletedAt: &completedAt,
		Notes:       "Completed execution",
	}

	err = repo.Create(ctx, exec2)
	assert.NoError(t, err, "Create should succeed with completed status for same plan")

	// Create third execution with cancelled status
	cancelledAt := now.Add(2 * time.Hour)
	exec3 := &execution.TestExecution{
		ID:          shared.NewID(),
		PlanID:      planID,
		Name:        "Third Execution",
		Status:      execution.ExecCancelled,
		ExecutorID:  shared.NewID(),
		StartedAt:   now,
		CompletedAt: &cancelledAt,
		Notes:       "Cancelled execution",
	}

	err = repo.Create(ctx, exec3)
	assert.NoError(t, err, "Create should succeed with cancelled status for same plan")

	// Verify all three executions exist
	found1, err := repo.FindByID(ctx, exec1.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecInProgress, found1.Status)

	found2, err := repo.FindByID(ctx, exec2.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecCompleted, found2.Status)

	found3, err := repo.FindByID(ctx, exec3.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecCancelled, found3.Status)
}

func TestExecutionRepository_ExecutionRepository_MultipleExecutionsPerPlan(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	now := time.Now()

	// Create first execution
	exec1 := createTestExecution(planID, "Execution 1")
	err := repo.Create(ctx, exec1)
	require.NoError(t, err)

	// Complete first execution
	completedAt := now.Add(1 * time.Hour)
	exec1.Status = execution.ExecCompleted
	exec1.CompletedAt = &completedAt

	// Create second execution (should succeed since first is completed)
	exec2 := createTestExecution(planID, "Execution 2")
	err = repo.Create(ctx, exec2)
	assert.NoError(t, err, "Should be able to create new execution after previous one completes")

	// Verify both executions exist
	found1, err := repo.FindByID(ctx, exec1.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecCompleted, found1.Status)

	found2, err := repo.FindByID(ctx, exec2.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ExecInProgress, found2.Status)
}

func TestExecutionRepository_ExecutionRepository_IntegrationWorkflow(t *testing.T) {
	t.Parallel()

	db := setupExecutionTestDB(t)
	repo := createExecutionRepo(db)
	ctx := context.Background()

	// Simulate a complete test execution workflow
	planID := shared.NewID()

	// 1. Create execution
	testExec := createTestExecution(planID, "Integration Test Execution")
	err := repo.Create(ctx, testExec)
	require.NoError(t, err)

	// 2. Batch submit results
	testCases := make([]shared.ID, 10)
	for i := range testCases {
		testCases[i] = shared.NewID()
	}

	results := make([]*execution.ExecutionResult, len(testCases))
	for i, tcID := range testCases {
		status := execution.ResultPassed
		if i == 2 || i == 5 {
			status = execution.ResultFailed
		} else if i == 7 {
			status = execution.ResultBlocked
		}
		results[i] = createTestResult(testExec.ID, tcID, status)
	}

	err = repo.BatchSubmitResults(ctx, results)
	require.NoError(t, err)

	// 3. Get summary
	summary, err := repo.GetSummary(ctx, testExec.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, summary.Total)
	assert.Equal(t, 7, summary.Passed)
	assert.Equal(t, 2, summary.Failed)
	assert.Equal(t, 1, summary.Blocked)
	assert.Equal(t, 0, summary.Skipped)

	// 4. Verify execution still exists
	found, err := repo.FindByID(ctx, testExec.ID)
	assert.NoError(t, err)
	assert.Equal(t, testExec.ID, found.ID)
	assert.Equal(t, planID, found.PlanID)
}

// Note: ExecutionRepository implementation is now in execution_repo.go
// This test file only contains test cases and helper functions
