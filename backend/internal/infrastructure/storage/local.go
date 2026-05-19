// tasks.md: T066 | spec.md: Local file storage implementation
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

// localStorage implements file storage using local filesystem
type localStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage implementation
func NewLocalStorage(basePath string) *localStorage {
	return &localStorage{
		basePath: basePath,
	}
}

// Save saves a file to local storage
func (s *localStorage) Save(ctx context.Context, projectID shared.ID, filename string, reader io.Reader) (string, error) {
	// Sanitize filename to prevent path traversal
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	// filepath.Base strips any directory components
	cleanFilename := filepath.Base(filename)
	if cleanFilename == "." || cleanFilename == "/" {
		return "", fmt.Errorf("invalid filename")
	}

	// Create directory structure: {basePath}/{project_id/YYYY/MM/}
	now := time.Now()
	projectPath := filepath.Join(s.basePath, projectID.String())
	datePath := filepath.Join(projectPath, now.Format("2006/01"))

	if err := os.MkdirAll(datePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename to avoid collisions
	uniqueFilename := fmt.Sprintf("%d_%s", now.UnixNano(), cleanFilename)
	fullPath := filepath.Join(datePath, uniqueFilename)

	// Verify the final path is within basePath (defense in depth)
	relPath, err := filepath.Rel(s.basePath, fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if relPath[0] == '.' {
		return "", fmt.Errorf("path traversal detected")
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// GetPath returns the full path for a given filename
func (s *localStorage) GetPath(filename string) string {
	return filepath.Join(s.basePath, filename)
}

// Delete removes a file from storage
func (s *localStorage) Delete(ctx context.Context, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Sanitize to prevent path traversal
	cleanFilename := filepath.Base(filename)
	if cleanFilename == "." || cleanFilename == "/" {
		return fmt.Errorf("invalid filename")
	}

	fullPath := s.GetPath(filename)

	// Verify path is within basePath
	relPath, err := filepath.Rel(s.basePath, fullPath)
	if err != nil || (relPath[0] == '.') {
		return fmt.Errorf("invalid path")
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
