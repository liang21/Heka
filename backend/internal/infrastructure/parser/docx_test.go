// tasks.md: T071 | spec.md: DOCX text extraction
package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDOCX(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reg := NewRegistry()
	parser, err := reg.GetParser("docx")
	assert.NoError(t, err)

	// This will fail because docx.go doesn't exist yet - TDD RED
	assert.NotNil(t, parser)
	_, err = parser.Parse(ctx, nil)
	assert.Error(t, err) // Expected: no implementation yet
}
