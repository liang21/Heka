package execution

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type ExecutionRepository interface {
	Create(ctx context.Context, exec *TestExecution) error
	FindByID(ctx context.Context, id shared.ID) (*TestExecution, error)
	SubmitResult(ctx context.Context, result *ExecutionResult) error
	BatchSubmitResults(ctx context.Context, results []*ExecutionResult) error
	GetSummary(ctx context.Context, executionID shared.ID) (*ExecutionSummary, error)
}
