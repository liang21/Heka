package plan

import (
	"context"
	"time"

	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T110 | spec.md: §4.8 PlanService Implementation

type Service struct {
	repo plan.TestPlanRepository
}

func NewService(repo plan.TestPlanRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID shared.ID, req CreatePlanReq) (*PlanResponse, error) {
	p := &plan.TestPlan{
		ID:        shared.NewID(),
		ProjectID: req.ProjectID,
		Name:      req.Name,
		Status:    plan.PlanDraft,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	if len(req.Cases) > 0 {
		cases := make([]plan.PlanTestCase, len(req.Cases))
		for i, c := range req.Cases {
			cases[i] = plan.PlanTestCase{
				PlanID:     p.ID,
				TestCaseID: c.TestCaseID,
				AssignedTo: c.AssignedTo,
				OrderIndex: i,
			}
		}
		if err := s.repo.AddCases(ctx, p.ID, cases); err != nil {
			return nil, err
		}
	}

	return &PlanResponse{
		ID:        p.ID,
		Name:      p.Name,
		Status:    p.Status,
		CreatedBy: p.CreatedBy,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (s *Service) Start(ctx context.Context, planID shared.ID) (*PlanDetailResponse, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidatePlanTransition(p.Status, plan.PlanActive); err != nil {
		return nil, err
	}

	if len(p.Cases) == 0 {
		return nil, shared.ErrPlanNeedsCases
	}

	p.Status = plan.PlanActive
	now := time.Now()
	p.StartedAt = &now

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.toDetailResponse(p)
}

func (s *Service) Pause(ctx context.Context, planID shared.ID) (*PlanDetailResponse, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidatePlanTransition(p.Status, plan.PlanPaused); err != nil {
		return nil, err
	}

	p.Status = plan.PlanPaused
	now := time.Now()
	p.PausedAt = &now

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.toDetailResponse(p)
}

func (s *Service) Resume(ctx context.Context, planID shared.ID) (*PlanDetailResponse, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidatePlanTransition(p.Status, plan.PlanActive); err != nil {
		return nil, err
	}

	p.Status = plan.PlanActive
	p.PausedAt = nil

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.toDetailResponse(p)
}

func (s *Service) Complete(ctx context.Context, planID shared.ID) (*PlanDetailResponse, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidatePlanTransition(p.Status, plan.PlanCompleted); err != nil {
		return nil, err
	}

	p.Status = plan.PlanCompleted
	now := time.Now()
	p.EndedAt = &now

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.toDetailResponse(p)
}

func (s *Service) Cancel(ctx context.Context, planID shared.ID) (*PlanDetailResponse, error) {
	p, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if err := plan.ValidatePlanTransition(p.Status, plan.PlanCancelled); err != nil {
		return nil, err
	}

	p.Status = plan.PlanCancelled

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.toDetailResponse(p)
}

func (s *Service) toDetailResponse(p *plan.TestPlan) (*PlanDetailResponse, error) {
	return &PlanDetailResponse{
		ID:         p.ID,
		ProjectID:  p.ProjectID,
		Name:       p.Name,
		Status:     p.Status,
		StartedAt:  p.StartedAt,
		PausedAt:   p.PausedAt,
		EndedAt:    p.EndedAt,
		CreatedBy:  p.CreatedBy,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}, nil
}
