package plan

import (
	"fmt"
)

type PlanStatus string

const (
	PlanDraft     PlanStatus = "draft"
	PlanActive    PlanStatus = "active"
	PlanPaused    PlanStatus = "paused"
	PlanCompleted PlanStatus = "completed"
	PlanCancelled PlanStatus = "cancelled"
)

// ValidatePlanTransition validates plan status transitions per spec.md §3.3:
// draft → active, cancelled
// active → paused, completed, cancelled
// paused → active, cancelled
// completed → (terminal state, no transitions allowed)
// cancelled → (terminal state, no transitions allowed)
func ValidatePlanTransition(from, to PlanStatus) error {
	if from == to {
		return fmt.Errorf("cannot transition from %s to same state", from)
	}

	// Terminal states - no outgoing transitions
	if from == PlanCompleted || from == PlanCancelled {
		return fmt.Errorf("cannot transition from terminal state %s", from)
	}

	validTransitions := map[PlanStatus][]PlanStatus{
		PlanDraft:  {PlanActive, PlanCancelled},
		PlanActive: {PlanPaused, PlanCompleted, PlanCancelled},
		PlanPaused: {PlanActive, PlanCancelled},
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
