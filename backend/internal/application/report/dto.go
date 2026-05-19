package report

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T123 | spec.md: §4.12 报告 DTO

type PlanReportResponse struct {
	PlanID      shared.ID               `json:"plan_id"`
	PlanName    string                  `json:"plan_name"`
	Status      string                  `json:"status"`
	TotalCases  int                     `json:"total_cases"`
	Passed      int                     `json:"passed"`
	Failed      int                     `json:"failed"`
	Blocked     int                     `json:"blocked"`
	Skipped     int                     `json:"skipped"`
	PassRate    float64                 `json:"pass_rate"`
	StartedAt   *time.Time              `json:"started_at"`
	CompletedAt *time.Time              `json:"completed_at"`
}

type CoverageResponse struct {
	ProjectID   shared.ID              `json:"project_id"`
	Modules     []ModuleCoverage       `json:"modules"`
	TotalCases  int                    `json:"total_cases"`
	CoveredCases int                   `json:"covered_cases"`
	CoverageRate float64               `json:"coverage_rate"`
}

type ModuleCoverage struct {
	ModuleID    shared.ID `json:"module_id"`
	ModuleName  string    `json:"module_name"`
	CaseCount   int       `json:"case_count"`
	ExecutedCount int     `json:"executed_count"`
}

type TrendResponse struct {
	ProjectID shared.ID    `json:"project_id"`
	Days      int          `json:"days"`
	Items     []TrendItem  `json:"items"`
}

type TrendItem struct {
	Date    string `json:"date"`
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Blocked int    `json:"blocked"`
	Skipped int    `json:"skipped"`
}

type BugDistributionResponse struct {
	ProjectID shared.ID         `json:"project_id"`
	Items     []BugDistItem     `json:"items"`
}

type BugDistItem struct {
	TestCaseID shared.ID `json:"test_case_id"`
	Title      string    `json:"title"`
	BugCount   int       `json:"bug_count"`
	BugIDs     []string  `json:"bug_ids"`
}

type WorkloadResponse struct {
	UserID   shared.ID      `json:"user_id"`
	UserName string         `json:"user_name"`
	Items    []WorkloadItem `json:"items"`
}

type WorkloadItem struct {
	PlanID      shared.ID `json:"plan_id"`
	PlanName    string    `json:"plan_name"`
	Assigned    int       `json:"assigned"`
	Completed   int       `json:"completed"`
	Passed      int       `json:"passed"`
	Failed      int       `json:"failed"`
}
