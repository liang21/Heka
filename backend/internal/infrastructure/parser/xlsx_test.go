// tasks.md: T073 | spec.md: XLSX text extraction
package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseXLSX(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reg := NewRegistry()
	parser, err := reg.GetParser("xlsx")
	assert.NoError(t, err)

	// This will fail because xlsx.go doesn't exist yet - TDD RED
	assert.NotNil(t, parser)
	_, err = parser.Parse(ctx, nil)
	assert.Error(t, err) // Expected: no implementation yet
}
