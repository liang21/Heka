// tasks.md: T076 | spec.md: Image OCR text extraction
package parser

import (
	"context"
	"fmt"
	"io"

	"github.com/otiai10/gosseract/v2"
)

// ImageParser implements OCR text extraction from images
type ImageParser struct{}

// SupportedTypes returns the file types supported by this parser
func (p *ImageParser) SupportedTypes() []string {
	return []string{"png", "jpg", "jpeg", "gif", "bmp", "tiff"}
}

// Parse extracts text content from an image using OCR
func (p *ImageParser) Parse(ctx context.Context, r io.Reader) (string, error) {
	// Check context
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Check for nil reader
	if r == nil {
		return "", fmt.Errorf("reader is nil")
	}

	// Read image content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Create OCR client
	client := gosseract.NewClient()
	defer client.Close()

	// Set image from bytes
	if err := client.SetImageFromBytes(content); err != nil {
		return "", fmt.Errorf("failed to set image for OCR: %w", err)
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("OCR extraction failed: %w", err)
	}

	return text, nil
}
