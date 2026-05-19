// tasks.md: T087 | spec.md: Claude AI provider
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

// ClaudeProvider implements Anthropic Claude API
// API Docs: https://docs.anthropic.com/claude/reference/messages_post
type ClaudeProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey, baseURL string) *ClaudeProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	return &ClaudeProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// claudeRequest represents Claude API request structure
type claudeRequest struct {
	Model       string      `json:"model"`
	MaxTokens   int         `json:"max_tokens"`
	Messages    []claudeMsg `json:"messages"`
	System      string      `json:"system,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
	Temperature float64     `json:"temperature,omitempty"`
}

type claudeMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse represents Claude API response structure
type claudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to Claude API
func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to Claude format
	messages := make([]claudeMsg, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = claudeMsg{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request body
	body := claudeRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Messages:    messages,
		Temperature: req.Temperature,
		Stream:      false,
	}

	// Marshal request
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/messages", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level error
	if claudeResp.Error != nil {
		return nil, fmt.Errorf("Claude API error: %s", claudeResp.Error.Message)
	}

	// Extract content
	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty response content")
	}

	content := claudeResp.Content[0].Text
	tokensUsed := claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      claudeResp.Model,
	}, nil
}
