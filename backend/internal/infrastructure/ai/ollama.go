// tasks.md: T090 | spec.md: Ollama AI provider
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements Ollama HTTP API
// API Docs: https://github.com/ollama/ollama/blob/main/docs/api.md
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // Ollama can be slower
		},
	}
}

// ollamaRequest represents Ollama API request structure
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// ollamaResponse represents Ollama API response structure
type ollamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// Chat sends a chat request to Ollama API
func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to Ollama format
	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request body
	body := ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
		Options: &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
		},
	}

	// Marshal request
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level error
	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
	}

	// Extract content
	content := ollamaResp.Message.Content

	return &ChatResponse{
		Content:    content,
		TokensUsed: 0, // Ollama doesn't return token count
		Model:      ollamaResp.Model,
	}, nil
}
