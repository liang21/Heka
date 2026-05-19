package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/plan"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T132 | spec.md: §4.8 Plan Handler TDD RED

type mockPlanService struct {
	mock.Mock
}

func (m *mockPlanService) Create(ctx context.Context, userID shared.ID, req plan.CreatePlanReq) (*plan.PlanResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanResponse), args.Error(1)
}

func (m *mockPlanService) GetByID(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) Start(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) Pause(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) Resume(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) Complete(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) Cancel(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*plan.PlanDetailResponse), args.Error(1)
}

func (m *mockPlanService) List(ctx context.Context, projectID shared.ID) ([]plan.PlanResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]plan.PlanResponse), args.Error(1)
}

func (m *mockPlanService) AddCase(ctx context.Context, planID, caseID shared.ID) error {
	args := m.Called(ctx, planID, caseID)
	return args.Error(0)
}

func (m *mockPlanService) RemoveCase(ctx context.Context, planID, caseID shared.ID) error {
	args := m.Called(ctx, planID, caseID)
	return args.Error(0)
}

func TestPlanHandler_Start(t *testing.T) {
	t.Parallel()
	svc := new(mockPlanService)
	h := NewPlanHandler(svc)

	planID := shared.NewID()
	svc.On("Start", mock.Anything, planID).Return(&plan.PlanDetailResponse{
		ID: planID, Status: "active",
	}, nil)

	req := httptest.NewRequest("POST", "/api/v1/testplans/"+planID.String()+"/start", nil)
	req = setChiURLParam(req, "id", planID.String())
	w := httptest.NewRecorder()

	h.Start(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Pause(t *testing.T) {
	t.Parallel()
	svc := new(mockPlanService)
	h := NewPlanHandler(svc)

	planID := shared.NewID()
	svc.On("Pause", mock.Anything, planID).Return(&plan.PlanDetailResponse{
		ID: planID, Status: "paused",
	}, nil)

	req := httptest.NewRequest("POST", "/api/v1/testplans/"+planID.String()+"/pause", nil)
	req = setChiURLParam(req, "id", planID.String())
	w := httptest.NewRecorder()

	h.Pause(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Complete(t *testing.T) {
	t.Parallel()
	svc := new(mockPlanService)
	h := NewPlanHandler(svc)

	planID := shared.NewID()
	svc.On("Complete", mock.Anything, planID).Return(&plan.PlanDetailResponse{
		ID: planID, Status: "completed",
	}, nil)

	req := httptest.NewRequest("POST", "/api/v1/testplans/"+planID.String()+"/complete", nil)
	req = setChiURLParam(req, "id", planID.String())
	w := httptest.NewRecorder()

	h.Complete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Cancel(t *testing.T) {
	t.Parallel()
	svc := new(mockPlanService)
	h := NewPlanHandler(svc)

	planID := shared.NewID()
	svc.On("Cancel", mock.Anything, planID).Return(&plan.PlanDetailResponse{
		ID: planID, Status: "cancelled",
	}, nil)

	req := httptest.NewRequest("POST", "/api/v1/testplans/"+planID.String()+"/cancel", nil)
	req = setChiURLParam(req, "id", planID.String())
	w := httptest.NewRecorder()

	h.Cancel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
