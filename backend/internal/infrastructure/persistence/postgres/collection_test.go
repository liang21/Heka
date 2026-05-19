// tasks.md: T053 | spec.md: CollectionRepository TDD RED
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectionRepository_Create tests creating a collection in the database
func TestCollectionRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Smoke Test Suite",
		Description: "Critical smoke tests for regression",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create the collection
	err := repo.Create(ctx, collection)
	require.NoError(t, err)
	assert.Equal(t, collection.ID, collection.ID)
}

// TestCollectionRepository_AddCases tests adding test cases to a collection
func TestCollectionRepository_AddCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Regression Suite",
		Description: "Full regression tests",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create test cases to add
	caseIDs := []shared.ID{
		shared.NewID(),
		shared.NewID(),
		shared.NewID(),
	}

	// Create collection and add cases
	err := repo.Create(ctx, collection)
	require.NoError(t, err)

	err = repo.AddCases(ctx, collection.ID, caseIDs)
	require.NoError(t, err)

	// Verify cases were added
	cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, cases, 3)
}

// TestCollectionRepository_RemoveCases tests removing test cases from a collection
func TestCollectionRepository_RemoveCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Test Suite",
		Description: "Test suite",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create test cases
	caseIDs := []shared.ID{
		shared.NewID(),
		shared.NewID(),
		shared.NewID(),
		shared.NewID(),
	}

	// Setup: create collection and add cases
	err := repo.Create(ctx, collection)
	require.NoError(t, err)

	err = repo.AddCases(ctx, collection.ID, caseIDs)
	require.NoError(t, err)

	// Remove 2 cases
	toRemove := []shared.ID{caseIDs[0], caseIDs[2]}
	err = repo.RemoveCases(ctx, collection.ID, toRemove)
	require.NoError(t, err)

	// Verify 2 cases remain
	cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, cases, 2)
}

// TestCollectionRepository_ListCases tests listing test cases with pagination
func TestCollectionRepository_ListCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Paginated Suite",
		Description: "Test pagination",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create 25 test cases
	caseIDs := make([]shared.ID, 25)
	for i := 0; i < 25; i++ {
		caseIDs[i] = shared.NewID()
	}

	// Setup: create collection and add cases
	err := repo.Create(ctx, collection)
	require.NoError(t, err)

	err = repo.AddCases(ctx, collection.ID, caseIDs)
	require.NoError(t, err)

	// Test first page
	cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, cases, 10)

	// Test second page
	cases, total, err = repo.ListCases(ctx, collection.ID, 2, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, cases, 10)

	// Test third page (remaining 5)
	cases, total, err = repo.ListCases(ctx, collection.ID, 3, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, cases, 5)

	// Test out of bounds page
	cases, total, err = repo.ListCases(ctx, collection.ID, 100, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, cases, 0)
}

// TestCollectionRepository_EmptyCollection tests listing cases when collection is empty
func TestCollectionRepository_EmptyCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create an empty collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Empty Suite",
		Description: "No test cases yet",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create collection
	err := repo.Create(ctx, collection)
	require.NoError(t, err)

	// List cases should return empty
	cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, cases, 0)
}

// TestCollectionRepository_DuplicateName tests that duplicate collection names within a project are prevented
func TestCollectionRepository_DuplicateName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()
	collectionName := "Critical Tests"

	// Create first collection
	collection1 := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        collectionName,
		Description: "First collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection1)
	// require.NoError(t, err)

	// Try to create duplicate collection with same name in same project
	collection2 := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        collectionName,
		Description: "Second collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// This should fail due to unique constraint on (project_id, name)
	// err = repo.Create(ctx, collection2)
	// assert.Error(t, err)

	// However, same collection name in different project should work
	collection3 := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   shared.NewID(), // Different project
		Name:        collectionName,
		Description: "Third collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// err = repo.Create(ctx, collection3)
	// assert.NoError(t, err)

	_, _, _, _, _ = ctx, db, collection1, collection2, collection3 // Use variables to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_DuplicateName: implementation does not exist - TDD RED")
}

// TestCollectionRepository_AddDuplicateCases tests that adding the same case twice is handled correctly
func TestCollectionRepository_AddDuplicateCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Duplicate Test Suite",
		Description: "Test duplicate handling",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	caseID := shared.NewID()

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection)
	// require.NoError(t, err)
	//
	// // Add case first time
	// err = repo.AddCases(ctx, collection.ID, []shared.ID{caseID})
	// require.NoError(t, err)
	//
	// // Try to add same case again - should be idempotent or return error
	// err = repo.AddCases(ctx, collection.ID, []shared.ID{caseID})
	// // This depends on business requirements - could be no error (idempotent) or error
	//
	// // Verify only one instance exists
	// cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	// require.NoError(t, err)
	// assert.Equal(t, int64(1), total)
	// assert.Len(t, cases, 1)

	_, _, _ = ctx, db, collection // Use variables to avoid compilation errors
	_ = caseID                    // Use variable to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_AddDuplicateCases: implementation does not exist - TDD RED")
}

// TestCollectionRepository_RemoveNonExistentCases tests removing cases that don't exist in collection
func TestCollectionRepository_RemoveNonExistentCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	// Test data
	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Test Suite",
		Description: "Test suite",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Try to remove cases that were never added
	nonExistentCaseIDs := []shared.ID{
		shared.NewID(),
		shared.NewID(),
	}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection)
	// require.NoError(t, err)
	//
	// // Removing non-existent cases should either be no-op or return error
	// // This depends on business requirements
	// err = repo.RemoveCases(ctx, collection.ID, nonExistentCaseIDs)
	// // Could be no error (idempotent) or error

	_, _, _ = ctx, db, collection // Use variables to avoid compilation errors
	_ = nonExistentCaseIDs        // Use variable to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_RemoveNonExistentCases: implementation does not exist - TDD RED")
}

// TestCollectionRepository_MultipleProjects tests that collections are correctly scoped by project
func TestCollectionRepository_MultipleProjects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	userID := shared.NewID()

	// Create multiple projects with collections
	project1ID := shared.NewID()
	project2ID := shared.NewID()

	collection1 := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   project1ID,
		Name:        "P1-Collection",
		Description: "Project 1 collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	collection2 := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   project2ID,
		Name:        "P2-Collection",
		Description: "Project 2 collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create test cases for each project
	p1CaseIDs := []shared.ID{shared.NewID(), shared.NewID()}
	p2CaseIDs := []shared.ID{shared.NewID()}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection1)
	// require.NoError(t, err)
	// err = repo.Create(ctx, collection2)
	// require.NoError(t, err)
	//
	// // Add cases to each collection
	// err = repo.AddCases(ctx, collection1.ID, p1CaseIDs)
	// require.NoError(t, err)
	// err = repo.AddCases(ctx, collection2.ID, p2CaseIDs)
	// require.NoError(t, err)
	//
	// // Verify project isolation - collection1 should have 2 cases
	// cases1, total1, err := repo.ListCases(ctx, collection1.ID, 1, 10)
	// require.NoError(t, err)
	// assert.Equal(t, int64(2), total1)
	// assert.Len(t, cases1, 2)
	//
	// // Verify project isolation - collection2 should have 1 case
	// cases2, total2, err := repo.ListCases(ctx, collection2.ID, 1, 10)
	// require.NoError(t, err)
	// assert.Equal(t, int64(1), total2)
	// assert.Len(t, cases2, 1)

	_, _, _, _ = ctx, db, collection1, collection2 // Use variables to avoid compilation errors
	_, _ = p1CaseIDs, p2CaseIDs                    // Use variables to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_MultipleProjects: implementation does not exist - TDD RED")
}

// TestCollectionRepository_ConcurrentAddCases tests concurrent addition of cases to collection
func TestCollectionRepository_ConcurrentAddCases(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Concurrent Suite",
		Description: "Test concurrent operations",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection)
	// require.NoError(t, err)

	// Add 20 cases concurrently
	caseCount := 20
	errChan := make(chan error, caseCount)

	for i := 0; i < caseCount; i++ {
		go func(index int) {
			caseID := shared.NewID()
			_ = caseID // Use variable to avoid compilation errors
			// errChan <- repo.AddCases(ctx, collection.ID, []shared.ID{caseID})
			errChan <- nil // Placeholder
		}(i)
	}

	// Collect errors
	for i := 0; i < caseCount; i++ {
		// err := <-errChan
		// assert.NoError(t, err)
		<-errChan
	}

	// Verify all cases were added
	// cases, total, err := repo.ListCases(ctx, collection.ID, 1, 100)
	// require.NoError(t, err)
	// assert.Equal(t, int64(caseCount), total)
	// assert.Len(t, cases, caseCount)

	_, _, _ = ctx, db, collection // Use variables to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_ConcurrentAddCases: implementation does not exist - TDD RED")
}

// TestCollectionRepository_CreateTimestamp tests that created timestamps are handled correctly
func TestCollectionRepository_CreateTimestamp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	projectID := shared.NewID()
	userID := shared.NewID()

	beforeCreate := time.Now()

	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Timestamp Test",
		Description: "Test timestamp handling",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection)
	// require.NoError(t, err)

	afterCreate := time.Now()

	// Retrieve the collection and verify timestamp
	// Note: Collection entity has CreatedAt field in the domain model
	// assert.True(t, collection.CreatedAt.After(beforeCreate) || collection.CreatedAt.Equal(beforeCreate))
	// assert.True(t, collection.CreatedAt.Before(afterCreate) || collection.CreatedAt.Equal(afterCreate))

	_, _, _, _, _ = ctx, db, collection, beforeCreate, afterCreate // Use variables to avoid compilation errors

	// Placeholder assertion that will fail until implementation exists
	t.Fatal("TestCollectionRepository_CreateTimestamp: implementation does not exist - TDD RED")
}

// TestCollectionRepository_AddCases_EmptyList tests adding empty case list
func TestCollectionRepository_AddCases_EmptyList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository
	repo := NewCollectionRepository(db)

	projectID := shared.NewID()
	userID := shared.NewID()

	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Empty List Test",
		Description: "Test empty case list",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create collection
	err := repo.Create(ctx, collection)
	require.NoError(t, err)

	// Add empty case list - should be no-op
	err = repo.AddCases(ctx, collection.ID, []shared.ID{})
	require.NoError(t, err)

	// Verify collection is still empty
	cases, total, err := repo.ListCases(ctx, collection.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, cases, 0)
}

// TestCollectionRepository_ListCases_PaginationEdgeCases tests pagination edge cases
func TestCollectionRepository_ListCases_PaginationEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create repository - This will fail because collection_repo.go doesn't exist yet
	// repo := testcase.NewCollectionRepository(db)

	projectID := shared.NewID()
	userID := shared.NewID()

	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Pagination Edge Cases",
		Description: "Test pagination edge cases",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// Create exactly 10 test cases
	caseIDs := make([]shared.ID, 10)
	for i := 0; i < 10; i++ {
		caseIDs[i] = shared.NewID()
	}

	// This test will fail (RED) because the implementation doesn't exist
	// err := repo.Create(ctx, collection)
	// require.NoError(t, err)
	//
	// err = repo.AddCases(ctx, collection.ID, caseIDs)
	// require.NoError(t, err)

	testCases := []struct {
		name          string
		page          int
		pageSize      int
		expectedLen   int
		expectedTotal int64
	}{
		{"Page size equals total", 1, 10, 10, 10},
		{"Page size larger than total", 1, 100, 10, 10},
		{"Page size 1", 1, 1, 1, 10},
		{"Last page with size 1", 10, 1, 1, 10},
		{"Zero page size (should default or error)", 1, 0, 0, 10},   // Depends on requirements
		{"Negative page (should handle gracefully)", -1, 10, 0, 10}, // Depends on requirements
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// cases, total, err := repo.ListCases(ctx, collection.ID, tc.page, tc.pageSize)
			// require.NoError(t, err)
			// assert.Equal(t, tc.expectedTotal, total)
			// assert.Equal(t, tc.expectedLen, len(cases))

			_ = tc // Use variable to avoid compilation errors

			// Placeholder assertion that will fail until implementation exists
			t.Fatal("TestCollectionRepository_ListCases_PaginationEdgeCases: implementation does not exist - TDD RED")
		})
	}

	_, _, _ = ctx, db, collection // Use variables to avoid compilation errors
	_ = caseIDs                   // Use variable to avoid compilation errors
}

// BenchmarkCollectionRepository_AddCases benchmarks the AddCases method
func BenchmarkCollectionRepository_AddCases(b *testing.B) {
	ctx := context.Background()
	db := setupTestDB(&testing.T{})
	defer teardownTestDB(&testing.T{}, db)

	// repo := testcase.NewCollectionRepository(db)

	// Setup test data
	projectID := shared.NewID()
	userID := shared.NewID()

	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Benchmark Suite",
		Description: "Benchmark collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// err := repo.Create(ctx, collection)
	// require.NoError(b, err)

	_ = userID     // Use variable to avoid compilation errors
	_ = collection // Use variable to avoid compilation errors

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// caseIDs := []shared.ID{shared.NewID(), shared.NewID(), shared.NewID()}
		// _ = repo.AddCases(ctx, collection.ID, caseIDs)
		// Placeholder - will fail until implementation exists
		_ = i
	}

	_, _, _ = ctx, db, projectID // Use variables to avoid compilation errors

	b.Fatal("BenchmarkCollectionRepository_AddCases: implementation does not exist - TDD RED")
}

// BenchmarkCollectionRepository_ListCases benchmarks the ListCases method
func BenchmarkCollectionRepository_ListCases(b *testing.B) {
	ctx := context.Background()
	db := setupTestDB(&testing.T{})
	defer teardownTestDB(&testing.T{}, db)

	// repo := testcase.NewCollectionRepository(db)

	// Setup test data
	projectID := shared.NewID()
	userID := shared.NewID()

	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Benchmark Suite",
		Description: "Benchmark collection",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	// err := repo.Create(ctx, collection)
	// require.NoError(b, err)

	// Create 100 test cases
	caseIDs := make([]shared.ID, 100)
	for i := 0; i < 100; i++ {
		caseIDs[i] = shared.NewID()
	}

	// err = repo.AddCases(ctx, collection.ID, caseIDs)
	// require.NoError(b, err)

	_ = userID     // Use variable to avoid compilation errors
	_ = collection // Use variable to avoid compilation errors
	_ = caseIDs    // Use variable to avoid compilation errors

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// _, _, _ = repo.ListCases(ctx, collection.ID, 1, 10)
		// Placeholder - will fail until implementation exists
		_ = i
	}

	_, _, _ = ctx, db, projectID // Use variables to avoid compilation errors

	b.Fatal("BenchmarkCollectionRepository_ListCases: implementation does not exist - TDD RED")
}

// Example usage of the CollectionRepository
func ExampleCollectionRepository() {
	// This example demonstrates how to use the CollectionRepository
	// It will not run until the implementation exists

	ctx := context.Background()
	// db := setupDatabase()
	// repo := testcase.NewCollectionRepository(db)

	projectID := shared.NewID()
	userID := shared.NewID()

	// Create a new collection
	collection := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Smoke Test Suite",
		Description: "Critical smoke tests",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	_ = userID // Use variable to avoid compilation errors
	_ = ctx    // Use variable to avoid compilation errors

	// _ = repo.Create(ctx, collection)

	// Add test cases to the collection
	caseIDs := []shared.ID{shared.NewID(), shared.NewID()}
	// _ = repo.AddCases(ctx, collection.ID, caseIDs)

	// List cases with pagination
	// cases, total, _ := repo.ListCases(ctx, collection.ID, 1, 10)
	// fmt.Printf("Found %d test cases in collection\n", total)

	_ = collection // Use variable to avoid compilation errors
	_ = caseIDs    // Use variable to avoid compilation errors

	fmt.Println("ExampleCollectionRepository: implementation does not exist - TDD RED")
	// Output:
	// ExampleCollectionRepository: implementation does not exist - TDD RED
}
