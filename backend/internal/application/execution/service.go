package execution

import (
	"context"
	"time"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T113 | spec.md: §4.9 ExecutionService Implementation

type Service struct {
	repo execution.ExecutionRepository
}

func NewService(repo execution.ExecutionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, executorID, planID shared.ID, name string) (*ExecutionResponse, error) {
	e := &execution.TestExecution{
		ID:         shared.NewID(),
		PlanID:     planID,
		Name:       name,
		Status:     execution.ExecInProgress,
		ExecutorID: executorID,
		StartedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		ID:         e.ID,
		PlanID:     e.PlanID,
		Name:       e.Name,
		Status:     e.Status,
		ExecutorID: e.ExecutorID,
		StartedAt:  e.StartedAt,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, executionID shared.ID) (*ExecutionResponse, error) {
	e, err := s.repo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		ID:         e.ID,
		PlanID:     e.PlanID,
		Name:       e.Name,
		Status:     e.Status,
		ExecutorID: e.ExecutorID,
		StartedAt:  e.StartedAt,
		PausedAt:   e.PausedAt,
		CompletedAt: e.CompletedAt,
		Notes:      e.Notes,
	}, nil
}

func (s *Service) GetSummary(ctx context.Context, executionID shared.ID) (*ExecutionSummaryResponse, error) {
	summary, err := s.repo.GetSummary(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return &ExecutionSummaryResponse{
		Total:   summary.Total,
		Passed:  summary.Passed,
		Failed:  summary.Failed,
		Blocked: summary.Blocked,
		Skipped: summary.Skipped,
	}, nil
}

func (s *Service) SubmitResult(ctx context.Context, executorID, executionID shared.ID, req SubmitResultReq) error {
	result := &execution.ExecutionResult{
		ID:         shared.NewID(),
		ExecutionID: executionID,
		TestCaseID: req.TestCaseID,
		ExecutorID: executorID,
		Status:     req.Status,
		BugID:      req.BugID,
		BugURL:     req.BugURL,
		Notes:      req.Notes,
		ExecutedAt: time.Now(),
	}

	return s.repo.SubmitResult(ctx, result)
}

func (s *Service) BatchSubmit(ctx context.Context, executorID, executionID shared.ID, req BatchSubmitReq) error {
	results := make([]*execution.ExecutionResult, len(req.Results))
	for i, r := range req.Results {
		results[i] = &execution.ExecutionResult{
			ID:         shared.NewID(),
			ExecutionID: executionID,
			TestCaseID: r.TestCaseID,
			ExecutorID: executorID,
			Status:     r.Status,
			BugID:      r.BugID,
			BugURL:     r.BugURL,
			Notes:      r.Notes,
			ExecutedAt: time.Now(),
		}
	}

	return s.repo.BatchSubmitResults(ctx, results)
}
