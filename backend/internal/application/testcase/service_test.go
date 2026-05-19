package testcase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
)

// tasks.md: T106 | spec.md: §4.4-4.7 TestCaseService TDD RED

// --- Mocks ---

type mockTestCaseRepo struct {
	mock.Mock
}

func (m *mockTestCaseRepo) Create(ctx context.Context, tc *testcase.TestCase) error {
	args := m.Called(ctx, tc)
	return args.Error(0)
}

func (m *mockTestCaseRepo) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*testcase.TestCase), args.Error(1)
}

func (m *mockTestCaseRepo) List(ctx context.Context, f testcase.TestCaseFilter) ([]*testcase.TestCase, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*testcase.TestCase), args.Get(1).(int64), args.Error(2)
}

func (m *mockTestCaseRepo) Update(ctx context.Context, tc *testcase.TestCase) error {
	args := m.Called(ctx, tc)
	return args.Error(0)
}

func (m *mockTestCaseRepo) SoftDelete(ctx context.Context, id shared.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTestCaseRepo) BatchUpdateStatus(ctx context.Context, ids []shared.ID, status testcase.CaseStatus) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

func (m *mockTestCaseRepo) BatchDelete(ctx context.Context, ids []shared.ID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *mockTestCaseRepo) BatchMove(ctx context.Context, ids []shared.ID, moduleID *shared.ID) error {
	args := m.Called(ctx, ids, moduleID)
	return args.Error(0)
}

type mockModuleRepo struct {
	mock.Mock
}

func (m *mockModuleRepo) Create(ctx context.Context, mod *testcase.Module) error {
	args := m.Called(ctx, mod)
	return args.Error(0)
}

func (m *mockModuleRepo) Update(ctx context.Context, mod *testcase.Module) error {
	args := m.Called(ctx, mod)
	return args.Error(0)
}

func (m *mockModuleRepo) Delete(ctx context.Context, id shared.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockModuleRepo) FindByProject(ctx context.Context, projectID shared.ID) ([]*testcase.Module, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*testcase.Module), args.Error(1)
}

type mockTagRepo struct {
	mock.Mock
}

func (m *mockTagRepo) FindByProject(ctx context.Context, projectID shared.ID) ([]*testcase.Tag, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*testcase.Tag), args.Error(1)
}

func (m *mockTagRepo) Create(ctx context.Context, tag *testcase.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

type mockCollectionRepo struct {
	mock.Mock
}

func (m *mockCollectionRepo) Create(ctx context.Context, c *testcase.Collection) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCollectionRepo) AddCases(ctx context.Context, collectionID shared.ID, caseIDs []shared.ID) error {
	args := m.Called(ctx, collectionID, caseIDs)
	return args.Error(0)
}

func (m *mockCollectionRepo) RemoveCases(ctx context.Context, collectionID shared.ID, caseIDs []shared.ID) error {
	args := m.Called(ctx, collectionID, caseIDs)
	return args.Error(0)
}

func (m *mockCollectionRepo) ListCases(ctx context.Context, collectionID shared.ID, page, pageSize int) ([]*testcase.TestCase, int64, error) {
	args := m.Called(ctx, collectionID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*testcase.TestCase), args.Get(1).(int64), args.Error(2)
}

type mockEventBus struct {
	mock.Mock
}

func (m *mockEventBus) Publish(ctx context.Context, events ...shared.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *mockEventBus) Subscribe(eventName string, handler shared.EventHandler) {
	m.Called(eventName, handler)
}

// --- Helper ---

func newTestService(tcRepo *mockTestCaseRepo, modRepo *mockModuleRepo, tagRepo *mockTagRepo, collRepo *mockCollectionRepo, bus *mockEventBus) *Service {
	return NewService(tcRepo, modRepo, tagRepo, collRepo, bus)
}

// --- TestCase Tests ---

func TestTestCaseService_CreateTestCase(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	modRepo := new(mockModuleRepo)
	tagRepo := new(mockTagRepo)
	collRepo := new(mockCollectionRepo)
	bus := new(mockEventBus)
	svc := newTestService(tcRepo, modRepo, tagRepo, collRepo, bus)

	userID := shared.NewID()
	projectID := shared.NewID()

	tcRepo.On("Create", mock.Anything, mock.AnythingOfType("*testcase.TestCase")).Return(nil)
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.CreateTestCase(context.Background(), userID, projectID, CreateTestCaseReq{
		Title:    "Login test",
		Priority: testcase.P1,
		Steps: []StepInput{
			{Action: "Open login page", Expected: "Login form displayed"},
			{Action: "Enter credentials", Expected: "Logged in"},
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Login test", resp.Title)
	assert.Equal(t, testcase.CaseDraft, resp.Status)
	assert.Len(t, resp.Steps, 2)
}

func TestTestCaseService_CreateWithSteps(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	modRepo := new(mockModuleRepo)
	tagRepo := new(mockTagRepo)
	collRepo := new(mockCollectionRepo)
	bus := new(mockEventBus)
	svc := newTestService(tcRepo, modRepo, tagRepo, collRepo, bus)

	userID := shared.NewID()
	projectID := shared.NewID()

	tcRepo.On("Create", mock.Anything, mock.AnythingOfType("*testcase.TestCase")).Return(nil)
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.CreateTestCase(context.Background(), userID, projectID, CreateTestCaseReq{
		Title:    "Complex test",
		Priority: testcase.P0,
		Tags:     []string{"smoke", "regression"},
		Steps: []StepInput{
			{Action: "Step 1", Expected: "Result 1"},
			{Action: "Step 2", Expected: "Result 2"},
			{Action: "Step 3", Expected: "Result 3"},
		},
	})

	require.NoError(t, err)
	assert.Len(t, resp.Steps, 3)
	assert.Contains(t, resp.Tags, "smoke")
}

func TestTestCaseService_GetByID(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	caseID := shared.NewID()
	projectID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()

	tc := &testcase.TestCase{
		ID:        caseID,
		ProjectID: projectID,
		Title:     "Test case",
		Status:    testcase.CaseDraft,
		Priority:  testcase.P1,
		Steps: []testcase.Step{
			{ID: shared.NewID(), TestCaseID: caseID, Number: 1, Action: "Do X", Expected: "See Y"},
		},
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcRepo.On("FindByID", mock.Anything, caseID).Return(tc, nil)

	resp, err := svc.GetTestCase(context.Background(), caseID)

	require.NoError(t, err)
	assert.Equal(t, caseID, resp.ID)
	assert.Equal(t, "Test case", resp.Title)
	assert.Len(t, resp.Steps, 1)
}

func TestTestCaseService_ListWithFilter(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	projectID := shared.NewID()
	status := testcase.CaseDraft
	cases := []*testcase.TestCase{
		{ID: shared.NewID(), ProjectID: projectID, Title: "Case 1", Status: testcase.CaseDraft},
		{ID: shared.NewID(), ProjectID: projectID, Title: "Case 2", Status: testcase.CaseDraft},
	}

	filter := testcase.TestCaseFilter{
		ProjectID: projectID,
		Status:    &status,
		Page:      1,
		PageSize:  20,
	}

	tcRepo.On("List", mock.Anything, filter).Return(cases, int64(2), nil)

	items, total, err := svc.ListTestCases(context.Background(), filter)

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(2), total)
}

func TestTestCaseService_UpdateStatusTransition(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	bus := new(mockEventBus)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), bus)

	caseID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()

	tc := &testcase.TestCase{
		ID:        caseID,
		Title:     "Test case",
		Status:    testcase.CaseDraft,
		Version:   1,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcRepo.On("FindByID", mock.Anything, caseID).Return(tc, nil)
	tcRepo.On("Update", mock.Anything, mock.AnythingOfType("*testcase.TestCase")).Return(nil)
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.UpdateTestCase(context.Background(), userID, caseID, UpdateTestCaseReq{
		Title:   "Updated title",
		Version: 1,
		Steps:   []StepInput{{Action: "New step", Expected: "New result"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated title", resp.Title)
}

func TestTestCaseService_UpdateVersionConflict(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	caseID := shared.NewID()
	userID := shared.NewID()

	tcRepo.On("FindByID", mock.Anything, caseID).Return(nil, shared.ErrTestCaseNotFound)

	resp, err := svc.UpdateTestCase(context.Background(), userID, caseID, UpdateTestCaseReq{
		Title:   "Updated",
		Version: 2,
		Steps:   []StepInput{{Action: "Step", Expected: "Result"}},
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestTestCaseService_BatchUpdateStatus(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	ids := []shared.ID{shared.NewID(), shared.NewID()}
	tcRepo.On("BatchUpdateStatus", mock.Anything, ids, testcase.CaseReady).Return(nil)

	err := svc.BatchUpdateStatus(context.Background(), BatchStatusReq{
		IDs:    ids,
		Status: testcase.CaseReady,
	})

	require.NoError(t, err)
}

func TestTestCaseService_BatchDelete(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	ids := []shared.ID{shared.NewID(), shared.NewID()}
	tcRepo.On("BatchDelete", mock.Anything, ids).Return(nil)

	err := svc.BatchDelete(context.Background(), BatchDeleteReq{IDs: ids})

	require.NoError(t, err)
}

func TestTestCaseService_BatchMove(t *testing.T) {
	t.Parallel()
	tcRepo := new(mockTestCaseRepo)
	svc := newTestService(tcRepo, new(mockModuleRepo), new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	ids := []shared.ID{shared.NewID(), shared.NewID()}
	targetModule := shared.NewID()
	tcRepo.On("BatchMove", mock.Anything, ids, &targetModule).Return(nil)

	err := svc.BatchMove(context.Background(), BatchMoveReq{
		IDs:      ids,
		ModuleID: &targetModule,
	})

	require.NoError(t, err)
}

// --- Module Tests ---

func TestTestCaseService_CreateModule(t *testing.T) {
	t.Parallel()
	modRepo := new(mockModuleRepo)
	svc := newTestService(new(mockTestCaseRepo), modRepo, new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	userID := shared.NewID()
	projectID := shared.NewID()

	modRepo.On("Create", mock.Anything, mock.AnythingOfType("*testcase.Module")).Return(nil)

	resp, err := svc.CreateModule(context.Background(), userID, CreateModuleReq{
		ProjectID: projectID,
		Name:      "Authentication",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Authentication", resp.Name)
}

func TestTestCaseService_GetModuleTree(t *testing.T) {
	t.Parallel()
	modRepo := new(mockModuleRepo)
	svc := newTestService(new(mockTestCaseRepo), modRepo, new(mockTagRepo), new(mockCollectionRepo), new(mockEventBus))

	projectID := shared.NewID()
	parentID := shared.NewID()
	childID := shared.NewID()
	userID := shared.NewID()

	modules := []*testcase.Module{
		{ID: parentID, ProjectID: projectID, Name: "Parent", ParentID: nil, CreatedBy: userID},
		{ID: childID, ProjectID: projectID, Name: "Child", ParentID: &parentID, CreatedBy: userID},
	}

	modRepo.On("FindByProject", mock.Anything, projectID).Return(modules, nil)

	tree, err := svc.GetModuleTree(context.Background(), projectID)

	require.NoError(t, err)
	assert.Len(t, tree, 1)            // 1 root
	assert.Len(t, tree[0].Children, 1) // 1 child
}

// --- Tag Tests ---

func TestTestCaseService_CreateTag(t *testing.T) {
	t.Parallel()
	tagRepo := new(mockTagRepo)
	svc := newTestService(new(mockTestCaseRepo), new(mockModuleRepo), tagRepo, new(mockCollectionRepo), new(mockEventBus))

	userID := shared.NewID()
	projectID := shared.NewID()

	tagRepo.On("Create", mock.Anything, mock.AnythingOfType("*testcase.Tag")).Return(nil)

	resp, err := svc.CreateTag(context.Background(), userID, CreateTagReq{
		ProjectID: projectID,
		Name:      "smoke",
		Color:     "#FF0000",
	})

	require.NoError(t, err)
	assert.Equal(t, "smoke", resp.Name)
}

// --- Collection Tests ---

func TestTestCaseService_CreateCollection(t *testing.T) {
	t.Parallel()
	collRepo := new(mockCollectionRepo)
	svc := newTestService(new(mockTestCaseRepo), new(mockModuleRepo), new(mockTagRepo), collRepo, new(mockEventBus))

	userID := shared.NewID()
	projectID := shared.NewID()

	collRepo.On("Create", mock.Anything, mock.AnythingOfType("*testcase.Collection")).Return(nil)

	resp, err := svc.CreateCollection(context.Background(), userID, CreateCollectionReq{
		ProjectID:   projectID,
		Name:        "Regression Suite",
		Description: "Full regression",
	})

	require.NoError(t, err)
	assert.Equal(t, "Regression Suite", resp.Name)
}
