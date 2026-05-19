// tasks.md: T067 | spec.md: Figma API client for document extraction
package figma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client wraps the Figma REST API
type Client struct {
	accessToken string
	httpClient  *http.Client
	baseURL     string
}

// NewClient creates a new Figma API client
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient:  &http.Client{},
		baseURL:     "https://api.figma.com/v1",
	}
}

// FigmaDocument represents a Figma file document structure
type FigmaDocument struct {
	Root    *Node `json:"document"`
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// Node represents a node in the Figma document tree
type Node struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Characters string  `json:"characters,omitempty"`
	Children   []*Node `json:"children,omitempty"`
}

// GetFile fetches a Figma file by its URL or key
func (c *Client) GetFile(ctx context.Context, fileURL string) (*FigmaDocument, error) {
	// Extract file key from URL
	// URL format: https://www.figma.com/file/{key}/{title}
	fileKey := c.extractFileKey(fileURL)
	if fileKey == "" {
		return nil, fmt.Errorf("invalid Figma file URL: %s", fileURL)
	}

	// Build request URL
	url := fmt.Sprintf("%s/files/%s", c.baseURL, fileKey)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Figma-Token", c.accessToken)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Figma API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var doc FigmaDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &doc, nil
}

// ExtractText extracts all text content from a Figma document
func (c *Client) ExtractText(doc *FigmaDocument) (string, error) {
	if doc == nil || doc.Root == nil {
		return "", fmt.Errorf("empty document")
	}

	var texts []string
	c.extractTextFromNode(doc.Root, &texts)

	return strings.Join(texts, "\n\n"), nil
}

// extractTextFromNode recursively extracts text from nodes
func (c *Client) extractTextFromNode(node *Node, texts *[]string) {
	if node == nil {
		return
	}

	// Extract text from TEXT nodes
	if node.Type == "TEXT" && node.Characters != "" {
		text := strings.TrimSpace(node.Characters)
		if text != "" {
			*texts = append(*texts, text)
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		c.extractTextFromNode(child, texts)
	}
}

// extractFileKey extracts the file key from a Figma URL
func (c *Client) extractFileKey(fileURL string) string {
	// Remove trailing slash
	fileURL = strings.TrimSuffix(fileURL, "/")

	// Split by /
	parts := strings.Split(fileURL, "/")
	for i, part := range parts {
		if part == "file" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}
