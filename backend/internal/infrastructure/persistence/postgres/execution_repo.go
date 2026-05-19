package postgres

import (
	"context"

	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

type ExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) execution.ExecutionRepository {
	return &ExecutionRepository{db: db}
}

func (r *ExecutionRepository) Create(ctx context.Context, exec *execution.TestExecution) error {
	db := DBOrTx(ctx, r.db)

	if err := db.Create(exec).Error; err != nil {
		// Handle partial unique index constraint violation
		// The idx_executions_single_active index only allows one in_progress execution per plan
		if isUniqueConstraintError(err, "") {
			return shared.ErrExecutionActiveConflict
		}
		return err
	}

	return nil
}

func (r *ExecutionRepository) FindByID(ctx context.Context, id shared.ID) (*execution.TestExecution, error) {
	db := DBOrTx(ctx, r.db)

	var exec execution.TestExecution
	if err := db.Where("id = ?", id).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrExecutionNotFound
		}
		return nil, err
	}

	return &exec, nil
}

func (r *ExecutionRepository) SubmitResult(ctx context.Context, result *execution.ExecutionResult) error {
	db := DBOrTx(ctx, r.db)

	if err := db.Create(result).Error; err != nil {
		// Handle unique constraint violation on (execution_id, test_case_id)
		if isUniqueConstraintError(err, "") {
			return shared.ErrExecutionConflict
		}
		return err
	}

	return nil
}

func (r *ExecutionRepository) BatchSubmitResults(ctx context.Context, results []*execution.ExecutionResult) error {
	if len(results) == 0 {
		return nil
	}

	db := DBOrTx(ctx, r.db)

	// Use CreateInBatches for efficient batch insert
	if err := db.CreateInBatches(results, 100).Error; err != nil {
		// Handle unique constraint violation
		if isUniqueConstraintError(err, "") {
			return shared.ErrExecutionConflict
		}
		return err
	}

	return nil
}

func (r *ExecutionRepository) GetSummary(ctx context.Context, executionID shared.ID) (*execution.ExecutionSummary, error) {
	db := DBOrTx(ctx, r.db)

	// First verify the execution exists
	var exec execution.TestExecution
	if err := db.Where("id = ?", executionID).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrExecutionNotFound
		}
		return nil, err
	}

	// Aggregate query to count results by status
	type ResultCount struct {
		Status string
		Count  int
	}

	var counts []ResultCount
	if err := db.Table("execution_results").
		Select("status, COUNT(*) as count").
		Where("execution_id = ?", executionID).
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, err
	}

	// Build summary from aggregation results
	summary := &execution.ExecutionSummary{}
	for _, c := range counts {
		summary.Total += c.Count
		switch execution.ResultStatus(c.Status) {
		case execution.ResultPassed:
			summary.Passed = c.Count
		case execution.ResultFailed:
			summary.Failed = c.Count
		case execution.ResultBlocked:
			summary.Blocked = c.Count
		case execution.ResultSkipped:
			summary.Skipped = c.Count
		}
	}

	return summary, nil
}
