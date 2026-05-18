// tasks.md: T070 | spec.md: PDF text extraction
package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ledongthuc/pdf"
)

// PDFParser implements text extraction from PDF files
type PDFParser struct{}

// SupportedTypes returns the file types supported by this parser
func (p *PDFParser) SupportedTypes() []string {
	return []string{"pdf", "application/pdf"}
}

// Parse extracts text content from a PDF file
func (p *PDFParser) Parse(ctx context.Context, r io.Reader) (string, error) {
	// Check context
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Check for nil reader
	if r == nil {
		return "", fmt.Errorf("reader is nil")
	}

	// Read all content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Create PDF reader - requires ReaderAt and size
	readerAt := bytes.NewReader(content)
	pdfReader, err := pdf.NewReader(readerAt, readerAt.Size())
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Extract text from all pages
	var text string
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)

		// Get plain text with empty font map
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue // Skip pages that fail
		}

		text += pageText + "\n\n"
	}

	return text, nil
}
