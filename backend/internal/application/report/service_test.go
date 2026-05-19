package report

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"github.com/liang21/heka/internal/domain/user"
)

// tasks.md: T124 | spec.md: §4.12 ReportService TDD RED

type mockReportPlanRepo struct {
	mock.Mock
}

func (m *mockReportPlanRepo) Create(ctx context.Context, p *plan.TestPlan) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockReportPlanRepo) FindByID(ctx context.Context, id shared.ID) (*plan.TestPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plan.TestPlan), args.Error(1)
}

func (m *mockReportPlanRepo) List(ctx context.Context, projectID shared.ID, status *plan.PlanStatus, page, pageSize int) ([]*plan.TestPlan, int64, error) {
	args := m.Called(ctx, projectID, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*plan.TestPlan), args.Get(1).(int64), args.Error(2)
}

func (m *mockReportPlanRepo) Update(ctx context.Context, p *plan.TestPlan) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockReportPlanRepo) AddCases(ctx context.Context, planID shared.ID, cases []plan.PlanTestCase) error {
	return m.Called(ctx, planID, cases).Error(0)
}

func (m *mockReportPlanRepo) RemoveCases(ctx context.Context, planID shared.ID, caseIDs []shared.ID) error {
	return m.Called(ctx, planID, caseIDs).Error(0)
}

type mockReportExecRepo struct {
	mock.Mock
}

func (m *mockReportExecRepo) Create(ctx context.Context, e *execution.TestExecution) error {
	return m.Called(ctx, e).Error(0)
}

func (m *mockReportExecRepo) FindByID(ctx context.Context, id shared.ID) (*execution.TestExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.TestExecution), args.Error(1)
}

func (m *mockReportExecRepo) SubmitResult(ctx context.Context, r *execution.ExecutionResult) error {
	return m.Called(ctx, r).Error(0)
}

func (m *mockReportExecRepo) BatchSubmitResults(ctx context.Context, results []*execution.ExecutionResult) error {
	return m.Called(ctx, results).Error(0)
}

func (m *mockReportExecRepo) GetSummary(ctx context.Context, executionID shared.ID) (*execution.ExecutionSummary, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.ExecutionSummary), args.Error(1)
}

type mockReportTCRepo struct {
	mock.Mock
}

func (m *mockReportTCRepo) Create(ctx context.Context, tc *testcase.TestCase) error {
	return m.Called(ctx, tc).Error(0)
}

func (m *mockReportTCRepo) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*testcase.TestCase), args.Error(1)
}

func (m *mockReportTCRepo) List(ctx context.Context, f testcase.TestCaseFilter) ([]*testcase.TestCase, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*testcase.TestCase), args.Get(1).(int64), args.Error(2)
}

func (m *mockReportTCRepo) Update(ctx context.Context, tc *testcase.TestCase) error {
	return m.Called(ctx, tc).Error(0)
}

func (m *mockReportTCRepo) SoftDelete(ctx context.Context, id shared.ID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockReportTCRepo) BatchUpdateStatus(ctx context.Context, ids []shared.ID, status testcase.CaseStatus) error {
	return m.Called(ctx, ids, status).Error(0)
}

func (m *mockReportTCRepo) BatchDelete(ctx context.Context, ids []shared.ID) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *mockReportTCRepo) BatchMove(ctx context.Context, ids []shared.ID, moduleID *shared.ID) error {
	return m.Called(ctx, ids, moduleID).Error(0)
}

type mockReportUserRepo struct {
	mock.Mock
}

func (m *mockReportUserRepo) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockReportUserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockReportUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockReportUserRepo) Update(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

// --- Tests ---

func TestReportService_PlanReport(t *testing.T) {
	t.Parallel()
	planRepo := new(mockReportPlanRepo)
	execRepo := new(mockReportExecRepo)
	tcRepo := new(mockReportTCRepo)
	userRepo := new(mockReportUserRepo)
	svc := NewService(planRepo, execRepo, tcRepo, userRepo)

	planID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()

	p := &plan.TestPlan{
		ID:        planID,
		Name:      "Sprint 1",
		Status:    plan.PlanCompleted,
		StartedAt: &now,
		Cases:     []plan.PlanTestCase{{TestCaseID: shared.NewID()}},
	}

	summary := &execution.ExecutionSummary{
		Total: 10, Passed: 7, Failed: 2, Blocked: 1, Skipped: 0,
	}

	planRepo.On("FindByID", mock.Anything, planID).Return(p, nil)
	execRepo.On("GetSummary", mock.Anything, mock.Anything).Return(summary, nil)

	resp, err := svc.PlanReport(context.Background(), planID)

	require.NoError(t, err)
	assert.Equal(t, planID, resp.PlanID)
	assert.Equal(t, 10, resp.TotalCases)
	assert.Equal(t, 7, resp.Passed)
}

func TestReportService_Coverage(t *testing.T) {
	t.Parallel()
	planRepo := new(mockReportPlanRepo)
	execRepo := new(mockReportExecRepo)
	tcRepo := new(mockReportTCRepo)
	userRepo := new(mockReportUserRepo)
	svc := NewService(planRepo, execRepo, tcRepo, userRepo)

	projectID := shared.NewID()

	tcRepo.On("List", mock.Anything, mock.Anything).Return([]*testcase.TestCase{}, int64(0), nil)

	resp, err := svc.Coverage(context.Background(), projectID)

	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ProjectID)
}

func TestReportService_Trend(t *testing.T) {
	t.Parallel()
	planRepo := new(mockReportPlanRepo)
	execRepo := new(mockReportExecRepo)
	tcRepo := new(mockReportTCRepo)
	userRepo := new(mockReportUserRepo)
	svc := NewService(planRepo, execRepo, tcRepo, userRepo)

	projectID := shared.NewID()

	resp, err := svc.Trend(context.Background(), projectID, 30)

	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ProjectID)
	assert.Equal(t, 30, resp.Days)
}

func TestReportService_BugDistribution(t *testing.T) {
	t.Parallel()
	planRepo := new(mockReportPlanRepo)
	execRepo := new(mockReportExecRepo)
	tcRepo := new(mockReportTCRepo)
	userRepo := new(mockReportUserRepo)
	svc := NewService(planRepo, execRepo, tcRepo, userRepo)

	projectID := shared.NewID()

	resp, err := svc.BugDistribution(context.Background(), projectID)

	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ProjectID)
}

func TestReportService_Workload(t *testing.T) {
	t.Parallel()
	planRepo := new(mockReportPlanRepo)
	execRepo := new(mockReportExecRepo)
	tcRepo := new(mockReportTCRepo)
	userRepo := new(mockReportUserRepo)
	svc := NewService(planRepo, execRepo, tcRepo, userRepo)

	userID := shared.NewID()

	resp, err := svc.Workload(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, userID, resp.UserID)
}
