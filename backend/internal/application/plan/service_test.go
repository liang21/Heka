package plan

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T109 | spec.md: §4.8 PlanService TDD RED

type mockPlanRepo struct {
	mock.Mock
}

func (m *mockPlanRepo) Create(ctx context.Context, p *plan.TestPlan) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockPlanRepo) FindByID(ctx context.Context, id shared.ID) (*plan.TestPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plan.TestPlan), args.Error(1)
}

func (m *mockPlanRepo) List(ctx context.Context, projectID shared.ID, status *plan.PlanStatus, page, pageSize int) ([]*plan.TestPlan, int64, error) {
	args := m.Called(ctx, projectID, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*plan.TestPlan), args.Get(1).(int64), args.Error(2)
}

func (m *mockPlanRepo) Update(ctx context.Context, p *plan.TestPlan) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockPlanRepo) AddCases(ctx context.Context, planID shared.ID, cases []plan.PlanTestCase) error {
	args := m.Called(ctx, planID, cases)
	return args.Error(0)
}

func (m *mockPlanRepo) RemoveCases(ctx context.Context, planID shared.ID, caseIDs []shared.ID) error {
	args := m.Called(ctx, planID, caseIDs)
	return args.Error(0)
}

// --- Tests ---

func TestPlanService_Create(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	userID := shared.NewID()
	projectID := shared.NewID()
	caseID := shared.NewID()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)
	repo.On("AddCases", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.Create(context.Background(), userID, CreatePlanReq{
		ProjectID:   projectID,
		Name:        "Sprint 1 Test Plan",
		Description: "Test plan for sprint 1",
		Cases: []PlanCaseItem{
			{TestCaseID: caseID},
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Sprint 1 Test Plan", resp.Name)
	assert.Equal(t, plan.PlanDraft, resp.Status)
}

func TestPlanService_Start(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	userID := shared.NewID()
	caseID := shared.NewID()
	now := time.Now()

	p := &plan.TestPlan{
		ID:        planID,
		Status:    plan.PlanDraft,
		CreatedBy: userID,
		CreatedAt: now,
		Cases:     []plan.PlanTestCase{{PlanID: planID, TestCaseID: caseID}},
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)

	resp, err := svc.Start(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, plan.PlanActive, resp.Status)
}

func TestPlanService_StartWithoutCases(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()

	p := &plan.TestPlan{
		ID:        planID,
		Status:    plan.PlanDraft,
		CreatedBy: userID,
		CreatedAt: now,
		Cases:     []plan.PlanTestCase{}, // empty
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)

	resp, err := svc.Start(context.Background(), planID)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestPlanService_Pause(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	p := &plan.TestPlan{
		ID:     planID,
		Status: plan.PlanActive,
		Cases:  []plan.PlanTestCase{{TestCaseID: shared.NewID()}},
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)

	resp, err := svc.Pause(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, plan.PlanPaused, resp.Status)
}

func TestPlanService_Resume(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	p := &plan.TestPlan{
		ID:     planID,
		Status: plan.PlanPaused,
		Cases:  []plan.PlanTestCase{{TestCaseID: shared.NewID()}},
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)

	resp, err := svc.Resume(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, plan.PlanActive, resp.Status)
}

func TestPlanService_Complete(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	p := &plan.TestPlan{
		ID:     planID,
		Status: plan.PlanActive,
		Cases:  []plan.PlanTestCase{{TestCaseID: shared.NewID()}},
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)

	resp, err := svc.Complete(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, plan.PlanCompleted, resp.Status)
}

func TestPlanService_Cancel(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	p := &plan.TestPlan{
		ID:     planID,
		Status: plan.PlanDraft,
		Cases:  []plan.PlanTestCase{{TestCaseID: shared.NewID()}},
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*plan.TestPlan")).Return(nil)

	resp, err := svc.Cancel(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, plan.PlanCancelled, resp.Status)
}

func TestPlanService_InvalidTransition(t *testing.T) {
	t.Parallel()
	repo := new(mockPlanRepo)
	svc := NewService(repo)

	planID := shared.NewID()
	p := &plan.TestPlan{
		ID:     planID,
		Status: plan.PlanCompleted, // terminal state
	}

	repo.On("FindByID", mock.Anything, planID).Return(p, nil)

	resp, err := svc.Start(context.Background(), planID)

	assert.Nil(t, resp)
	assert.Error(t, err)
}
