package project

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T103 | spec.md: §4.3 ProjectService TDD RED

type mockProjectRepo struct {
	mock.Mock
}

func (m *mockProjectRepo) Create(ctx context.Context, p *project.Project) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockProjectRepo) FindByID(ctx context.Context, id shared.ID) (*project.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *mockProjectRepo) FindByUserID(ctx context.Context, userID shared.ID) ([]*project.Project, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.Project), args.Error(1)
}

func (m *mockProjectRepo) IsMember(ctx context.Context, projectID, userID shared.ID) (bool, error) {
	args := m.Called(ctx, projectID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockProjectRepo) AddMember(ctx context.Context, member *project.ProjectMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockProjectRepo) CountMembers(ctx context.Context, projectID shared.ID) (int64, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(int64), args.Error(1)
}

// --- Tests ---

func TestProjectService_Create(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	userID := shared.NewID()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*project.Project")).Return(nil)

	resp, err := svc.Create(context.Background(), userID, CreateProjectReq{
		Name:        "Test Project",
		Description: "A test project",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Test Project", resp.Name)
	repo.AssertExpectations(t)
}

func TestProjectService_GetByID(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	projectID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()
	p := &project.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "A test project",
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo.On("FindByID", mock.Anything, projectID).Return(p, nil)
	repo.On("CountMembers", mock.Anything, projectID).Return(int64(3), nil)

	resp, err := svc.GetByID(context.Background(), projectID)

	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ID)
	assert.Equal(t, "Test Project", resp.Name)
	repo.AssertExpectations(t)
}

func TestProjectService_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	projectID := shared.NewID()
	repo.On("FindByID", mock.Anything, projectID).Return(nil, shared.ErrProjectNotFound)

	resp, err := svc.GetByID(context.Background(), projectID)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestProjectService_ListByUser(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	userID := shared.NewID()
	projects := []*project.Project{
		{ID: shared.NewID(), Name: "Project 1", CreatedBy: userID},
		{ID: shared.NewID(), Name: "Project 2", CreatedBy: userID},
	}

	repo.On("FindByUserID", mock.Anything, userID).Return(projects, nil)

	resp, err := svc.ListByUser(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, resp, 2)
	repo.AssertExpectations(t)
}

func TestProjectService_AddMember(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	projectID := shared.NewID()
	creatorID := shared.NewID()
	newUserID := shared.NewID()

	repo.On("IsMember", mock.Anything, projectID, newUserID).Return(false, nil)
	repo.On("AddMember", mock.Anything, mock.AnythingOfType("*project.ProjectMember")).Return(nil)

	err := svc.AddMember(context.Background(), projectID, creatorID, AddMemberReq{UserID: newUserID})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestProjectService_AddMember_AlreadyExists(t *testing.T) {
	t.Parallel()
	repo := new(mockProjectRepo)
	svc := NewService(repo)

	projectID := shared.NewID()
	creatorID := shared.NewID()
	existingUserID := shared.NewID()

	repo.On("IsMember", mock.Anything, projectID, existingUserID).Return(true, nil)

	err := svc.AddMember(context.Background(), projectID, creatorID, AddMemberReq{UserID: existingUserID})

	assert.Error(t, err)
	repo.AssertNotCalled(t, "AddMember")
}
