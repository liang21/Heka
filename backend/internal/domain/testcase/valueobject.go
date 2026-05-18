package testcase

import (
	"fmt"
)

type Priority int

const (
	P0 Priority = 0
	P1 Priority = 1
	P2 Priority = 2
	P3 Priority = 3
)

func (p Priority) Valid() bool {
	return p >= 0 && p <= 3
}

type CaseStatus string

const (
	CaseDraft    CaseStatus = "draft"
	CaseReady    CaseStatus = "ready"
	CaseArchived CaseStatus = "archived"
)

// ValidateCaseTransition validates status transitions per spec.md §3.2:
// draft → ready, archived
// ready → archived, draft
// archived → ready
func ValidateCaseTransition(from, to CaseStatus) error {
	if from == to {
		return fmt.Errorf("cannot transition from %s to same state", from)
	}

	validTransitions := map[CaseStatus][]CaseStatus{
		CaseDraft:    {CaseReady, CaseArchived},
		CaseReady:    {CaseArchived, CaseDraft},
		CaseArchived: {CaseReady},
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("unknown source state: %s", from)
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return nil
		}
	}

	return fmt.Errorf("invalid transition from %s to %s", from, to)
}
