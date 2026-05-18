package execution

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type ExecStatus string

const (
	ExecInProgress ExecStatus = "in_progress"
	ExecPaused     ExecStatus = "paused"
	ExecCompleted  ExecStatus = "completed"
	ExecCancelled  ExecStatus = "cancelled"
)

type ResultStatus string

const (
	ResultPassed  ResultStatus = "passed"
	ResultFailed  ResultStatus = "failed"
	ResultBlocked ResultStatus = "blocked"
	ResultSkipped ResultStatus = "skipped"
)

type TestExecution struct {
	ID          shared.ID
	PlanID      shared.ID
	Name        string
	Status      ExecStatus
	ExecutorID  shared.ID
	StartedAt   time.Time
	PausedAt    *time.Time
	CompletedAt *time.Time
	Notes       string
}

type ExecutionResult struct {
	ID          shared.ID
	ExecutionID shared.ID
	TestCaseID  shared.ID
	ExecutorID  shared.ID
	Status      ResultStatus
	BugID       string
	BugURL      string
	Notes       string
	ExecutedAt  time.Time
}

type ExecutionSummary struct {
	Total   int
	Passed  int
	Failed  int
	Blocked int
	Skipped int
}
