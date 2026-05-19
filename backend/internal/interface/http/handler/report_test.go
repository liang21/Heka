package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/report"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T140 | spec.md: §4.12 Report Handler TDD RED

type mockReportService struct {
	mock.Mock
}

func (m *mockReportService) PlanReport(ctx context.Context, planID shared.ID) (*report.PlanReportResponse, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*report.PlanReportResponse), args.Error(1)
}

func (m *mockReportService) Coverage(ctx context.Context, projectID shared.ID) (*report.CoverageResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*report.CoverageResponse), args.Error(1)
}

func (m *mockReportService) Trend(ctx context.Context, projectID shared.ID, days int) (*report.TrendResponse, error) {
	args := m.Called(ctx, projectID, days)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*report.TrendResponse), args.Error(1)
}

func (m *mockReportService) BugDistribution(ctx context.Context, projectID shared.ID) (*report.BugDistributionResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*report.BugDistributionResponse), args.Error(1)
}

func (m *mockReportService) Workload(ctx context.Context, userID shared.ID) (*report.WorkloadResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*report.WorkloadResponse), args.Error(1)
}

func TestReportHandler_PlanReport(t *testing.T) {
	t.Parallel()
	svc := new(mockReportService)
	h := NewReportHandler(svc)

	planID := shared.NewID()
	svc.On("PlanReport", mock.Anything, planID).Return(&report.PlanReportResponse{
		PlanID: planID, PlanName: "Sprint 1", TotalCases: 10,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/reports/plan/"+planID.String(), nil)
	req = setChiURLParam(req, "id", planID.String())
	w := httptest.NewRecorder()

	h.PlanReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Coverage(t *testing.T) {
	t.Parallel()
	svc := new(mockReportService)
	h := NewReportHandler(svc)

	projectID := shared.NewID()
	svc.On("Coverage", mock.Anything, projectID).Return(&report.CoverageResponse{
		ProjectID: projectID, TotalCases: 100, CoveredCases: 80,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/reports/coverage?project_id="+projectID.String(), nil)
	w := httptest.NewRecorder()

	h.Coverage(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Trend(t *testing.T) {
	t.Parallel()
	svc := new(mockReportService)
	h := NewReportHandler(svc)

	projectID := shared.NewID()
	svc.On("Trend", mock.Anything, projectID, 30).Return(&report.TrendResponse{
		ProjectID: projectID, Days: 30,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/reports/trend?project_id="+projectID.String()+"&days=30", nil)
	w := httptest.NewRecorder()

	h.Trend(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_BugDistribution(t *testing.T) {
	t.Parallel()
	svc := new(mockReportService)
	h := NewReportHandler(svc)

	projectID := shared.NewID()
	svc.On("BugDistribution", mock.Anything, projectID).Return(&report.BugDistributionResponse{
		ProjectID: projectID,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/reports/bugs?project_id="+projectID.String(), nil)
	w := httptest.NewRecorder()

	h.BugDistribution(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Workload(t *testing.T) {
	t.Parallel()
	svc := new(mockReportService)
	h := NewReportHandler(svc)

	userID := shared.NewID()
	svc.On("Workload", mock.Anything, userID).Return(&report.WorkloadResponse{
		UserID: userID,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/reports/workload", nil)
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.Workload(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
