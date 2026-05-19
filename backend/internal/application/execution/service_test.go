package execution

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T112 | spec.md: §4.9 ExecutionService TDD RED

type mockExecRepo struct {
	mock.Mock
}

func (m *mockExecRepo) Create(ctx context.Context, e *execution.TestExecution) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockExecRepo) FindByID(ctx context.Context, id shared.ID) (*execution.TestExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.TestExecution), args.Error(1)
}

func (m *mockExecRepo) SubmitResult(ctx context.Context, r *execution.ExecutionResult) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *mockExecRepo) BatchSubmitResults(ctx context.Context, results []*execution.ExecutionResult) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *mockExecRepo) GetSummary(ctx context.Context, executionID shared.ID) (*execution.ExecutionSummary, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.ExecutionSummary), args.Error(1)
}

func TestExecutionService_CreateExecution(t *testing.T) {
	t.Parallel()
	repo := new(mockExecRepo)
	svc := NewService(repo)

	executorID := shared.NewID()
	planID := shared.NewID()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*execution.TestExecution")).Return(nil)

	resp, err := svc.Create(context.Background(), executorID, planID, "Execution 1")

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, execution.ExecInProgress, resp.Status)
	assert.Equal(t, planID, resp.PlanID)
}

func TestExecutionService_SubmitResult(t *testing.T) {
	t.Parallel()
	repo := new(mockExecRepo)
	svc := NewService(repo)

	execID := shared.NewID()
	caseID := shared.NewID()
	executorID := shared.NewID()

	repo.On("SubmitResult", mock.Anything, mock.AnythingOfType("*execution.ExecutionResult")).Return(nil)

	err := svc.SubmitResult(context.Background(), executorID, execID, SubmitResultReq{
		TestCaseID: caseID,
		Status:     execution.ResultPassed,
		Notes:      "All good",
	})

	require.NoError(t, err)
}

func TestExecutionService_BatchSubmit(t *testing.T) {
	t.Parallel()
	repo := new(mockExecRepo)
	svc := NewService(repo)

	execID := shared.NewID()
	executorID := shared.NewID()

	repo.On("BatchSubmitResults", mock.Anything, mock.Anything).Return(nil)

	err := svc.BatchSubmit(context.Background(), executorID, execID, BatchSubmitReq{
		Results: []SubmitResultReq{
			{TestCaseID: shared.NewID(), Status: execution.ResultPassed},
			{TestCaseID: shared.NewID(), Status: execution.ResultFailed, BugID: "BUG-001"},
		},
	})

	require.NoError(t, err)
}

func TestExecutionService_GetByID(t *testing.T) {
	t.Parallel()
	repo := new(mockExecRepo)
	svc := NewService(repo)

	execID := shared.NewID()
	planID := shared.NewID()
	executorID := shared.NewID()
	now := time.Now()

	e := &execution.TestExecution{
		ID:         execID,
		PlanID:     planID,
		Name:       "Execution 1",
		Status:     execution.ExecInProgress,
		ExecutorID: executorID,
		StartedAt:  now,
	}

	repo.On("FindByID", mock.Anything, execID).Return(e, nil)

	resp, err := svc.GetByID(context.Background(), execID)

	require.NoError(t, err)
	assert.Equal(t, execID, resp.ID)
	assert.Equal(t, execution.ExecInProgress, resp.Status)
}

func TestExecutionService_GetSummary(t *testing.T) {
	t.Parallel()
	repo := new(mockExecRepo)
	svc := NewService(repo)

	execID := shared.NewID()
	summary := &execution.ExecutionSummary{
		Total: 10, Passed: 7, Failed: 2, Blocked: 1, Skipped: 0,
	}

	repo.On("GetSummary", mock.Anything, execID).Return(summary, nil)

	resp, err := svc.GetSummary(context.Background(), execID)

	require.NoError(t, err)
	assert.Equal(t, 10, resp.Total)
	assert.Equal(t, 7, resp.Passed)
	assert.Equal(t, 2, resp.Failed)
}
