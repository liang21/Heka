//go:build plan_test
// +build plan_test

// tasks.md: T055 | TDD RED phase - TestPlanRepository implementation tests
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupPlanTestDB creates a test database connection using testcontainers-go
// This function should be replaced with actual testcontainers setup once configured
func setupPlanTestDB(t *testing.T) *gorm.DB {
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

	// Create tables manually using raw SQL since domain entities don't have GORM tags
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_plans (
			id VARCHAR(36) PRIMARY KEY,
			project_id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL,
			current_execution_id VARCHAR(36),
			started_at TIMESTAMP,
			paused_at TIMESTAMP,
			ended_at TIMESTAMP,
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS plan_test_cases (
			plan_id VARCHAR(36) NOT NULL,
			test_case_id VARCHAR(36) NOT NULL,
			assigned_to VARCHAR(36),
			order_index INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (plan_id, test_case_id)
		);
		CREATE INDEX IF NOT EXISTS idx_test_plans_project_id ON test_plans(project_id);
		CREATE INDEX IF NOT EXISTS idx_test_plans_status ON test_plans(status);
	`).Error
	if err != nil {
		t.Fatalf("Failed to create TestPlan tables: %v", err)
	}

	// Clean tables before each test
	db.Exec("DELETE FROM plan_test_cases")
	db.Exec("DELETE FROM test_plans")

	return db
}

// createPlanRepo creates a new PlanRepository instance for testing
func createPlanRepo(db *gorm.DB) *PlanRepository {
	// This will fail to compile until PlanRepository is implemented (TDD RED)
	return &PlanRepository{db: db}
}

// helper function to create a valid test plan
func createTestPlan(projectID, createdBy shared.ID, name string) *plan.TestPlan {
	now := time.Now()
	return &plan.TestPlan{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        name,
		Description: fmt.Sprintf("%s description", name),
		Status:      plan.PlanDraft,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestCreatePlan(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Test Plan 1")

	err := repo.Create(ctx, testPlan)
	assert.NoError(t, err, "Create should succeed")

	// Verify plan was actually created
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPlan.ID, found.ID)
	assert.Equal(t, testPlan.Name, found.Name)
	assert.Equal(t, testPlan.ProjectID, found.ProjectID)
	assert.Equal(t, testPlan.Status, found.Status)
	assert.Equal(t, testPlan.CreatedBy, found.CreatedBy)
}

func TestCreatePlan_WithCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Plan with Cases")

	// Create test cases to associate
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	case3ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
		{PlanID: testPlan.ID, TestCaseID: case3ID, AssignedTo: &createdBy, OrderIndex: 2},
	}

	err := repo.Create(ctx, testPlan)
	assert.NoError(t, err, "Create plan should succeed")

	err = repo.AddCases(ctx, testPlan.ID, cases)
	assert.NoError(t, err, "AddCases should succeed")

	// Verify cases were added
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 3, "Should have 3 associated cases")
}

func TestFindByID(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Find By ID Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testPlan.ID, found.ID)
	assert.Equal(t, testPlan.Name, found.Name)
	assert.Equal(t, testPlan.Description, found.Description)
	assert.Equal(t, testPlan.Status, found.Status)
	assert.Equal(t, testPlan.CreatedBy, found.CreatedBy)
}

func TestFindByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, shared.NewID())
	assert.Error(t, err, "FindByID with non-existent ID should return error")
	assert.Nil(t, found)
}

func TestFindByID_WithCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Plan with Test Cases")

	case1ID := shared.NewID()
	case2ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
	}

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 2)

	// Verify case details
	assert.Equal(t, case1ID, found.Cases[0].TestCaseID)
	assert.Equal(t, case2ID, found.Cases[1].TestCaseID)
	assert.Equal(t, 0, found.Cases[0].OrderIndex)
	assert.Equal(t, 1, found.Cases[1].OrderIndex)
	assert.NotNil(t, found.Cases[0].AssignedTo)
	assert.Nil(t, found.Cases[1].AssignedTo)
}

func TestList(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Create multiple plans
	for i := 1; i <= 5; i++ {
		testPlan := createTestPlan(projectID, createdBy, fmt.Sprintf("Plan %d", i))
		err := repo.Create(ctx, testPlan)
		require.NoError(t, err)
	}

	// Test listing all plans
	plans, total, err := repo.List(ctx, projectID, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, plans, 5, "Should return 5 plans")
	assert.Equal(t, int64(5), total, "Total count should be 5")
}

func TestList_WithPagination(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Create 15 plans
	for i := 1; i <= 15; i++ {
		testPlan := createTestPlan(projectID, createdBy, fmt.Sprintf("Plan %d", i))
		err := repo.Create(ctx, testPlan)
		require.NoError(t, err)
	}

	// Test first page
	page1, total1, err := repo.List(ctx, projectID, nil, 1, 5)
	assert.NoError(t, err)
	assert.Len(t, page1, 5, "First page should have 5 plans")
	assert.Equal(t, int64(15), total1, "Total count should be 15")

	// Test second page
	page2, total2, err := repo.List(ctx, projectID, nil, 2, 5)
	assert.NoError(t, err)
	assert.Len(t, page2, 5, "Second page should have 5 plans")
	assert.Equal(t, int64(15), total2, "Total count should still be 15")

	// Verify different plans on different pages
	page1IDs := make(map[shared.ID]bool)
	for _, p := range page1 {
		page1IDs[p.ID] = true
	}

	for _, p := range page2 {
		assert.NotContains(t, page1IDs, p.ID, "Second page should not contain plans from first page")
	}
}

func TestList_WithStatusFilter(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Create plans with different statuses
	draftPlan := createTestPlan(projectID, createdBy, "Draft Plan")
	draftPlan.Status = plan.PlanDraft
	err := repo.Create(ctx, draftPlan)
	require.NoError(t, err)

	activePlan := createTestPlan(projectID, createdBy, "Active Plan")
	activePlan.Status = plan.PlanActive
	err = repo.Create(ctx, activePlan)
	require.NoError(t, err)

	completedPlan := createTestPlan(projectID, createdBy, "Completed Plan")
	completedPlan.Status = plan.PlanCompleted
	err = repo.Create(ctx, completedPlan)
	require.NoError(t, err)

	// Filter by active status
	activeStatus := plan.PlanActive
	plans, total, err := repo.List(ctx, projectID, &activeStatus, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, plans, 1, "Should return only 1 active plan")
	assert.Equal(t, int64(1), total)
	assert.Equal(t, activePlan.ID, plans[0].ID)
	assert.Equal(t, plan.PlanActive, plans[0].Status)

	// Filter by draft status
	draftStatus := plan.PlanDraft
	plans, total, err = repo.List(ctx, projectID, &draftStatus, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, plans, 1, "Should return only 1 draft plan")
	assert.Equal(t, draftPlan.ID, plans[0].ID)
	assert.Equal(t, plan.PlanDraft, plans[0].Status)
}

func TestList_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	project1ID := shared.NewID()
	project2ID := shared.NewID()
	createdBy := shared.NewID()

	// Create plans for project 1
	for i := 1; i <= 3; i++ {
		testPlan := createTestPlan(project1ID, createdBy, fmt.Sprintf("Project1 Plan %d", i))
		err := repo.Create(ctx, testPlan)
		require.NoError(t, err)
	}

	// Create plans for project 2
	for i := 1; i <= 5; i++ {
		testPlan := createTestPlan(project2ID, createdBy, fmt.Sprintf("Project2 Plan %d", i))
		err := repo.Create(ctx, testPlan)
		require.NoError(t, err)
	}

	// List plans for project 1
	project1Plans, total1, err := repo.List(ctx, project1ID, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, project1Plans, 3, "Project 1 should have 3 plans")
	assert.Equal(t, int64(3), total1)

	// List plans for project 2
	project2Plans, total2, err := repo.List(ctx, project2ID, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, project2Plans, 5, "Project 2 should have 5 plans")
	assert.Equal(t, int64(5), total2)
}

func TestList_EmptyProject(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()

	plans, total, err := repo.List(ctx, projectID, nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, plans, 0, "Should return empty list for project with no plans")
	assert.Equal(t, int64(0), total)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Original Plan Name")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Update plan fields
	testPlan.Name = "Updated Plan Name"
	testPlan.Description = "Updated description"
	testPlan.Status = plan.PlanActive
	testPlan.UpdatedAt = time.Now()

	err = repo.Update(ctx, testPlan)
	assert.NoError(t, err, "Update should succeed")

	// Verify changes
	updated, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Plan Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, plan.PlanActive, updated.Status)
}

func TestUpdate_StatusTransitions(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Status Transition Plan")
	testPlan.Status = plan.PlanDraft

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Valid transition: draft → active
	testPlan.Status = plan.PlanActive
	testPlan.UpdatedAt = time.Now()
	err = repo.Update(ctx, testPlan)
	assert.NoError(t, err, "Draft to Active transition should succeed")

	// Valid transition: active → paused
	testPlan.Status = plan.PlanPaused
	testPlan.UpdatedAt = time.Now()
	err = repo.Update(ctx, testPlan)
	assert.NoError(t, err, "Active to Paused transition should succeed")

	// Valid transition: paused → active
	testPlan.Status = plan.PlanActive
	testPlan.UpdatedAt = time.Now()
	err = repo.Update(ctx, testPlan)
	assert.NoError(t, err, "Paused to Active transition should succeed")

	// Valid transition: active → completed
	testPlan.Status = plan.PlanCompleted
	testPlan.UpdatedAt = time.Now()
	err = repo.Update(ctx, testPlan)
	assert.NoError(t, err, "Active to Completed transition should succeed")
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	nonExistentPlan := createTestPlan(shared.NewID(), shared.NewID(), "Ghost Plan")

	err := repo.Update(ctx, nonExistentPlan)
	assert.Error(t, err, "Update non-existent plan should fail")
}

func TestAddCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Add Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add test cases
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	case3ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
		{PlanID: testPlan.ID, TestCaseID: case3ID, AssignedTo: &createdBy, OrderIndex: 2},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	assert.NoError(t, err, "AddCases should succeed")

	// Verify cases were added
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 3)
}

func TestAddCases_EmptyList(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Empty Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	err = repo.AddCases(ctx, testPlan.ID, []plan.PlanTestCase{})
	assert.NoError(t, err, "AddCases with empty list should succeed")

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 0)
}

func TestAddCases_PlanNotFound(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	createdBy := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: planID, TestCaseID: shared.NewID(), AssignedTo: &createdBy, OrderIndex: 0},
	}

	err := repo.AddCases(ctx, planID, cases)
	assert.Error(t, err, "AddCases to non-existent plan should fail")
}

func TestAddCases_DuplicateCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Duplicate Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add cases
	case1ID := shared.NewID()
	case2ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Try to add the same cases again (should either succeed idempotently or fail)
	err = repo.AddCases(ctx, testPlan.ID, cases)
	// This test verifies the behavior - implementation should handle duplicates
	assert.NoError(t, err, "AddCases with duplicate cases should handle gracefully")
}

func TestRemoveCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Remove Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add test cases
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	case3ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
		{PlanID: testPlan.ID, TestCaseID: case3ID, AssignedTo: &createdBy, OrderIndex: 2},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Remove some cases
	err = repo.RemoveCases(ctx, testPlan.ID, []shared.ID{case1ID, case3ID})
	assert.NoError(t, err, "RemoveCases should succeed")

	// Verify cases were removed
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 1, "Should have 1 remaining case")
	assert.Equal(t, case2ID, found.Cases[0].TestCaseID)
}

func TestRemoveCases_AllCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Remove All Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add test cases
	case1ID := shared.NewID()
	case2ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Remove all cases
	err = repo.RemoveCases(ctx, testPlan.ID, []shared.ID{case1ID, case2ID})
	assert.NoError(t, err, "RemoveCases should succeed")

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 0, "Should have no remaining cases")
}

func TestRemoveCases_EmptyList(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Empty Remove Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add cases first
	case1ID := shared.NewID()
	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Remove empty list (should succeed but do nothing)
	err = repo.RemoveCases(ctx, testPlan.ID, []shared.ID{})
	assert.NoError(t, err, "RemoveCases with empty list should succeed")

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 1, "Should still have 1 case")
}

func TestRemoveCases_NonExistentCases(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Non-existent Remove Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add a case
	case1ID := shared.NewID()
	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Try to remove cases that don't exist
	nonExistentID := shared.NewID()
	err = repo.RemoveCases(ctx, testPlan.ID, []shared.ID{nonExistentID})
	assert.NoError(t, err, "RemoveCases with non-existent case IDs should succeed")

	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 1, "Should still have 1 case")
}

func TestRemoveCases_PlanNotFound(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	planID := shared.NewID()
	case1ID := shared.NewID()

	err := repo.RemoveCases(ctx, planID, []shared.ID{case1ID})
	assert.Error(t, err, "RemoveCases from non-existent plan should fail")
}

func TestPlanRepository_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	// Test concurrent creates
	const numPlans = 10
	plans := make([]*plan.TestPlan, numPlans)

	for i := 0; i < numPlans; i++ {
		plans[i] = createTestPlan(projectID, createdBy, fmt.Sprintf("Concurrent Plan %d", i))
	}

	// Create all plans
	for _, p := range plans {
		err := repo.Create(ctx, p)
		assert.NoError(t, err)
	}

	// Verify all plans can be found
	for _, p := range plans {
		found, err := repo.FindByID(ctx, p.ID)
		assert.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)
	}

	// Verify list returns all plans
	allPlans, total, err := repo.List(ctx, projectID, nil, 1, 20)
	assert.NoError(t, err)
	assert.Len(t, allPlans, numPlans)
	assert.Equal(t, int64(numPlans), total)
}

func TestPlanRepository_CaseOrdering(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Ordered Cases Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add cases with specific order
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	case3ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case3ID, AssignedTo: &createdBy, OrderIndex: 2},
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: nil, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: &createdBy, OrderIndex: 1},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Verify cases are returned in correct order
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 3)

	// Check order is preserved
	assert.Equal(t, 0, found.Cases[0].OrderIndex)
	assert.Equal(t, 1, found.Cases[1].OrderIndex)
	assert.Equal(t, 2, found.Cases[2].OrderIndex)
}

func TestPlanRepository_AssignedToTracking(t *testing.T) {
	t.Parallel()

	db := setupPlanTestDB(t)
	repo := createPlanRepo(db)
	ctx := context.Background()

	projectID := shared.NewID()
	createdBy := shared.NewID()
	assignee1 := shared.NewID()
	assignee2 := shared.NewID()

	testPlan := createTestPlan(projectID, createdBy, "Assignment Tracking Plan")

	err := repo.Create(ctx, testPlan)
	require.NoError(t, err)

	// Add cases with different assignments
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	case3ID := shared.NewID()
	case4ID := shared.NewID()

	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &assignee1, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},           // Unassigned
		{PlanID: testPlan.ID, TestCaseID: case3ID, AssignedTo: &assignee2, OrderIndex: 2},
		{PlanID: testPlan.ID, TestCaseID: case4ID, AssignedTo: &assignee1, OrderIndex: 3},
	}

	err = repo.AddCases(ctx, testPlan.ID, cases)
	require.NoError(t, err)

	// Verify assignments are preserved
	found, err := repo.FindByID(ctx, testPlan.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Cases, 4)

	// Count assignments
	assignee1Count := 0
	assignee2Count := 0
	unassignedCount := 0

	for _, c := range found.Cases {
		if c.AssignedTo == nil {
			unassignedCount++
		} else if *c.AssignedTo == assignee1 {
			assignee1Count++
		} else if *c.AssignedTo == assignee2 {
			assignee2Count++
		}
	}

	assert.Equal(t, 2, assignee1Count, "Assignee1 should have 2 cases")
	assert.Equal(t, 1, assignee2Count, "Assignee2 should have 1 case")
	assert.Equal(t, 1, unassignedCount, "Should have 1 unassigned case")
}
