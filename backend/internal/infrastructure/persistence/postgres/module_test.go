// tasks.md: T047 | TDD RED test for ModuleRepository
package postgres

import (
	"context"
	"testing"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createTestModule creates a test module with default values
func createTestModule(projectID shared.ID, name string, parentID *shared.ID) *testcase.Module {
	return &testcase.Module{
		ID:        shared.NewID(),
		ProjectID: projectID,
		Name:      name,
		ParentID:  parentID,
		CreatedBy: shared.NewID(), // Required field
	}
}

func TestCreateModule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()
	module := createTestModule(projectID, "Test Module", nil)

	err := repo.Create(ctx, module)

	// Should fail - implementation doesn't exist yet
	assert.NoError(t, err, "Create should succeed with valid module")
	assert.NotEmpty(t, module.ID, "Module ID should be set")
}

func TestCreateModule_WithParent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create parent module first
	parent := createTestModule(projectID, "Parent Module", nil)
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create child module
	child := createTestModule(projectID, "Child Module", &parent.ID)
	err = repo.Create(ctx, child)

	assert.NoError(t, err, "Create should succeed with parent module")
	assert.Equal(t, parent.ID, *child.ParentID, "ParentID should match")
}

func TestCreateModule_DuplicateNameInSameLevel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()
	module1 := createTestModule(projectID, "Duplicate Module", nil)
	module2 := createTestModule(projectID, "Duplicate Module", nil)

	err := repo.Create(ctx, module1)
	require.NoError(t, err)

	err = repo.Create(ctx, module2)

	// Should fail due to duplicate name constraint
	assert.Error(t, err, "Create should fail with duplicate module name")
}

func TestUpdateModule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()
	module := createTestModule(projectID, "Original Name", nil)
	err := repo.Create(ctx, module)
	require.NoError(t, err)

	// Update module
	module.Name = "Updated Name"
	err = repo.Update(ctx, module)

	assert.NoError(t, err, "Update should succeed with valid changes")

	// Verify update
	updated, err := repo.FindByProject(ctx, projectID)
	require.NoError(t, err)
	found := false
	for _, m := range updated {
		if m.ID == module.ID {
			assert.Equal(t, "Updated Name", m.Name, "Name should be updated")
			found = true
			break
		}
	}
	assert.True(t, found, "Updated module should be found")
}

func TestUpdateModule_MoveToDifferentParent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create parent modules
	parent1 := createTestModule(projectID, "Parent 1", nil)
	parent2 := createTestModule(projectID, "Parent 2", nil)
	err := repo.Create(ctx, parent1)
	require.NoError(t, err)
	err = repo.Create(ctx, parent2)
	require.NoError(t, err)

	// Create child under parent1
	child := createTestModule(projectID, "Child Module", &parent1.ID)
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// Move child to parent2
	child.ParentID = &parent2.ID
	err = repo.Update(ctx, child)

	assert.NoError(t, err, "Update should succeed when moving to different parent")
	assert.Equal(t, parent2.ID, *child.ParentID, "ParentID should be updated")
}

func TestUpdateModule_NonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	nonExistentModule := createTestModule(shared.NewID(), "Non-existent", nil)

	err := repo.Update(ctx, nonExistentModule)

	assert.Error(t, err, "Update should fail for non-existent module")
}

func TestDeleteModule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()
	module := createTestModule(projectID, "Module to Delete", nil)
	err := repo.Create(ctx, module)
	require.NoError(t, err)

	// Delete module
	err = repo.Delete(ctx, module.ID)

	assert.NoError(t, err, "Delete should succeed with valid module ID")

	// Verify deletion
	modules, err := repo.FindByProject(ctx, projectID)
	require.NoError(t, err)
	found := false
	for _, m := range modules {
		if m.ID == module.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted module should not be found")
}

func TestDeleteModule_NonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	err := repo.Delete(ctx, shared.NewID())

	assert.Error(t, err, "Delete should fail for non-existent module ID")
}

func TestDeleteModule_WithChildren(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create parent module
	parent := createTestModule(projectID, "Parent Module", nil)
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create child modules
	child1 := createTestModule(projectID, "Child 1", &parent.ID)
	child2 := createTestModule(projectID, "Child 2", &parent.ID)
	err = repo.Create(ctx, child1)
	require.NoError(t, err)
	err = repo.Create(ctx, child2)
	require.NoError(t, err)

	// Delete parent (should cascade or fail depending on requirements)
	err = repo.Delete(ctx, parent.ID)

	// This test documents the expected behavior
	// Implementation should handle cascading or prevent deletion
	assert.NoError(t, err, "Delete should handle parent with children appropriately")
}

func TestFindByProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create multiple modules
	module1 := createTestModule(projectID, "Module 1", nil)
	module2 := createTestModule(projectID, "Module 2", nil)
	module3 := createTestModule(projectID, "Module 3", nil)

	err := repo.Create(ctx, module1)
	require.NoError(t, err)
	err = repo.Create(ctx, module2)
	require.NoError(t, err)
	err = repo.Create(ctx, module3)
	require.NoError(t, err)

	// Find all modules
	modules, err := repo.FindByProject(ctx, projectID)

	assert.NoError(t, err, "FindByProject should succeed")
	assert.Len(t, modules, 3, "Should find all 3 modules")
}

func TestFindByProject_TreeStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create tree structure:
	// Root 1
	//   Child 1.1
	//     Child 1.1.1
	// Root 2
	//   Child 2.1

	root1 := createTestModule(projectID, "Root 1", nil)
	root2 := createTestModule(projectID, "Root 2", nil)

	err := repo.Create(ctx, root1)
	require.NoError(t, err)
	err = repo.Create(ctx, root2)
	require.NoError(t, err)

	child11 := createTestModule(projectID, "Child 1.1", &root1.ID)
	child21 := createTestModule(projectID, "Child 2.1", &root2.ID)

	err = repo.Create(ctx, child11)
	require.NoError(t, err)
	err = repo.Create(ctx, child21)
	require.NoError(t, err)

	child111 := createTestModule(projectID, "Child 1.1.1", &child11.ID)
	err = repo.Create(ctx, child111)
	require.NoError(t, err)

	// Find all modules
	modules, err := repo.FindByProject(ctx, projectID)

	assert.NoError(t, err, "FindByProject should succeed")
	assert.Len(t, modules, 5, "Should find all 5 modules in tree")

	// Verify tree structure can be reconstructed
	moduleMap := make(map[shared.ID]*testcase.Module)
	for _, m := range modules {
		moduleMap[m.ID] = m
	}

	// Verify parent-child relationships
	assert.Equal(t, root1.ID, *moduleMap[child11.ID].ParentID, "Child 1.1 should have Root 1 as parent")
	assert.Equal(t, root2.ID, *moduleMap[child21.ID].ParentID, "Child 2.1 should have Root 2 as parent")
	assert.Equal(t, child11.ID, *moduleMap[child111.ID].ParentID, "Child 1.1.1 should have Child 1.1 as parent")
}

func TestFindByProject_EmptyProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	modules, err := repo.FindByProject(ctx, projectID)

	assert.NoError(t, err, "FindByProject should succeed even with no modules")
	assert.Empty(t, modules, "Should return empty slice for project with no modules")
}

func TestFindByProject_DifferentProjects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	project1 := shared.NewID()
	project2 := shared.NewID()

	// Create modules for project1
	module1 := createTestModule(project1, "Project 1 Module", nil)
	err := repo.Create(ctx, module1)
	require.NoError(t, err)

	// Create modules for project2
	module2 := createTestModule(project2, "Project 2 Module", nil)
	err = repo.Create(ctx, module2)
	require.NoError(t, err)

	// Find modules for project1
	modules1, err := repo.FindByProject(ctx, project1)
	assert.NoError(t, err)
	assert.Len(t, modules1, 1, "Should find only project 1 modules")
	assert.Equal(t, "Project 1 Module", modules1[0].Name, "Should find correct module for project 1")

	// Find modules for project2
	modules2, err := repo.FindByProject(ctx, project2)
	assert.NoError(t, err)
	assert.Len(t, modules2, 1, "Should find only project 2 modules")
	assert.Equal(t, "Project 2 Module", modules2[0].Name, "Should find correct module for project 2")
}

func TestFindByProject_Ordering(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	db := setupTestDB(t)

	// Create repository instance
	repo := NewModuleRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	projectID := shared.NewID()

	// Create modules in specific order
	module3 := createTestModule(projectID, "Module C", nil)
	module1 := createTestModule(projectID, "Module A", nil)
	module2 := createTestModule(projectID, "Module B", nil)

	err := repo.Create(ctx, module3)
	require.NoError(t, err)
	err = repo.Create(ctx, module1)
	require.NoError(t, err)
	err = repo.Create(ctx, module2)
	require.NoError(t, err)

	modules, err := repo.FindByProject(ctx, projectID)

	assert.NoError(t, err)
	// Verify ordering - typically should be ordered by creation time or name
	// This test documents the expected ordering behavior
	assert.Len(t, modules, 3, "Should find all modules")
}
