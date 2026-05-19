// tasks.md: T045 | TDD RED Phase for ProjectRepository
package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestCreateProject tests creating a new project in the database
func TestCreateProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Setup test database connection
	// This will fail until testcontainers-go is properly configured
	db, err := setupTestDB(ctx)
	require.NoError(t, err, "Failed to setup test database")

	// Create repository instance
	// This will fail until projectRepo is implemented
	repo := NewProjectRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	// Create test project
	testProject := &project.Project{
		ID:          shared.NewID(),
		Name:        "Test Project",
		Description: "A test project for TDD",
		CreatedBy:   shared.NewID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test Create
	err = repo.Create(ctx, testProject)
	assert.NoError(t, err, "Create should succeed")
	assert.False(t, testProject.ID.IsEmpty(), "Project ID should not be empty after creation")
}

// TestFindByUserID tests retrieving projects by user ID
func TestFindByUserID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := setupTestDB(ctx)
	require.NoError(t, err, "Failed to setup test database")

	repo := NewProjectRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	// Create test user and projects
	userID := shared.NewID()
	project1 := &project.Project{
		ID:          shared.NewID(),
		Name:        "User Project 1",
		Description: "First project",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	project2 := &project.Project{
		ID:          shared.NewID(),
		Name:        "User Project 2",
		Description: "Second project",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create projects
	err = repo.Create(ctx, project1)
	require.NoError(t, err, "Failed to create project1")
	err = repo.Create(ctx, project2)
	require.NoError(t, err, "Failed to create project2")

	// Test FindByUserID
	projects, err := repo.FindByUserID(ctx, userID)
	assert.NoError(t, err, "FindByUserID should succeed")
	assert.Len(t, projects, 2, "Should find 2 projects for user")
}

// TestIsMember tests checking if a user is a member of a project
func TestIsMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := setupTestDB(ctx)
	require.NoError(t, err, "Failed to setup test database")

	repo := NewProjectRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	// Create test project and member
	projectID := shared.NewID()
	userID := shared.NewID()

	testProject := &project.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "A test project",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, testProject)
	require.NoError(t, err, "Failed to create project")

	// Test IsMember before adding member
	isMember, err := repo.IsMember(ctx, projectID, userID)
	assert.NoError(t, err, "IsMember should succeed")
	assert.False(t, isMember, "User should not be a member yet")

	// Add member
	member := &project.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}
	err = repo.AddMember(ctx, member)
	require.NoError(t, err, "Failed to add member")

	// Test IsMember after adding member
	isMember, err = repo.IsMember(ctx, projectID, userID)
	assert.NoError(t, err, "IsMember should succeed")
	assert.True(t, isMember, "User should be a member after adding")
}

// TestAddMember tests adding a member to a project
func TestAddMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := setupTestDB(ctx)
	require.NoError(t, err, "Failed to setup test database")

	repo := NewProjectRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	// Create test project
	projectID := shared.NewID()
	userID := shared.NewID()

	testProject := &project.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "A test project",
		CreatedBy:   shared.NewID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, testProject)
	require.NoError(t, err, "Failed to create project")

	// Test AddMember
	member := &project.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}

	err = repo.AddMember(ctx, member)
	assert.NoError(t, err, "AddMember should succeed")

	// Verify member was added
	isMember, err := repo.IsMember(ctx, projectID, userID)
	assert.NoError(t, err, "IsMember should succeed")
	assert.True(t, isMember, "User should be a member after adding")
}

// TestCountMembers tests counting members in a project
func TestCountMembers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := setupTestDB(ctx)
	require.NoError(t, err, "Failed to setup test database")

	repo := NewProjectRepository(db)
	require.NotNil(t, repo, "Repository should not be nil")

	// Create test project
	projectID := shared.NewID()

	testProject := &project.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "A test project",
		CreatedBy:   shared.NewID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, testProject)
	require.NoError(t, err, "Failed to create project")

	// Test CountMembers with no members
	count, err := repo.CountMembers(ctx, projectID)
	assert.NoError(t, err, "CountMembers should succeed")
	assert.Equal(t, int64(0), count, "Should have 0 members initially")

	// Add multiple members
	user1 := shared.NewID()
	user2 := shared.NewID()
	user3 := shared.NewID()

	members := []*project.ProjectMember{
		{ProjectID: projectID, UserID: user1, JoinedAt: time.Now()},
		{ProjectID: projectID, UserID: user2, JoinedAt: time.Now()},
		{ProjectID: projectID, UserID: user3, JoinedAt: time.Now()},
	}

	for _, member := range members {
		err = repo.AddMember(ctx, member)
		require.NoError(t, err, "Failed to add member")
	}

	// Test CountMembers after adding members
	count, err = repo.CountMembers(ctx, projectID)
	assert.NoError(t, err, "CountMembers should succeed")
	assert.Equal(t, int64(3), count, "Should have 3 members after adding")
}

