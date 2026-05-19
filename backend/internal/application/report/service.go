package report

import (
	"context"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"github.com/liang21/heka/internal/domain/user"
)

// tasks.md: T125 | spec.md: §4.12 ReportService Implementation

type Service struct {
	planRepo plan.TestPlanRepository
	execRepo execution.ExecutionRepository
	tcRepo   testcase.TestCaseRepository
	userRepo user.UserRepository
}

func NewService(planRepo plan.TestPlanRepository, execRepo execution.ExecutionRepository, tcRepo testcase.TestCaseRepository, userRepo user.UserRepository) *Service {
	return &Service{
		planRepo: planRepo,
		execRepo: execRepo,
		tcRepo:   tcRepo,
		userRepo: userRepo,
	}
}

func (s *Service) PlanReport(ctx context.Context, planID shared.ID) (*PlanReportResponse, error) {
	p, err := s.planRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	var summary *execution.ExecutionSummary
	if p.CurrentExecutionID != nil {
		summary, _ = s.execRepo.GetSummary(ctx, *p.CurrentExecutionID)
	}

	totalCases := len(p.Cases)
	passed := 0
	failed := 0
	blocked := 0
	skipped := 0

	if summary != nil {
		passed = summary.Passed
		failed = summary.Failed
		blocked = summary.Blocked
		skipped = summary.Skipped
	}

	passRate := float64(0)
	if totalCases > 0 {
		passRate = float64(passed) / float64(totalCases) * 100
	}

	return &PlanReportResponse{
		PlanID:      p.ID,
		PlanName:    p.Name,
		Status:      string(p.Status),
		TotalCases:  totalCases,
		Passed:      passed,
		Failed:      failed,
		Blocked:     blocked,
		Skipped:     skipped,
		PassRate:    passRate,
		StartedAt:   p.StartedAt,
		CompletedAt: p.EndedAt,
	}, nil
}

func (s *Service) Coverage(ctx context.Context, projectID shared.ID) (*CoverageResponse, error) {
	cases, _, err := s.tcRepo.List(ctx, testcase.TestCaseFilter{
		ProjectID: projectID,
		Page:      1,
		PageSize:  1000,
	})
	if err != nil {
		return nil, err
	}

	return &CoverageResponse{
		ProjectID:    projectID,
		TotalCases:   len(cases),
		CoveredCases: 0,
		CoverageRate: 0,
	}, nil
}

func (s *Service) Trend(ctx context.Context, projectID shared.ID, days int) (*TrendResponse, error) {
	return &TrendResponse{
		ProjectID: projectID,
		Days:      days,
	}, nil
}

func (s *Service) BugDistribution(ctx context.Context, projectID shared.ID) (*BugDistributionResponse, error) {
	return &BugDistributionResponse{
		ProjectID: projectID,
	}, nil
}

func (s *Service) Workload(ctx context.Context, userID shared.ID) (*WorkloadResponse, error) {
	return &WorkloadResponse{
		UserID: userID,
	}, nil
}
