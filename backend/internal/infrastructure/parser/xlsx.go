// tasks.md: T074 | spec.md: XLSX text extraction
package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// XLSXParser implements text extraction from XLSX files
type XLSXParser struct{}

// SupportedTypes returns the file types supported by this parser
func (p *XLSXParser) SupportedTypes() []string {
	return []string{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}
}

// Parse extracts text content from an XLSX file
func (p *XLSXParser) Parse(ctx context.Context, r io.Reader) (string, error) {
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
		return "", fmt.Errorf("failed to read XLSX: %w", err)
	}

	// Open XLSX file
	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to open XLSX: %w", err)
	}
	defer f.Close()

	// Get all sheets
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", nil
	}

	var text strings.Builder

	// Extract data from each sheet
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue // Skip sheets that fail
		}

		// Write sheet header
		text.WriteString(fmt.Sprintf("--- Sheet: %s ---\n", sheet))

		// Write each row
		for _, row := range rows {
			rowText := strings.Join(row, " | ")
			if rowText != "" {
				text.WriteString(rowText + "\n")
			}
		}

		text.WriteString("\n")
	}

	return text.String(), nil
}
