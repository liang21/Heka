// tasks.md: T072 | spec.md: DOCX text extraction
package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/nguyenthenguyen/docx"
)

// DOCXParser implements text extraction from DOCX files
type DOCXParser struct{}

// SupportedTypes returns the file types supported by this parser
func (p *DOCXParser) SupportedTypes() []string {
	return []string{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
}

// Parse extracts text content from a DOCX file
func (p *DOCXParser) Parse(ctx context.Context, r io.Reader) (string, error) {
	// Check context
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Check for nil reader
	if r == nil {
		return "", fmt.Errorf("reader is nil")
	}

	// Read content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX: %w", err)
	}

	// Create ReaderAt and size
	readerAt := bytes.NewReader(content)

	// Parse DOCX
	doc, err := docx.ReadDocxFromMemory(readerAt, readerAt.Size())
	if err != nil {
		return "", fmt.Errorf("failed to parse DOCX: %w", err)
	}
	defer doc.Close()

	// Get the editable content
	editable := doc.Editable()
	text := editable.GetContent()

	return text, nil
}
