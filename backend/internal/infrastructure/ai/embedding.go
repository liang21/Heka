// tasks.md: T094 | spec.md: Embedding client for vector search
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

// EmbeddingClient handles text embedding generation for vector search
// Supports OpenAI embeddings API (can be extended to other providers)
type EmbeddingClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
	dimension  int
}

// NewEmbeddingClient creates a new embedding client
func NewEmbeddingClient(apiKey, model string) *EmbeddingClient {
	baseURL := "https://api.openai.com"
	dimension := 1536 // Default for text-embedding-ada-002

	// Adjust dimension based on model
	switch model {
	case "text-embedding-ada-002":
		dimension = 1536
	case "text-embedding-3-small":
		dimension = 1536
	case "text-embedding-3-large":
		dimension = 3072
	}

	return &EmbeddingClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:    apiKey,
		baseURL:   baseURL,
		model:     model,
		dimension: dimension,
	}
}

// embeddingRequest represents OpenAI embeddings API request
type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// embeddingResponse represents OpenAI embeddings API response
type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Embed generates embeddings for a list of texts
func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("%w: no texts provided", shared.ErrAIInvalidInput)
	}

	// Sanitize inputs
	sanitizedTexts := make([]string, len(texts))
	for i, text := range texts {
		sanitizedTexts[i] = SanitizeInput(text)
	}

	// Build request body
	body := embeddingRequest{
		Input: sanitizedTexts,
		Model: c.model,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/embeddings", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute request
	resp, err := c.httpClient.Do(req)
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
	var embedResp embeddingResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level error
	if embedResp.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", embedResp.Error.Message)
	}

	// Extract embeddings
	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Sort results by index to ensure correct order
	embeddings := make([][]float32, len(embedResp.Data))
	for _, item := range embedResp.Data {
		if item.Index < 0 || item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index: %d", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// EmbedSingle generates embedding for a single text
func (c *EmbeddingClient) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := c.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}

	return embeddings[0], nil
}

// GetDimension returns the embedding dimension
func (c *EmbeddingClient) GetDimension() int {
	return c.dimension
}

// GetModel returns the model name
func (c *EmbeddingClient) GetModel() string {
	return c.model
}

// BatchEmbed embeds texts in batches to handle large inputs
func (c *EmbeddingClient) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("%w: no texts provided", shared.ErrAIInvalidInput)
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := c.Embed(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d failed: %w", i/batchSize, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// CalculateCosineSimilarity calculates cosine similarity between two embeddings
func CalculateCosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("embedding dimensions mismatch: %d vs %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("empty embeddings")
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero norm embedding")
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB)), nil
}

// sqrt32 calculates square root for float32
func sqrt32(x float32) float32 {
	return float32(sqrt(float64(x)))
}

// sqrt is a simple square root implementation
func sqrt(x float64) float64 {
	// Newton's method
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// ValidateEmbedding checks if an embedding is valid
func ValidateEmbedding(embedding []float32, expectedDim int) error {
	if embedding == nil {
		return fmt.Errorf("nil embedding")
	}

	if len(embedding) != expectedDim {
		return fmt.Errorf("invalid embedding dimension: expected %d, got %d", expectedDim, len(embedding))
	}

	// Check for NaN or Inf
	for _, v := range embedding {
		if v != v { // NaN check
			return fmt.Errorf("embedding contains NaN")
		}
	}

	return nil
}
