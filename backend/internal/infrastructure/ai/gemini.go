// tasks.md: T089 | spec.md: Gemini AI provider
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

// GeminiProvider implements Google Gemini GenerateContent API
// API Docs: https://ai.google.dev/api
type GeminiProvider struct {
	apiKey string
	client *http.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// geminiRequest represents Gemini API request structure
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// geminiResponse represents Gemini API response structure
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
			Role  string       `json:"role,omitempty"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
		Index        int    `json:"index"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to Gemini API
func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to Gemini format
	contents := make([]geminiContent, 0, len(req.Messages))
	for _, msg := range req.Messages {
		// Map roles: "user" -> "user", "assistant" -> "model"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	// Build request body
	body := geminiRequest{
		Contents: contents,
	}

	// Marshal request
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", req.Model, p.apiKey)
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
	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level error
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	// Extract content
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response candidates")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response parts")
	}

	content := geminiResp.Candidates[0].Content.Parts[0].Text
	tokensUsed := geminiResp.UsageMetadata.TotalTokenCount

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      req.Model,
	}, nil
}
