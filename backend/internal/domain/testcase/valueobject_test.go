// tasks.md: T012 | spec.md: §3.2 用例状态转换规则
package testcase

import (
	"testing"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidateCaseTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    CaseStatus
		to      CaseStatus
		wantErr bool
	}{
		// Valid transitions per spec.md §3.2
		{
			name:    "draft to ready - valid",
			from:    CaseDraft,
			to:      CaseReady,
			wantErr: false,
		},
		{
			name:    "draft to archived - valid",
			from:    CaseDraft,
			to:      CaseArchived,
			wantErr: false,
		},
		{
			name:    "ready to archived - valid",
			from:    CaseReady,
			to:      CaseArchived,
			wantErr: false,
		},
		{
			name:    "ready to draft - valid",
			from:    CaseReady,
			to:      CaseDraft,
			wantErr: false,
		},
		{
			name:    "archived to ready - valid",
			from:    CaseArchived,
			to:      CaseReady,
			wantErr: false,
		},
		// Invalid transitions
		{
			name:    "draft to draft - invalid same state",
			from:    CaseDraft,
			to:      CaseDraft,
			wantErr: true,
		},
		{
			name:    "archived to draft - invalid not allowed",
			from:    CaseArchived,
			to:      CaseDraft,
			wantErr: true,
		},
		{
			name:    "ready to ready - invalid same state",
			from:    CaseReady,
			to:      CaseReady,
			wantErr: true,
		},
		{
			name:    "archived to archived - invalid same state",
			from:    CaseArchived,
			to:      CaseArchived,
			wantErr: true,
		},
		{
			name:    "ready to draft - valid (round trip)",
			from:    CaseReady,
			to:      CaseDraft,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCaseTransition(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPriority_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Priority
		want bool
	}{
		{
			name: "P0 is valid",
			p:    P0,
			want: true,
		},
		{
			name: "P1 is valid",
			p:    P1,
			want: true,
		},
		{
			name: "P2 is valid",
			p:    P2,
			want: true,
		},
		{
			name: "P3 is valid",
			p:    P3,
			want: true,
		},
		{
			name: "negative priority is invalid",
			p:    Priority(-1),
			want: false,
		},
		{
			name: "priority > 3 is invalid",
			p:    Priority(4),
			want: false,
		},
		{
			name: "priority 10 is invalid",
			p:    Priority(10),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.p.Valid())
		})
	}
}

func TestNewID(t *testing.T) {
	t.Parallel()

	id := shared.NewID()
	assert.False(t, id.IsEmpty(), "NewID should not be empty")
}

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			s:       "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "invalid UUID format",
			s:       "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			s:       "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := shared.ParseID(tt.s)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, id.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, id.IsEmpty())
			}
		})
	}
}
