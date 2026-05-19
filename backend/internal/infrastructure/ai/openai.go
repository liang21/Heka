// tasks.md: T088 | spec.md: OpenAI provider
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

// OpenAIProvider implements OpenAI Chat Completions API
// API Docs: https://platform.openai.com/docs/api-reference/chat/create
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// openaiRequest represents OpenAI API request structure
type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openaiResponse represents OpenAI API response structure
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to OpenAI API
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to OpenAI format
	messages := make([]openaiMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request body
	body := openaiRequest{
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
	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

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
	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level error
	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	// Extract content
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response choices")
	}

	content := openaiResp.Choices[0].Message.Content
	tokensUsed := openaiResp.Usage.TotalTokens

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      openaiResp.Model,
	}, nil
}
