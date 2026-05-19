// tasks.md: T062a | TDD RED phase - AsyncTaskRepository implementation tests
package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTaskTestDB creates a test database connection using testcontainers-go
// This function should be replaced with actual testcontainers setup once added to go.mod
func setupTaskTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// TODO: Replace with testcontainers-go setup
	dsn := "host=localhost port=5432 user=test_user password=test_pass dbname=test_db sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
	}

	// Auto-migrate the AsyncTask and IndexTask schemas
	err = db.AutoMigrate(&shared.AsyncTask{}, &shared.IndexTask{})
	if err != nil {
		t.Fatalf("Failed to migrate task schemas: %v", err)
	}

	// Clean tables before each test
	db.Exec("DELETE FROM index_tasks")
	db.Exec("DELETE FROM async_tasks")

	return db
}

// createAsyncTaskRepo creates a new AsyncTaskRepository instance for testing
func createAsyncTaskRepo(db *gorm.DB) *AsyncTaskRepository {
	// This will fail to compile until AsyncTaskRepository is implemented
	return &AsyncTaskRepository{db: db}
}

// createIndexTaskRepo creates a new IndexTaskRepository instance for testing
func createIndexTaskRepo(db *gorm.DB) *IndexTaskRepository {
	// This will fail to compile until IndexTaskRepository is implemented
	return &IndexTaskRepository{db: db}
}

func TestAsyncTaskRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	taskID := shared.NewID()
	projectID := shared.NewID()
	createdBy := shared.NewID()

	input := json.RawMessage(`{"file_id": "test-file-123"}`)

	testTask := &shared.AsyncTask{
		ID:        taskID,
		ProjectID: projectID,
		Type:      "file_indexing",
		Status:    "pending",
		Input:     input,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, testTask)
	assert.NoError(t, err, "Create should succeed")

	// Verify task was actually created
	found, err := repo.FindByID(ctx, testTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, testTask.ID, found.ID)
	assert.Equal(t, testTask.Type, found.Type)
	assert.Equal(t, testTask.Status, found.Status)
	assert.Equal(t, testTask.ProjectID, found.ProjectID)
}

func TestAsyncTaskRepository_Create_EmptyID(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	invalidTask := &shared.AsyncTask{
		ID:        "", // Empty ID
		ProjectID: shared.NewID(),
		Type:      "test",
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, invalidTask)
	assert.Error(t, err, "Create with empty ID should fail")
}

func TestAsyncTaskRepository_FindByID(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	taskID := shared.NewID()
	projectID := shared.NewID()
	createdBy := shared.NewID()

	testTask := &shared.AsyncTask{
		ID:        taskID,
		ProjectID: projectID,
		Type:      "test_execution",
		Status:    "completed",
		Result:    json.RawMessage(`{"success": true}`),
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, testTask)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, testTask.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testTask.ID, found.ID)
	assert.Equal(t, testTask.Type, found.Type)
	assert.Equal(t, testTask.Status, found.Status)
	assert.Equal(t, testTask.ProjectID, found.ProjectID)
}

func TestAsyncTaskRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, shared.NewID())
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestAsyncTaskRepository_FindPendingByType(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Create tasks with different statuses
	tasks := []*shared.AsyncTask{
		{
			ID:        shared.NewID(),
			ProjectID: projectID,
			Type:      "file_indexing",
			Status:    "pending",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		},
		{
			ID:        shared.NewID(),
			ProjectID: projectID,
			Type:      "file_indexing",
			Status:    "pending",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		},
		{
			ID:        shared.NewID(),
			ProjectID: projectID,
			Type:      "file_indexing",
			Status:    "processing", // Not pending
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		},
		{
			ID:        shared.NewID(),
			ProjectID: projectID,
			Type:      "test_execution", // Different type
			Status:    "pending",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		},
	}

	for _, task := range tasks {
		err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Find pending tasks of type "file_indexing"
	pending, err := repo.FindPendingByType(ctx, projectID, "file_indexing", 10)
	assert.NoError(t, err)
	assert.Len(t, pending, 2, "Should find 2 pending file_indexing tasks")

	for _, task := range pending {
		assert.Equal(t, "file_indexing", task.Type)
		assert.Equal(t, "pending", task.Status)
	}
}

func TestAsyncTaskRepository_FindPendingByType_WithLimit(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Create more pending tasks than the limit
	for i := 0; i < 10; i++ {
		task := &shared.AsyncTask{
			ID:        shared.NewID(),
			ProjectID: projectID,
			Type:      "batch_processing",
			Status:    "pending",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Request only 5 tasks
	pending, err := repo.FindPendingByType(ctx, projectID, "batch_processing", 5)
	assert.NoError(t, err)
	assert.Len(t, pending, 5, "Should return only 5 tasks as requested")
}

func TestAsyncTaskRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	taskID := shared.NewID()
	projectID := shared.NewID()
	createdBy := shared.NewID()

	testTask := &shared.AsyncTask{
		ID:        taskID,
		ProjectID: projectID,
		Type:      "test_execution",
		Status:    "pending",
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, testTask)
	require.NoError(t, err)

	// Update task status and add result
	now := time.Now()
	testTask.Status = "completed"
	testTask.ProgressCurrent = 100
	testTask.ProgressTotal = 100
	testTask.Result = json.RawMessage(`{"tests_passed": 42}`)
	testTask.CompletedAt = &now

	err = repo.Update(ctx, testTask)
	assert.NoError(t, err, "Update should succeed")

	// Verify changes
	updated, err := repo.FindByID(ctx, testTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", updated.Status)
	assert.Equal(t, 100, updated.ProgressCurrent)
	assert.Equal(t, 100, updated.ProgressTotal)
	assert.NotNil(t, updated.Result)
	assert.NotNil(t, updated.CompletedAt)
}

func TestAsyncTaskRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createAsyncTaskRepo(db)
	ctx := context.Background()

	nonExistentTask := &shared.AsyncTask{
		ID:        shared.NewID(),
		ProjectID: shared.NewID(),
		Type:      "ghost",
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	err := repo.Update(ctx, nonExistentTask)
	assert.Error(t, err, "Update non-existent task should fail")
}

func TestIndexTaskRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	fileID := shared.NewID()

	testTask := &shared.IndexTask{
		ID:         shared.NewID(),
		FileID:     fileID,
		Status:     "pending",
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, testTask)
	assert.NoError(t, err, "Create should succeed")

	// Verify task was created
	pending, err := repo.FindPending(ctx, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(pending), 1)

	// Find the created task
	var found *shared.IndexTask
	for _, task := range pending {
		if task.ID == testTask.ID {
			found = task
			break
		}
	}

	assert.NotNil(t, found)
	assert.Equal(t, testTask.FileID, found.FileID)
	assert.Equal(t, testTask.Status, found.Status)
	assert.Equal(t, testTask.MaxRetries, found.MaxRetries)
}

func TestIndexTaskRepository_FindPending(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	// Create multiple index tasks with different statuses
	now := time.Now()
	tasks := []*shared.IndexTask{
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-2 * time.Hour),
		},
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 1,
			MaxRetries: 3,
			CreatedAt:  now.Add(-1 * time.Hour),
			UpdatedAt:  now.Add(-30 * time.Minute),
		},
		{
			ID:          shared.NewID(),
			FileID:      shared.NewID(),
			Status:      "completed", // Not pending
			RetryCount:  0,
			MaxRetries:  3,
			CreatedAt:   now.Add(-3 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
			CompletedAt: &now,
		},
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "failed", // Not pending
			RetryCount: 3,
			MaxRetries: 3,
			Error:      "max retries exceeded",
			CreatedAt:  now.Add(-4 * time.Hour),
			UpdatedAt:  now.Add(-1 * time.Hour),
		},
	}

	for _, task := range tasks {
		err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Find pending tasks
	pending, err := repo.FindPending(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, pending, 2, "Should find 2 pending index tasks")

	for _, task := range pending {
		assert.Equal(t, "pending", task.Status)
	}
}

func TestIndexTaskRepository_FindStale(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	// Create tasks with different ages
	now := time.Now()
	oldTime := now.Add(-3 * time.Hour)
	recentTime := now.Add(-10 * time.Minute)

	tasks := []*shared.IndexTask{
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  oldTime,
			UpdatedAt:  oldTime,
		},
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 1,
			MaxRetries: 3,
			CreatedAt:  oldTime.Add(-30 * time.Minute),
			UpdatedAt:  oldTime.Add(-30 * time.Minute),
		},
		{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  recentTime, // Too recent
			UpdatedAt:  recentTime,
		},
		{
			ID:          shared.NewID(),
			FileID:      shared.NewID(),
			Status:      "completed", // Not pending
			RetryCount:  0,
			MaxRetries:  3,
			CreatedAt:   oldTime,
			UpdatedAt:   oldTime,
			CompletedAt: &now,
		},
	}

	for _, task := range tasks {
		err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Find stale tasks (older than 1 hour)
	staleTime := "1h"
	stale, err := repo.FindStale(ctx, staleTime, 10)
	assert.NoError(t, err)
	assert.Len(t, stale, 2, "Should find 2 stale pending tasks")

	for _, task := range stale {
		assert.Equal(t, "pending", task.Status)
		assert.True(t, task.UpdatedAt.Before(now.Add(-1*time.Hour)))
	}
}

func TestIndexTaskRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	testTask := &shared.IndexTask{
		ID:         shared.NewID(),
		FileID:     shared.NewID(),
		Status:     "pending",
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, testTask)
	require.NoError(t, err)

	// Update task status and retry count
	testTask.Status = "processing"
	testTask.RetryCount = 1
	testTask.UpdatedAt = time.Now()

	err = repo.Update(ctx, testTask)
	assert.NoError(t, err, "Update should succeed")

	// Verify changes
	pending, err := repo.FindPending(ctx, 10)
	assert.NoError(t, err)

	// Task should no longer be in pending
	for _, task := range pending {
		if task.ID == testTask.ID {
			t.Errorf("Updated task should not be in pending list")
		}
	}
}

func TestIndexTaskRepository_Update_RetryCountExceeded(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	testTask := &shared.IndexTask{
		ID:         shared.NewID(),
		FileID:     shared.NewID(),
		Status:     "pending",
		RetryCount: 2, // One more try will exceed max
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, testTask)
	require.NoError(t, err)

	// Simulate failed retry
	testTask.RetryCount = 3
	testTask.Status = "failed"
	testTask.Error = "max retries exceeded"
	testTask.UpdatedAt = time.Now()

	err = repo.Update(ctx, testTask)
	assert.NoError(t, err)

	// Verify task is marked as failed
	pending, err := repo.FindPending(ctx, 10)
	assert.NoError(t, err)

	for _, task := range pending {
		assert.NotEqual(t, testTask.ID, task.ID, "Failed task should not be in pending list")
	}
}

func TestIndexTaskRepository_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	db := setupTaskTestDB(t)
	repo := createIndexTaskRepo(db)
	ctx := context.Background()

	// Create multiple tasks concurrently
	const numTasks = 5
	tasks := make([]*shared.IndexTask, numTasks)

	for i := 0; i < numTasks; i++ {
		tasks[i] = &shared.IndexTask{
			ID:         shared.NewID(),
			FileID:     shared.NewID(),
			Status:     "pending",
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err := repo.Create(ctx, tasks[i])
		require.NoError(t, err)
	}

	// Verify all tasks can be found
	pending, err := repo.FindPending(ctx, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(pending), numTasks)
}
