// tasks.md: T069 | spec.md: PDF text extraction
package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePDF(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reg := NewRegistry()
	parser, err := reg.GetParser("pdf")
	assert.NoError(t, err)

	// This will fail because pdf.go doesn't exist yet - TDD RED
	reader := strings.NewReader("%PDF-1.4 test content")
	_, err = parser.Parse(ctx, reader)
	assert.Error(t, err) // Will fail with "no parser registered" initially
}
