package plan

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type TestPlanRepository interface {
	Create(ctx context.Context, plan *TestPlan) error
	FindByID(ctx context.Context, id shared.ID) (*TestPlan, error)
	List(ctx context.Context, projectID shared.ID, status *PlanStatus, page, pageSize int) ([]*TestPlan, int64, error)
	Update(ctx context.Context, plan *TestPlan) error
	AddCases(ctx context.Context, planID shared.ID, cases []PlanTestCase) error
	RemoveCases(ctx context.Context, planID shared.ID, caseIDs []shared.ID) error
}
