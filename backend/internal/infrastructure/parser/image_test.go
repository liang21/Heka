// tasks.md: T075 | spec.md: Image OCR text extraction
package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reg := NewRegistry()
	parser, err := reg.GetParser("image")
	assert.NoError(t, err)

	// This will fail because image.go doesn't exist yet - TDD RED
	assert.NotNil(t, parser)
	_, err = parser.Parse(ctx, nil)
	assert.Error(t, err) // Expected: no implementation yet
}
