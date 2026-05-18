// tasks.md: T014 | spec.md: §3.3 计划状态转换
package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePlanTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    PlanStatus
		to      PlanStatus
		wantErr bool
	}{
		// Valid transitions per spec.md §3.3
		{
			name:    "draft to active - valid",
			from:    PlanDraft,
			to:      PlanActive,
			wantErr: false,
		},
		{
			name:    "draft to cancelled - valid",
			from:    PlanDraft,
			to:      PlanCancelled,
			wantErr: false,
		},
		{
			name:    "active to paused - valid",
			from:    PlanActive,
			to:      PlanPaused,
			wantErr: false,
		},
		{
			name:    "active to completed - valid",
			from:    PlanActive,
			to:      PlanCompleted,
			wantErr: false,
		},
		{
			name:    "active to cancelled - valid",
			from:    PlanActive,
			to:      PlanCancelled,
			wantErr: false,
		},
		{
			name:    "paused to active - valid",
			from:    PlanPaused,
			to:      PlanActive,
			wantErr: false,
		},
		{
			name:    "paused to cancelled - valid",
			from:    PlanPaused,
			to:      PlanCancelled,
			wantErr: false,
		},
		// Invalid transitions - terminal states
		{
			name:    "completed to active - invalid terminal state",
			from:    PlanCompleted,
			to:      PlanActive,
			wantErr: true,
		},
		{
			name:    "completed to draft - invalid terminal state",
			from:    PlanCompleted,
			to:      PlanDraft,
			wantErr: true,
		},
		{
			name:    "completed to paused - invalid terminal state",
			from:    PlanCompleted,
			to:      PlanPaused,
			wantErr: true,
		},
		{
			name:    "cancelled to draft - invalid terminal state",
			from:    PlanCancelled,
			to:      PlanDraft,
			wantErr: true,
		},
		{
			name:    "cancelled to active - invalid terminal state",
			from:    PlanCancelled,
			to:      PlanActive,
			wantErr: true,
		},
		{
			name:    "cancelled to paused - invalid terminal state",
			from:    PlanCancelled,
			to:      PlanPaused,
			wantErr: true,
		},
		// Invalid transitions - not allowed
		{
			name:    "draft to paused - invalid not allowed",
			from:    PlanDraft,
			to:      PlanPaused,
			wantErr: true,
		},
		{
			name:    "draft to completed - invalid not allowed",
			from:    PlanDraft,
			to:      PlanCompleted,
			wantErr: true,
		},
		{
			name:    "draft to draft - invalid same state",
			from:    PlanDraft,
			to:      PlanDraft,
			wantErr: true,
		},
		{
			name:    "active to draft - invalid not allowed",
			from:    PlanActive,
			to:      PlanDraft,
			wantErr: true,
		},
		{
			name:    "active to active - invalid same state",
			from:    PlanActive,
			to:      PlanActive,
			wantErr: true,
		},
		{
			name:    "paused to draft - invalid not allowed",
			from:    PlanPaused,
			to:      PlanDraft,
			wantErr: true,
		},
		{
			name:    "paused to completed - invalid not allowed",
			from:    PlanPaused,
			to:      PlanCompleted,
			wantErr: true,
		},
		{
			name:    "paused to paused - invalid same state",
			from:    PlanPaused,
			to:      PlanPaused,
			wantErr: true,
		},
		{
			name:    "completed to completed - invalid same state",
			from:    PlanCompleted,
			to:      PlanCompleted,
			wantErr: true,
		},
		{
			name:    "cancelled to cancelled - invalid same state",
			from:    PlanCancelled,
			to:      PlanCancelled,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePlanTransition(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
