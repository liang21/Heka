// tasks.md: T091 | spec.md: AI provider manager with circuit breaker and failover
package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/liang21/heka/internal/domain/shared"
)

// Manager manages multiple AI providers with circuit breaker and automatic failover
type Manager struct {
	providers []ProviderConfig
	breakers  map[string]*CircuitBreaker
	clients   map[string]LLMClient
	mu        sync.RWMutex
}

// NewManager creates a new AI provider manager
func NewManager(providers []ProviderConfig) *Manager {
	m := &Manager{
		providers: providers,
		breakers:  make(map[string]*CircuitBreaker),
		clients:   make(map[string]LLMClient),
	}

	// Initialize providers and circuit breakers
	for _, cfg := range providers {
		// Initialize circuit breaker for each provider
		m.breakers[cfg.Name] = NewCircuitBreaker(5, DefaultResetTimeout)

		// Initialize client based on provider name
		var client LLMClient
		switch cfg.Name {
		case "claude":
			client = NewClaudeProvider(cfg.APIKey, cfg.BaseURL)
		case "openai":
			client = NewOpenAIProvider(cfg.APIKey, cfg.BaseURL)
		case "gemini":
			client = NewGeminiProvider(cfg.APIKey)
		case "ollama":
			client = NewOllamaProvider(cfg.BaseURL)
		default:
			continue // Skip unknown providers
		}

		m.clients[cfg.Name] = client
	}

	return m
}

// Chat sends a chat request using available providers with failover
func (m *Manager) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Sort providers by priority (lower number = higher priority)
	sortedProviders := make([]ProviderConfig, len(m.providers))
	copy(sortedProviders, m.providers)
	sort.Slice(sortedProviders, func(i, j int) bool {
		return sortedProviders[i].Priority < sortedProviders[j].Priority
	})

	var lastErr error
	for _, provider := range sortedProviders {
		client, exists := m.clients[provider.Name]
		if !exists {
			continue
		}

		breaker := m.breakers[provider.Name]

		// Use closure to capture response
		var result *ChatResponse
		execErr := breaker.Execute(ctx, func() error {
			return Retry(ctx, func() error {
				// Clone request to avoid modifying original
				chatReq := req
				if chatReq.Model == "" && provider.Model != "" {
					chatReq.Model = provider.Model
				}
				if chatReq.MaxTokens == 0 && provider.MaxTokens > 0 {
					chatReq.MaxTokens = provider.MaxTokens
				}
				if chatReq.Temperature == 0 && provider.Temperature > 0 {
					chatReq.Temperature = provider.Temperature
				}

				resp, err := client.Chat(ctx, chatReq)
				if err != nil {
					return err
				}

				result = resp
				return nil
			}, DefaultMaxAttempts)
		})

		if execErr == nil && result != nil {
			return result, nil
		}

		if execErr != nil {
			lastErr = execErr
		}

		// This provider failed, try next one
		continue
	}

	// All providers failed
	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", shared.ErrAIServiceUnavailable, lastErr)
	}

	return nil, shared.ErrAIServiceUnavailable
}

// GetProviderStatus returns the status of all providers
func (m *Manager) GetProviderStatus() map[string]State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]State)
	for name, breaker := range m.breakers {
		status[name] = breaker.GetState()
	}

	return status
}

// ResetBreaker manually resets a provider's circuit breaker
func (m *Manager) ResetBreaker(providerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.breakers[providerName]
	if !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	// Create new circuit breaker to reset
	m.breakers[providerName] = NewCircuitBreaker(5, DefaultResetTimeout)
	return nil
}
