package ai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/ai"
)

// tasks.md: T121 | spec.md: §4.11 AIService TDD RED

type mockAIManager struct {
	mock.Mock
}

func (m *mockAIManager) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.ChatResponse), args.Error(1)
}

type mockAsyncTaskRepo struct {
	mock.Mock
}

func (m *mockAsyncTaskRepo) Create(ctx context.Context, task *shared.AsyncTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockAsyncTaskRepo) FindByID(ctx context.Context, id shared.ID) (*shared.AsyncTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.AsyncTask), args.Error(1)
}

func (m *mockAsyncTaskRepo) FindPendingByType(ctx context.Context, projectID shared.ID, taskType string, limit int) ([]*shared.AsyncTask, error) {
	args := m.Called(ctx, projectID, taskType, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*shared.AsyncTask), args.Error(1)
}

func (m *mockAsyncTaskRepo) Update(ctx context.Context, task *shared.AsyncTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

type mockAIEventBus struct {
	mock.Mock
}

func (m *mockAIEventBus) Publish(ctx context.Context, events ...shared.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *mockAIEventBus) Subscribe(eventName string, handler shared.EventHandler) {
	m.Called(eventName, handler)
}

func TestAIService_GenerateTestCases(t *testing.T) {
	t.Parallel()
	manager := new(mockAIManager)
	taskRepo := new(mockAsyncTaskRepo)
	bus := new(mockAIEventBus)
	svc := NewService(manager, nil, taskRepo, bus)

	projectID := shared.NewID()
	userID := shared.NewID()
	fileID := shared.NewID()

	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*shared.AsyncTask")).Return(nil)

	resp, err := svc.GenerateTestCases(context.Background(), userID, GenerateRequest{
		ProjectID:       projectID,
		FileID:          fileID,
		Count:           5,
		IncludeNegative: true,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.TaskID)
	assert.Equal(t, "pending", resp.Status)
}

func TestAIService_Analyze(t *testing.T) {
	t.Parallel()
	manager := new(mockAIManager)
	taskRepo := new(mockAsyncTaskRepo)
	bus := new(mockAIEventBus)
	svc := NewService(manager, nil, taskRepo, bus)

	projectID := shared.NewID()
	userID := shared.NewID()

	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*shared.AsyncTask")).Return(nil)

	resp, err := svc.Analyze(context.Background(), userID, AnalyzeRequest{
		ProjectID:   projectID,
		Description: "Changed login flow to use OAuth2",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.TaskID)
	assert.Equal(t, "pending", resp.Status)
}

func TestAIService_GetTask(t *testing.T) {
	t.Parallel()
	manager := new(mockAIManager)
	taskRepo := new(mockAsyncTaskRepo)
	bus := new(mockAIEventBus)
	svc := NewService(manager, nil, taskRepo, bus)

	taskID := shared.NewID()
	result := json.RawMessage(`{"test_cases": []}`)
	task := &shared.AsyncTask{
		ID:              taskID,
		Type:            "generate_testcases",
		Status:          "completed",
		ProgressCurrent: 5,
		ProgressTotal:   5,
		Result:          result,
	}

	taskRepo.On("FindByID", mock.Anything, taskID).Return(task, nil)

	resp, err := svc.GetTask(context.Background(), taskID)

	require.NoError(t, err)
	assert.Equal(t, taskID, resp.TaskID)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 5, resp.Progress)
	assert.Equal(t, 5, resp.Total)
}
