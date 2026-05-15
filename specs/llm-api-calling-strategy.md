# Heka LLM API 调用技术方案

> 完整的 LLM API 调用容错方案，涵盖线程管理、容错、超时、重试四个维度

## 问题分析

| 问题 | 影响 | 解决方向 |
|------|------|----------|
| 调用慢（10-60s） | 阻塞用户请求，超时风险 | 异步化 + 超时控制 |
| 不稳定（失败率 5-20%） | 用户体验差，功能不可用 | 重试 + 故障转移 + 熔断 |
| 多 Provider 成本/性能差异 | 成本优化，体验提升 | 智能路由，降级策略 |

---

## 1. 线程管理（Goroutine Pool）

**方案**：Worker Pool 模式，控制并发 LLM 调用数量。

### 1.1 Pool 实现

```go
// internal/infrastructure/ai/pool/pool.go
package pool

import (
    "context"
    "errors"
    "sync"
    "time"
)

// Task LLM 调用任务
type Task func(ctx context.Context) error

// Pool Worker Pool 管理 goroutine
type Pool struct {
    maxWorkers int
    taskQueue  chan Task
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

// NewPool 创建 Worker Pool
func NewPool(maxWorkers int, queueSize int) *Pool {
    ctx, cancel := context.WithCancel(context.Background())

    p := &Pool{
        maxWorkers: maxWorkers,
        taskQueue:  make(chan Task, queueSize),
        ctx:        ctx,
        cancel:     cancel,
    }

    // 启动 worker
    for i := 0; i < maxWorkers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }

    return p
}

// worker 处理任务
func (p *Pool) worker(id int) {
    defer p.wg.Done()

    for {
        select {
        case task := <-p.taskQueue:
            // 执行任务
            if err := task(p.ctx); err != nil {
                // 记录错误，但不中断 worker
                // 可以添加日志或指标
            }
        case <-p.ctx.Done():
            return
        }
    }
}

// Submit 提交任务（异步）
func (p *Pool) Submit(ctx context.Context, task Task) error {
    select {
    case p.taskQueue <- task:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-p.ctx.Done():
        return ErrPoolClosed
    }
}

// SubmitAndWait 提交任务并等待完成
func (p *Pool) SubmitAndWait(ctx context.Context, task Task) error {
    errCh := make(chan error, 1)

    wrappedTask := func(taskCtx context.Context) error {
        errCh <- task(taskCtx)
        return nil
    }

    if err := p.Submit(ctx, wrappedTask); err != nil {
        return err
    }

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}

// SubmitWithResult 提交任务并返回结果
func (p *Pool) SubmitWithResult(ctx context.Context, task func(context.Context) (interface{}, error)) (interface{}, error) {
    resultCh := make(chan result, 1)

    wrappedTask := func(taskCtx context.Context) error {
        res, err := task(taskCtx)
        resultCh <- result{value: res, err: err}
        return nil
    }

    if err := p.Submit(ctx, wrappedTask); err != nil {
        return nil, err
    }

    select {
    case res := <-resultCh:
        return res.value, res.err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

type result struct {
    value interface{}
    err   error
}

// Close 关闭 Pool
func (p *Pool) Close() error {
    p.cancel()
    p.wg.Wait()
    close(p.taskQueue)
    return nil
}

// Stats 获取 Pool 状态
func (p *Pool) Stats() PoolStats {
    return PoolStats{
        QueueLength: len(p.taskQueue),
        MaxWorkers:  p.maxWorkers,
        ActiveTasks: p.maxWorkers - len(p.taskQueue),
    }
}

// PoolStats Pool 状态
type PoolStats struct {
    QueueLength int `json:"queue_length"`
    MaxWorkers  int `json:"max_workers"`
    ActiveTasks int `json:"active_tasks"`
}

var ErrPoolClosed = errors.New("pool is closed")
```

### 1.2 配置

```go
// internal/shared/config/ai.go
package config

import "time"

type AIConfig struct {
    // Worker Pool 配置
    PoolMaxWorkers int `json:"pool_max_workers" mapstructure:"pool_max_workers"`
    PoolQueueSize  int `json:"pool_queue_size" mapstructure:"pool_queue_size"`

    // 重试配置
    RetryMaxAttempts int           `json:"retry_max_attempts" mapstructure:"retry_max_attempts"`
    RetryBaseDelay   time.Duration `json:"retry_base_delay" mapstructure:"retry_base_delay"`
    RetryMaxDelay    time.Duration `json:"retry_max_delay" mapstructure:"retry_max_delay"`
    RetryMultiplier  float64       `json:"retry_multiplier" mapstructure:"retry_multiplier"`

    // 超时配置
    TimeoutDial           time.Duration `json:"timeout_dial" mapstructure:"timeout_dial"`
    TimeoutTLSHandshake   time.Duration `json:"timeout_tls_handshake" mapstructure:"timeout_tls_handshake"`
    TimeoutResponseHeader time.Duration `json:"timeout_response_header" mapstructure:"timeout_response_header"`
    TimeoutRequest        time.Duration `json:"timeout_request" mapstructure:"timeout_request"`
    TimeoutGeneration     time.Duration `json:"timeout_generation" mapstructure:"timeout_generation"`

    // Provider 配置
    Providers []ProviderConfig `json:"providers" mapstructure:"providers"`
}

type ProviderConfig struct {
    Name     Provider `json:"name" mapstructure:"name"`
    APIKey   string   `json:"api_key" mapstructure:"api_key"`
    BaseURL  string   `json:"base_url" mapstructure:"base_url"`
    Model    string   `json:"model" mapstructure:"model"`
    Priority int      `json:"priority" mapstructure:"priority"` // 优先级，越小越高
    Enabled  bool     `json:"enabled" mapstructure:"enabled"`
}

type Provider string

const (
    ProviderOpenAI Provider = "openai"
    ProviderClaude Provider = "claude"
    ProviderGemini Provider = "gemini"
    ProviderOllama Provider = "ollama"
)

// DefaultAIConfig 默认配置
func DefaultAIConfig() AIConfig {
    return AIConfig{
        PoolMaxWorkers:        10,                   // 最多 10 个并发 LLM 调用
        PoolQueueSize:         100,                  // 队列长度 100
        RetryMaxAttempts:      3,                    // 最多重试 3 次
        RetryBaseDelay:        1 * time.Second,      // 基础延迟 1 秒
        RetryMaxDelay:         30 * time.Second,     // 最大延迟 30 秒
        RetryMultiplier:       2.0,                  // 延迟倍数 2
        TimeoutDial:           10 * time.Second,
        TimeoutTLSHandshake:   5 * time.Second,
        TimeoutResponseHeader: 30 * time.Second,
        TimeoutRequest:        60 * time.Second,
        TimeoutGeneration:     55 * time.Second,
    }
}
```

---

## 2. 容错（Circuit Breaker + Failover）

**方案**：熔断器 + 多 Provider 故障转移。

### 2.1 熔断器实现

```go
// internal/infrastructure/ai/circuit/breaker.go
package circuit

import (
    "fmt"
    "sync"
    "time"
)

// State 熔断器状态
type State int

const (
    StateClosed State = iota // 关闭：正常工作
    StateOpen               // 打开：熔断中
    StateHalfOpen           // 半开：试探恢复
)

func (s State) String() string {
    switch s {
    case StateClosed:
        return "closed"
    case StateOpen:
        return "open"
    case StateHalfOpen:
        return "half_open"
    default:
        return "unknown"
    }
}

// Breaker 熔断器
type Breaker struct {
    mu               sync.Mutex
    state            State
    failureCount     int
    successCount     int
    lastFailureTime  time.Time
    lastSuccessTime  time.Time

    // 配置
    threshold        int           // 失败阈值
    timeout          time.Duration // 熔断后等待时间
    halfOpenMaxCalls int           // 半开状态最大尝试次数
}

// NewBreaker 创建熔断器
func NewBreaker(threshold int, timeout time.Duration) *Breaker {
    return &Breaker{
        state:            StateClosed,
        threshold:        threshold,
        timeout:          timeout,
        halfOpenMaxCalls: 3,
    }
}

// Allow 执行前检查是否允许
func (b *Breaker) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    switch b.state {
    case StateClosed:
        return true

    case StateOpen:
        // 检查是否超时
        if time.Since(b.lastFailureTime) > b.timeout {
            b.state = StateHalfOpen
            b.successCount = 0
            return true
        }
        return false

    case StateHalfOpen:
        return b.successCount < b.halfOpenMaxCalls

    default:
        return false
    }
}

// RecordSuccess 记录成功
func (b *Breaker) RecordSuccess() {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.successCount++
    b.lastSuccessTime = time.Now()

    if b.state == StateHalfOpen {
        if b.successCount >= b.halfOpenMaxCalls {
            b.state = StateClosed
            b.failureCount = 0
        }
    } else if b.state == StateClosed {
        // 重置失败计数
        b.failureCount = 0
    }
}

// RecordFailure 记录失败
func (b *Breaker) RecordFailure() {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.failureCount++
    b.lastFailureTime = time.Now()

    if b.failureCount >= b.threshold {
        b.state = StateOpen
    }
}

// GetState 获取当前状态
func (b *Breaker) GetState() State {
    b.mu.Lock()
    defer b.mu.Unlock()
    return b.state
}

// GetStats 获取熔断器统计信息
func (b *Breaker) GetStats() BreakerStats {
    b.mu.Lock()
    defer b.mu.Unlock()

    return BreakerStats{
        State:           b.state.String(),
        FailureCount:    b.failureCount,
        SuccessCount:    b.successCount,
        LastFailureTime: b.lastFailureTime,
        LastSuccessTime: b.lastSuccessTime,
    }
}

type BreakerStats struct {
    State           string    `json:"state"`
    FailureCount    int       `json:"failure_count"`
    SuccessCount    int       `json:"success_count"`
    LastFailureTime time.Time `json:"last_failure_time"`
    LastSuccessTime time.Time `json:"last_success_time"`
}

// Reset 重置熔断器
func (b *Breaker) Reset() {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.state = StateClosed
    b.failureCount = 0
    b.successCount = 0
}
```

### 2.2 多 Provider 故障转移

```go
// internal/infrastructure/ai/multi_provider.go
package ai

import (
    "context"
    "fmt"
    "sort"
    "sync"
    "time"

    "heka-backend/internal/infrastructure/ai/circuit"
    "heka-backend/internal/infrastructure/ai/pool"
)

// LLMClient 统一 LLM 客户端接口
type LLMClient interface {
    // Chat 对话
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    // GenerateText 生成文本
    GenerateText(ctx context.Context, prompt string) (string, error)
}

// ChatRequest 聊天请求
type ChatRequest struct {
    Messages    []Message  `json:"messages"`
    Model       string     `json:"model"`
    MaxTokens   int        `json:"max_tokens"`
    Temperature float64    `json:"temperature"`
    Stream      bool       `json:"stream"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
    Content string `json:"content"`
    Usage   Usage  `json:"usage"`
    Model   string `json:"model"`
}

// Usage Token 使用情况
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// Message 消息
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// MultiProviderClient 多 Provider 客户端
type MultiProviderClient struct {
    providers      map[Provider]LLMClient
    providerConfig []ProviderConfig
    breakers       map[Provider]*circuit.Breaker
    pool           *pool.Pool
    mu             sync.RWMutex
}

// NewMultiProviderClient 创建多 Provider 客户端
func NewMultiProviderClient(cfgs []ProviderConfig, pool *pool.Pool) *MultiProviderClient {
    m := &MultiProviderClient{
        providers:      make(map[Provider]LLMClient),
        providerConfig: cfgs,
        breakers:       make(map[Provider]*circuit.Breaker),
        pool:           pool,
    }

    // 初始化 Provider
    for _, cfg := range cfgs {
        if !cfg.Enabled {
            continue
        }

        var client LLMClient
        switch cfg.Name {
        case ProviderOpenAI:
            client = NewOpenAIClient(cfg.APIKey, cfg.BaseURL)
        case ProviderClaude:
            client = NewClaudeClient(cfg.APIKey, cfg.BaseURL)
        case ProviderGemini:
            client = NewGeminiClient(cfg.APIKey, cfg.BaseURL)
        case ProviderOllama:
            client = NewOllamaClient(cfg.BaseURL)
        }

        if client != nil {
            m.providers[cfg.Name] = client
            // 每个独立的熔断器
            m.breakers[cfg.Name] = circuit.NewBreaker(5, 30*time.Second)
        }
    }

    return m
}

// Chat 发起对话（自动故障转移）
func (m *MultiProviderClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // 按 Priority 排序
    sortedProviders := m.getSortedProviders()

    var lastErr error
    attemptedProviders := []Provider{}

    for _, cfg := range sortedProviders {
        provider := cfg.Name
        attemptedProviders = append(attemptedProviders, provider)

        client, exists := m.providers[provider]
        if !exists {
            continue
        }

        breaker := m.breakers[provider]

        // 检查熔断器
        if !breaker.Allow() {
            lastErr = fmt.Errorf("provider %s is circuit open", provider)
            continue
        }

        // 执行调用（在 Pool 中）
        var resp *ChatResponse
        var err error

        poolErr := m.pool.SubmitAndWait(ctx, func(poolCtx context.Context) error {
            resp, err = m.callWithTimeout(poolCtx, client, req)
            return err
        })

        if poolErr != nil {
            breaker.RecordFailure()
            lastErr = fmt.Errorf("provider %s failed: %w", provider, poolErr)
            continue
        }

        // 成功
        breaker.RecordSuccess()
        return resp, nil
    }

    // 所有 Provider 都失败
    return nil, fmt.Errorf("all providers failed (attempted: %v), last error: %w", attemptedProviders, lastErr)
}

// callWithTimeout 调用带超时
func (m *MultiProviderClient) callWithTimeout(ctx context.Context, client LLMClient, req ChatRequest) (*ChatResponse, error) {
    // 使用 context.WithTimeout 确保整体超时
    timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()

    return client.Chat(timeoutCtx, req)
}

// GenerateText 生成文本（自动故障转移）
func (m *MultiProviderClient) GenerateText(ctx context.Context, prompt string) (string, error) {
    req := ChatRequest{
        Messages: []Message{
            {Role: "user", Content: prompt},
        },
        MaxTokens: 4096,
    }

    resp, err := m.Chat(ctx, req)
    if err != nil {
        return "", err
    }

    return resp.Content, nil
}

// getSortedProviders 获取按优先级排序的 Provider
func (m *MultiProviderClient) getSortedProviders() []ProviderConfig {
    m.mu.RLock()
    defer m.mu.RUnlock()

    cfgs := make([]ProviderConfig, len(m.providerConfig))
    copy(cfgs, m.providerConfig)

    // 按 Priority 排序
    sort.Slice(cfgs, func(i, j int) bool {
        // 优先级相同则随机（负载均衡）
        if cfgs[i].Priority == cfgs[j].Priority {
            return time.Now().UnixNano()%2 == 0
        }
        return cfgs[i].Priority < cfgs[j].Priority
    })

    return cfgs
}

// GetStats 获取各 Provider 状态
func (m *MultiProviderClient) GetStats() map[string]ProviderStats {
    m.mu.RLock()
    defer m.mu.RUnlock()

    stats := make(map[string]ProviderStats)
    for name, breaker := range m.breakers {
        stats[string(name)] = ProviderStats{
            State:        breaker.GetState().String(),
            BreakerStats: breaker.GetStats(),
        }
    }

    return stats
}

// ProviderStats Provider 状态
type ProviderStats struct {
    State        string              `json:"state"`
    BreakerStats circuit.BreakerStats `json:"breaker_stats"`
}

// SetProviderEnabled 动态启用/禁用 Provider
func (m *MultiProviderClient) SetProviderEnabled(provider Provider, enabled bool) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i, cfg := range m.providerConfig {
        if cfg.Name == provider {
            m.providerConfig[i].Enabled = enabled
            if enabled {
                // 重置熔断器
                if breaker, exists := m.breakers[provider]; exists {
                    breaker.Reset()
                }
            }
            break
        }
    }
}
```

---

## 3. 超时（分级超时）

**方案**：使用 `context.WithTimeout` 实现分级超时。

### 3.1 超时配置

```go
// internal/infrastructure/ai/timeout/timeout.go
package timeout

import (
    "context"
    "net"
    "net/http"
    "time"
)

// Config 超时配置
type Config struct {
    // HTTP 层超时
    DialTimeout      time.Duration // 建立连接
    TLSHandshake     time.Duration // TLS 握手
    ResponseHeader   time.Duration // 等待响应头
    IdleConnTimeout  time.Duration // 空闲连接超时

    // 应用层超时
    RequestTimeout  time.Duration // 完整请求超时
    StreamTimeout   time.Duration // 流式请求超时

    // LLM 特定
    GenerationTimeout time.Duration // 生成超时
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
    DialTimeout:        10 * time.Second,
    TLSHandshake:       5 * time.Second,
    ResponseHeader:     30 * time.Second,
    IdleConnTimeout:    90 * time.Second,
    RequestTimeout:     60 * time.Second,
    StreamTimeout:      120 * time.Second,
    GenerationTimeout:  55 * time.Second, // 略小于 RequestTimeout
}

// HTTPClientWithTimeout 创建带超时的 HTTP 客户端
func HTTPClientWithTimeout(cfg Config) *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            DialContext: (&net.Dialer{
                Timeout:   cfg.DialTimeout,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            TLSHandshakeTimeout:   cfg.TLSHandshake,
            ResponseHeaderTimeout: cfg.ResponseHeader,
            IdleConnTimeout:       cfg.IdleConnTimeout,
            MaxIdleConns:          100,
            MaxIdleConnsPerHost:   10,
            // 连接池配置
            ForceAttemptHTTP2:     true,
        },
        Timeout: cfg.RequestTimeout,
    }
}

// WithGenerationTimeout 创建生成超时上下文
func WithGenerationTimeout(parent context.Context, cfg Config) (context.Context, context.CancelFunc) {
    return context.WithTimeout(parent, cfg.GenerationTimeout)
}

// WithRequestTimeout 创建请求超时上下文
func WithRequestTimeout(parent context.Context, cfg Config) (context.Context, context.CancelFunc) {
    return context.WithTimeout(parent, cfg.RequestTimeout)
}
```

### 3.2 Provider 客户端实现（以 OpenAI 为例）

```go
// internal/infrastructure/ai/openai/client.go
package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "heka-backend/internal/infrastructure/ai"
    "heka-backend/internal/infrastructure/ai/timeout"
)

type Client struct {
    apiKey string
    model  string
    client *http.Client
    cfg    timeout.Config
}

func NewClient(apiKey, baseURL string) *Client {
    cfg := timeout.DefaultConfig

    // 可以根据 baseURL 调整配置
    if baseURL != "" {
        // 自定义端点（如 Ollama）
        cfg.GenerationTimeout = 120 * time.Second
        cfg.RequestTimeout = 130 * time.Second
    }

    return &Client{
        apiKey: apiKey,
        model:  "gpt-4o",
        client: timeout.HTTPClientWithTimeout(cfg),
        cfg:    cfg,
    }
}

func (c *Client) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
    // 创建生成超时
    genCtx, cancel := timeout.WithGenerationTimeout(ctx, c.cfg)
    defer cancel()

    // 构建请求体
    reqBody := map[string]interface{}{
        "model":       c.model,
        "messages":    req.Messages,
        "max_tokens":  req.MaxTokens,
        "temperature": req.Temperature,
        "stream":      req.Stream,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    // 创建 HTTP 请求
    url := "https://api.openai.com/v1/chat/completions"
    httpReq, err := http.NewRequestWithContext(genCtx, "POST", url, bytes.NewReader(jsonData))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    // 发送请求
    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    // 检查状态码
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, &HTTPError{
            StatusCode: resp.StatusCode,
            Body:       string(body),
        }
    }

    // 解析响应
    var result struct {
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
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    if len(result.Choices) == 0 {
        return nil, fmt.Errorf("no choices returned")
    }

    return &ai.ChatResponse{
        Content: result.Choices[0].Message.Content,
        Usage: ai.Usage{
            PromptTokens:     result.Usage.PromptTokens,
            CompletionTokens: result.Usage.CompletionTokens,
            TotalTokens:      result.Usage.TotalTokens,
        },
        Model: result.Model,
    }, nil
}

func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
    req := ai.ChatRequest{
        Messages: []ai.Message{
            {Role: "user", Content: prompt},
        },
        MaxTokens: 4096,
    }

    resp, err := c.Chat(ctx, req)
    if err != nil {
        return "", err
    }

    return resp.Content, nil
}

// HTTPError HTTP 错误
type HTTPError struct {
    StatusCode int
    Body       string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}
```

---

## 4. 重试（指数退避）

**方案**：指数退避 + 最大重试次数 + 可重试错误判断。

### 4.1 重试实现

```go
// internal/infrastructure/ai/retry/retry.go
package retry

import (
    "context"
    "errors"
    "fmt"
    "math"
    "net"
    "strings"
    "time"
)

// Config 重试配置
type Config struct {
    MaxAttempts int           // 最大尝试次数
    BaseDelay   time.Duration // 基础延迟
    MaxDelay    time.Duration // 最大延迟
    Multiplier  float64       // 延迟倍数
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
    MaxAttempts: 3,
    BaseDelay:   1 * time.Second,
    MaxDelay:    30 * time.Second,
    Multiplier:  2.0,
}

// IsRetryable 判断错误是否可重试
func IsRetryable(err error) bool {
    if err == nil {
        return false
    }

    // 网络错误
    var netErr net.Error
    if errors.As(err, &netErr) {
        if netErr.Timeout() || netErr.Temporary() {
            return true
        }
    }

    // HTTP 错误
    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        // 429 Too Many Requests
        // 500 Internal Server Error
        // 502 Bad Gateway
        // 503 Service Unavailable
        // 504 Gateway Timeout
        switch httpErr.StatusCode {
        case 429, 500, 502, 503, 504:
            return true
        }
    }

    // 检查错误消息
    errMsg := err.Error()
    retryableMessages := []string{
        "connection reset",
        "connection refused",
        "broken pipe",
        "timeout",
        "temporary failure",
        "rate limit",
        "service unavailable",
        "connection timed out",
        "deadline exceeded",
        "i/o timeout",
        "EOF",
    }

    lowerErrMsg := strings.ToLower(errMsg)
    for _, msg := range retryableMessages {
        if strings.Contains(lowerErrMsg, strings.ToLower(msg)) {
            return true
        }
    }

    return false
}

// Do 执行带重试的操作
func Do(ctx context.Context, cfg Config, fn func() error) error {
    var lastErr error

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        if attempt > 0 {
            // 计算延迟（指数退避 + 随机抖动）
            delay := calculateDelay(cfg, attempt)

            // 等待或取消
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return fmt.Errorf("retry canceled: %w", ctx.Err())
            }
        }

        // 执行操作
        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err

        // 检查是否可重试
        if !IsRetryable(err) {
            return fmt.Errorf("non-retryable error: %w", err)
        }

        // 最后一次尝试，不再重试
        if attempt == cfg.MaxAttempts-1 {
            break
        }
    }

    return fmt.Errorf("after %d attempts, last error: %w", cfg.MaxAttempts, lastErr)
}

// DoWithResult 执行带重试的操作（带返回值）
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
    var lastErr error
    var zero T

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        if attempt > 0 {
            delay := calculateDelay(cfg, attempt)

            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return zero, fmt.Errorf("retry canceled: %w", ctx.Err())
            }
        }

        result, err := fn()
        if err == nil {
            return result, nil
        }

        lastErr = err

        if !IsRetryable(err) {
            return zero, fmt.Errorf("non-retryable error: %w", err)
        }

        if attempt == cfg.MaxAttempts-1 {
            break
        }
    }

    return zero, fmt.Errorf("after %d attempts, last error: %w", cfg.MaxAttempts, lastErr)
}

// calculateDelay 计算重试延迟（指数退避 + 随机抖动）
func calculateDelay(cfg Config, attempt int) time.Duration {
    // 指数退避
    delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt-1)))

    // 限制最大延迟
    if delay > cfg.MaxDelay {
        delay = cfg.MaxDelay
    }

    // 添加随机抖动（±25%）
    jitter := time.Duration(float64(delay) * 0.25 * (2.0*time.Now().UnixNano()%1000/1000.0 - 0.5))

    return delay + jitter
}

// HTTPError HTTP 错误（需要和 openai 包中的 HTTPError 合并或统一）
type HTTPError struct {
    StatusCode int
    Body       string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// AsHTTPError 将 error 转换为 HTTPError
func AsHTTPError(err error) (*HTTPError, bool) {
    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        return httpErr, true
    }

    // 尝试从其他类型的错误中提取 HTTP 错误
    return nil, false
}
```

---

## 5. 整合：完整 LLM 调用层

### 5.1 LLM Manager

```go
// internal/infrastructure/ai/manager.go
package ai

import (
    "context"
    "fmt"
    "sync"
    "time"

    "heka-backend/internal/infrastructure/ai/circuit"
    "heka-backend/internal/infrastructure/ai/pool"
    "heka-backend/internal/infrastructure/ai/retry"
    "heka-backend/internal/infrastructure/ai/timeout"
)

// Manager LLM 调用管理器
type Manager struct {
    multiProvider *MultiProviderClient
    retryConfig   retry.Config
    timeoutConfig timeout.Config
    metrics       *Metrics
    mu            sync.RWMutex
}

// NewManager 创建 LLM 调用管理器
func NewManager(
    providerConfigs []ProviderConfig,
    poolCfg struct {
        MaxWorkers int
        QueueSize  int
    },
    retryCfg retry.Config,
    timeoutCfg timeout.Config,
) *Manager {
    // 创建 Pool
    pool := pool.NewPool(poolCfg.MaxWorkers, poolCfg.QueueSize)

    return &Manager{
        multiProvider: NewMultiProviderClient(providerConfigs, pool),
        retryConfig:   retryCfg,
        timeoutConfig: timeoutCfg,
        metrics:       NewMetrics(),
    }
}

// Chat 发起对话（完整容错）
func (m *Manager) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    start := time.Now()
    provider := "unknown"

    // 使用重试包装
    var resp *ChatResponse
    err := retry.Do(ctx, m.retryConfig, func() error {
        var innerErr error
        resp, innerErr = m.multiProvider.Chat(ctx, req)
        return innerErr
    })

    // 记录指标
    duration := time.Since(start)
    m.metrics.RecordRequest("chat", provider, err == nil, duration)

    if err != nil {
        return nil, fmt.Errorf("chat failed: %w", err)
    }

    return resp, nil
}

// GenerateText 生成文本（完整容错）
func (m *Manager) GenerateText(ctx context.Context, prompt string) (string, error) {
    req := ChatRequest{
        Messages: []Message{
            {Role: "user", Content: prompt},
        },
        MaxTokens: 4096,
    }

    resp, err := m.Chat(ctx, req)
    if err != nil {
        return "", err
    }

    return resp.Content, nil
}

// StreamChat 流式对话
func (m *Manager) StreamChat(ctx context.Context, req ChatRequest, callback func(chunk string)) error {
    start := time.Now()

    err := retry.Do(ctx, m.retryConfig, func() error {
        return m.multiProvider.StreamChat(ctx, req, callback)
    })

    duration := time.Since(start)
    m.metrics.RecordRequest("stream_chat", "unknown", err == nil, duration)

    return err
}

// GetMetrics 获取指标
func (m *Manager) GetMetrics() *MetricsSnapshot {
    return m.metrics.Snapshot()
}

// GetProviderStats 获取 Provider 状态
func (m *Manager) GetProviderStats() map[string]ProviderStats {
    return m.multiProvider.GetStats()
}

// GetPoolStats 获取 Pool 状态
func (m *Manager) GetPoolStats() pool.PoolStats {
    return m.multiProvider.GetPoolStats()
}

// Close 关闭管理器
func (m *Manager) Close() error {
    return m.multiProvider.Close()
}
```

### 5.2 Multi Provider 扩展（支持流式）

```go
// internal/infrastructure/ai/multi_provider.go (扩展)

// StreamChat 流式对话
func (m *MultiProviderClient) StreamChat(ctx context.Context, req ChatRequest, callback func(chunk string)) error {
    sortedProviders := m.getSortedProviders()

    var lastErr error
    attemptedProviders := []Provider{}

    for _, cfg := range sortedProviders {
        provider := cfg.Name
        attemptedProviders = append(attemptedProviders, provider)

        client, exists := m.providers[provider]
        if !exists {
            continue
        }

        breaker := m.breakers[provider]

        if !breaker.Allow() {
            lastErr = fmt.Errorf("provider %s is circuit open", provider)
            continue
        }

        // 检查客户端是否支持流式
        streamClient, ok := client.(StreamingLLMClient)
        if !ok {
            lastErr = fmt.Errorf("provider %s does not support streaming", provider)
            continue
        }

        err := m.pool.SubmitAndWait(ctx, func(poolCtx context.Context) error {
            timeoutCtx, cancel := context.WithTimeout(poolCtx, 120*time.Second)
            defer cancel()

            return streamClient.StreamChat(timeoutCtx, req, callback)
        })

        if err != nil {
            breaker.RecordFailure()
            lastErr = fmt.Errorf("provider %s failed: %w", provider, err)
            continue
        }

        breaker.RecordSuccess()
        return nil
    }

    return fmt.Errorf("all providers failed (attempted: %v), last error: %w", attemptedProviders, lastErr)
}

// GetPoolStats 获取 Pool 状态
func (m *MultiProviderClient) GetPoolStats() pool.PoolStats {
    return m.pool.Stats()
}

// Close 关闭客户端
func (m *MultiProviderClient) Close() error {
    return m.pool.Close()
}

// StreamingLLMClient 支持流式的 LLM 客户端
type StreamingLLMClient interface {
    LLMClient
    StreamChat(ctx context.Context, req ChatRequest, callback func(chunk string)) error
}
```

---

## 6. 监控指标

```go
// internal/infrastructure/ai/metrics/metrics.go
package metrics

import (
    "sync"
    "time"
)

type Metrics struct {
    mu               sync.RWMutex
    requestsTotal    int64
    requestsSuccess  int64
    requestsFailed   int64
    totalDuration    time.Duration
    providerStats    map[string]*ProviderMetrics
    operationStats   map[string]*OperationMetrics
}

type ProviderMetrics struct {
    Requests      int64
    Successes     int64
    Failures      int64
    AvgDuration   time.Duration
    LastFailure   time.Time
    LastSuccess   time.Time
    ConsecutiveFailures int
}

type OperationMetrics struct {
    Requests     int64
    Successes    int64
    Failures     int64
    AvgDuration  time.Duration
    P95Duration  time.Duration
    P99Duration  time.Duration
    durations    []time.Duration
}

type MetricsSnapshot struct {
    RequestsTotal   int64                       `json:"requests_total"`
    RequestsSuccess int64                       `json:"requests_success"`
    RequestsFailed  int64                       `json:"requests_failed"`
    SuccessRate     float64                     `json:"success_rate"`
    AvgDuration     time.Duration               `json:"avg_duration"`
    ProviderStats   map[string]ProviderSnapshot `json:"provider_stats"`
    OperationStats  map[string]OperationSnapshot `json:"operation_stats"`
}

type ProviderSnapshot struct {
    Requests            int64     `json:"requests"`
    Successes           int64     `json:"successes"`
    Failures            int64     `json:"failures"`
    SuccessRate         float64   `json:"success_rate"`
    AvgDuration         time.Duration `json:"avg_duration"`
    LastFailure         time.Time `json:"last_failure"`
    LastSuccess         time.Time `json:"last_success"`
    ConsecutiveFailures int       `json:"consecutive_failures"`
}

type OperationSnapshot struct {
    Requests    int64         `json:"requests"`
    Successes   int64         `json:"successes"`
    Failures    int64         `json:"failures"`
    SuccessRate float64       `json:"success_rate"`
    AvgDuration time.Duration `json:"avg_duration"`
    P95Duration time.Duration `json:"p95_duration"`
    P99Duration time.Duration `json:"p99_duration"`
}

func NewMetrics() *Metrics {
    return &Metrics{
        providerStats:  make(map[string]*ProviderMetrics),
        operationStats: make(map[string]*OperationMetrics),
    }
}

func (m *Metrics) RecordRequest(operation string, provider string, success bool, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.requestsTotal++
    m.totalDuration += duration

    if success {
        m.requestsSuccess++
    } else {
        m.requestsFailed++
    }

    // Provider 统计
    stats, exists := m.providerStats[provider]
    if !exists {
        stats = &ProviderMetrics{}
        m.providerStats[provider] = stats
    }

    stats.Requests++
    if success {
        stats.Successes++
        stats.LastSuccess = time.Now()
        stats.ConsecutiveFailures = 0
    } else {
        stats.Failures++
        stats.LastFailure = time.Now()
        stats.ConsecutiveFailures++
    }

    // Operation 统计
    opStats, exists := m.operationStats[operation]
    if !exists {
        opStats = &OperationMetrics{
            durations: make([]time.Duration, 0, 1000),
        }
        m.operationStats[operation] = opStats
    }

    opStats.Requests++
    if success {
        opStats.Successes++
    } else {
        opStats.Failures++
    }

    // 只保留最近 1000 个时长用于百分位计算
    if len(opStats.durations) >= 1000 {
        opStats.durations = opStats.durations[1:]
    }
    opStats.durations = append(opStats.durations, duration)
}

func (m *Metrics) Snapshot() *MetricsSnapshot {
    m.mu.RLock()
    defer m.mu.RUnlock()

    snapshot := &MetricsSnapshot{
        RequestsTotal:   m.requestsTotal,
        RequestsSuccess: m.requestsSuccess,
        RequestsFailed:  m.requestsFailed,
        AvgDuration:     m.totalDuration / time.Duration(m.requestsTotal),
        ProviderStats:   make(map[string]ProviderSnapshot),
        OperationStats:  make(map[string]OperationSnapshot),
    }

    if m.requestsTotal > 0 {
        snapshot.SuccessRate = float64(m.requestsSuccess) / float64(m.requestsTotal)
    }

    for name, stats := range m.providerStats {
        var avgDuration time.Duration
        if stats.Requests > 0 {
            avgDuration = time.Duration(int64(stats.AvgDuration) / stats.Requests)
        }

        var successRate float64
        if stats.Requests > 0 {
            successRate = float64(stats.Successes) / float64(stats.Requests)
        }

        snapshot.ProviderStats[name] = ProviderSnapshot{
            Requests:            stats.Requests,
            Successes:           stats.Successes,
            Failures:            stats.Failures,
            SuccessRate:         successRate,
            AvgDuration:         avgDuration,
            LastFailure:         stats.LastFailure,
            LastSuccess:         stats.LastSuccess,
            ConsecutiveFailures: stats.ConsecutiveFailures,
        }
    }

    for name, stats := range m.operationStats {
        var avgDuration time.Duration
        var p95Duration, p99Duration time.Duration

        if len(stats.durations) > 0 {
            // 计算百分位
            sorted := make([]time.Duration, len(stats.durations))
            copy(sorted, stats.durations)
            // 简单排序（实际应用可以用更高效的算法）
            for i := 0; i < len(sorted); i++ {
                for j := i + 1; j < len(sorted); j++ {
                    if sorted[i] > sorted[j] {
                        sorted[i], sorted[j] = sorted[j], sorted[i]
                    }
                }
            }

            avgDuration = sorted[len(sorted)/2]
            p95Index := len(sorted) * 95 / 100
            p99Index := len(sorted) * 99 / 100
            if p95Index < len(sorted) {
                p95Duration = sorted[p95Index]
            }
            if p99Index < len(sorted) {
                p99Duration = sorted[p99Index]
            }
        }

        var successRate float64
        if stats.Requests > 0 {
            successRate = float64(stats.Successes) / float64(stats.Requests)
        }

        snapshot.OperationStats[name] = OperationSnapshot{
            Requests:    stats.Requests,
            Successes:   stats.Successes,
            Failures:    stats.Failures,
            SuccessRate: successRate,
            AvgDuration: avgDuration,
            P95Duration: p95Duration,
            P99Duration: p99Duration,
        }
    }

    return snapshot
}

// Reset 重置指标
func (m *Metrics) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.requestsTotal = 0
    m.requestsSuccess = 0
    m.requestsFailed = 0
    m.totalDuration = 0
    m.providerStats = make(map[string]*ProviderMetrics)
    m.operationStats = make(map[string]*OperationMetrics)
}
```

---

## 7. 配置文件示例

```yaml
# config/ai.yaml
ai:
  # Worker Pool 配置
  pool_max_workers: 10
  pool_queue_size: 100

  # 重试配置
  retry_max_attempts: 3
  retry_base_delay: 1s
  retry_max_delay: 30s
  retry_multiplier: 2.0

  # 超时配置
  timeout_dial: 10s
  timeout_tls_handshake: 5s
  timeout_response_header: 30s
  timeout_request: 60s
  timeout_generation: 55s

  # Provider 配置
  providers:
    - name: claude
      api_key: ${CLAUDE_API_KEY}
      base_url: https://api.anthropic.com
      model: claude-3-5-sonnet-20241022
      priority: 1
      enabled: true

    - name: openai
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com
      model: gpt-4o
      priority: 2
      enabled: true

    - name: gemini
      api_key: ${GEMINI_API_KEY}
      base_url: https://generativelanguage.googleapis.com
      model: gemini-1.5-pro
      priority: 3
      enabled: true

    - name: ollama
      api_key: ""
      base_url: http://localhost:11434
      model: llama3.1
      priority: 4
      enabled: false
```

---

## 8. 使用示例

### 8.1 初始化

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "os"

    "heka-backend/internal/infrastructure/ai"
    "heka-backend/internal/infrastructure/ai/pool"
    "heka-backend/internal/infrastructure/ai/retry"
    "heka-backend/internal/infrastructure/ai/timeout"
    "heka-backend/internal/shared/config"
)

func main() {
    // 加载配置
    cfg := config.Load()

    // 创建 LLM Manager
    llmManager := ai.NewManager(
        cfg.AI.Providers,
        struct {
            MaxWorkers int
            QueueSize  int
        }{
            MaxWorkers: cfg.AI.PoolMaxWorkers,
            QueueSize:  cfg.AI.PoolQueueSize,
        },
        retry.Config{
            MaxAttempts: cfg.AI.RetryMaxAttempts,
            BaseDelay:   cfg.AI.RetryBaseDelay,
            MaxDelay:    cfg.AI.RetryMaxDelay,
            Multiplier:  cfg.AI.RetryMultiplier,
        },
        timeout.Config{
            DialTimeout:        cfg.AI.TimeoutDial,
            TLSHandshake:       cfg.AI.TimeoutTLSHandshake,
            ResponseHeader:     cfg.AI.TimeoutResponseHeader,
            RequestTimeout:     cfg.AI.TimeoutRequest,
            GenerationTimeout:  cfg.AI.TimeoutGeneration,
        },
    )
    defer llmManager.Close()

    // 在应用服务中使用
    aiSvc := aiservice.NewService(llmManager)

    // 启动 HTTP 服务
    // ...
}
```

### 8.2 在应用服务中使用

```go
// internal/application/testcase/service.go
package testcase

import (
    "context"
    "fmt"

    "heka-backend/internal/infrastructure/ai"
)

type Service struct {
    tcRepo  testcase.Repository
    fileRepo file.Repository
    llm     *ai.Manager
}

func NewService(
    tcRepo testcase.Repository,
    fileRepo file.Repository,
    llm *ai.Manager,
) *Service {
    return &Service{
        tcRepo:  tcRepo,
        fileRepo: fileRepo,
        llm:     llm,
    }
}

// GenerateByAI AI 生成测试用例
func (s *Service) GenerateByAI(ctx context.Context, req dto.GenerateByAIRequest) (*dto.GenerateByAIResponse, error) {
    // 1. 获取文件内容
    f, err := s.fileRepo.FindByID(ctx, req.FileID)
    if err != nil {
        return nil, fmt.Errorf("get file: %w", err)
    }

    // 2. 构建 prompt
    prompt := s.buildPrompt(f, req.Query)

    // 3. 调用 LLM（自动容错、重试、故障转移）
    result, err := s.llm.GenerateText(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("LLM generation failed: %w", err)
    }

    // 4. 解析结果
    testCases, err := s.parseTestCases(result)
    if err != nil {
        return nil, fmt.Errorf("parse test cases: %w", err)
    }

    // 5. 批量创建用例
    var createdIDs []shared.ID
    for _, tc := range testCases {
        created, err := s.CreateTestCase(ctx, tc)
        if err != nil {
            return nil, fmt.Errorf("create test case: %w", err)
        }
        createdIDs = append(createdIDs, created.ID)
    }

    return &dto.GenerateByAIResponse{
        TestCaseIDs: createdIDs,
        Count:       len(createdIDs),
    }, nil
}

func (s *Service) buildPrompt(f *file.File, query string) string {
    return fmt.Sprintf(`
你是一个专业的测试工程师。根据以下需求文档生成测试用例。

文档类型: %s
文档内容:
%s

用户问题: %s

请生成 JSON 格式的测试用例，包含以下字段：
- title: 测试用例标题
- description: 测试用例描述
- steps: 测试步骤数组，每个步骤包含 action 和 expected
- priority: 优先级（0=低, 1=中, 2=高, 3=紧急）
- tags: 标签数组

返回格式示例:
[
  {
    "title": "用户登录成功",
    "description": "验证用户使用正确的用户名和密码可以成功登录",
    "steps": [
      {"action": "打开登录页面", "expected": "显示登录表单"},
      {"action": "输入正确的用户名和密码", "expected": "输入框显示输入内容"},
      {"action": "点击登录按钮", "expected": "登录成功，跳转到首页"}
    ],
    "priority": 1,
    "tags": ["登录", "功能测试"]
  }
]
`, f.Type, f.Content, query)
}
```

### 8.3 监控端点

```go
// internal/interface/http/handler/monitoring.go
package handler

import (
    "net/http"

    "heka-backend/internal/infrastructure/ai"
    "heka-backend/internal/interface/http/response"
)

type MonitoringHandler struct {
    llm *ai.Manager
}

func NewMonitoringHandler(llm *ai.Manager) *MonitoringHandler {
    return &MonitoringHandler{llm: llm}
}

// GetLLMMetrics 获取 LLM 调用指标
// GET /api/monitoring/llm/metrics
func (h *MonitoringHandler) GetLLMMetrics(w http.ResponseWriter, r *http.Request) {
    metrics := h.llm.GetMetrics()
    response.JSON(w, http.StatusOK, metrics)
}

// GetLLMProviderStats 获取 Provider 状态
// GET /api/monitoring/llm/providers
func (h *MonitoringHandler) GetLLMProviderStats(w http.ResponseWriter, r *http.Request) {
    stats := h.llm.GetProviderStats()
    response.JSON(w, http.StatusOK, stats)
}

// GetLLMPoolStats 获取 Pool 状态
// GET /api/monitoring/llm/pool
func (h *MonitoringHandler) GetLLMPoolStats(w http.ResponseWriter, r *http.Request) {
    stats := h.llm.GetPoolStats()
    response.JSON(w, http.StatusOK, stats)
}
```

---

## 9. 方案总结

| 维度 | 方案 | 关键点 |
|------|------|--------|
| **线程管理** | Worker Pool | 控制并发数，防止资源耗尽，默认 10 个 worker |
| **容错** | 熔断器 + 故障转移 | Provider 故障自动切换，熔断防止雪崩 |
| **超时** | 分级超时 | 连接、TLS、响应头、请求、生成五级超时 |
| **重试** | 指数退避 + 随机抖动 | 最多 3 次，延迟 1s → 2s → 4s → 8s，最大 30s |

**关键特性**：
- ✅ 多 Provider 支持（OpenAI、Claude、Gemini、Ollama）
- ✅ 自动故障转移（按优先级）
- ✅ 熔断器保护（连续失败自动熔断）
- ✅ 智能重试（只重试可重试错误）
- ✅ 分级超时（防止长时间阻塞）
- ✅ 线程池管理（控制并发）
- ✅ 完整监控（成功率、延迟、Provider 状态）

---

## 5. AI 降级策略

### 5.1 全 Provider 故障降级

当所有 AI Provider 都不可用时，采用 graceful degradation：

```go
// internal/infrastructure/ai/fallback.go

// FallbackStrategy 降级策略
type FallbackStrategy int

const (
    FallbackCache       FallbackStrategy = iota // 返回缓存的相似结果
    FallbackTemplate                             // 返回预定义模板
    FallbackReject                               // 直接拒绝，提示用户稍后重试
)

// DegradationManager 降级管理器
type DegradationManager struct {
    cache       *ResponseCache
    strategy    FallbackStrategy
    logger      Logger
}

func (d *DegradationManager) HandleAllProvidersDown(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
    switch d.strategy {
    case FallbackCache:
        // 尝试返回缓存的相似请求结果
        if cached, ok := d.cache.FindSimilar(ctx, req.Prompt); ok {
            d.logger.Warn("returning cached AI response due to provider outage")
            return cached, nil
        }
        // 缓存也没有，降级到模板
        return d.returnTemplate(req)
        
    case FallbackTemplate:
        return d.returnTemplate(req)
        
    default:
        return nil, fmt.Errorf("AI-IE-001: all AI providers unavailable, please try again later")
    }
}

// returnTemplate 返回预定义的用例模板
func (d *DegradationManager) returnTemplate(req *GenerateRequest) (*GenerateResponse, error) {
    d.logger.Warn("returning template AI response due to provider outage")
    
    return &GenerateResponse{
        TestCases: []TestCase{
            {
                Title:       "基础功能验证",
                Description: "验证核心功能正常工作",
                Steps: []Step{
                    {Action: "执行基本操作", Expected: "操作成功，无异常"},
                    {Action: "验证返回结果", Expected: "返回结果符合预期"},
                },
                Priority: 1,
                Tags:     []string{"基础验证"},
            },
        },
        IsTemplate: true,
        Message:    "AI 服务暂时不可用，已返回默认模板，请手动调整",
    }, nil
}
```

### 5.2 降级触发条件

| 条件 | 动作 | 恢复条件 |
|------|------|----------|
| 单个 Provider 连续 5 次失败 | 熔断该 Provider | 30s 后半开尝试 |
| 所有 Provider 熔断 | 启用降级策略 | 任意 Provider 恢复 |
| AI 调用排队超过 50 | 拒绝新请求（AI-RT-001） | 队列降到 30 以下 |
| 单次调用超时 60s | 记录超时，尝试下一个 Provider | - |

---

## 6. 成本控制

### 6.1 Token 用量追踪

```go
// internal/infrastructure/ai/cost/tracker.go
package cost

type TokenUsage struct {
    Provider   string
    Model      string
    PromptTokens   int
    OutputTokens   int
    TotalTokens    int
    EstimatedCost  float64 // USD
    Timestamp  time.Time
    ProjectID  shared.ID
}

type Tracker struct {
    store TokenStore // Redis 或 PostgreSQL
}

func (t *Tracker) Record(ctx context.Context, usage *TokenUsage) error {
    // 记录到存储
    return t.store.Save(ctx, usage)
}

func (t *Tracker) GetDailyUsage(ctx context.Context, projectID shared.ID) (*DailySummary, error) {
    // 查询当日用量
    return t.store.GetDailySummary(ctx, projectID, time.Now())
}
```

### 6.2 预算限制

```go
// internal/infrastructure/ai/cost/budget.go
package cost

type BudgetLimiter struct {
    dailyBudget    float64 // 每日预算上限（USD）
    monthlyBudget  float64 // 每月预算上限（USD）
    store          BudgetStore
}

func (l *BudgetLimiter) Allow(ctx context.Context, projectID shared.ID) error {
    daily, _ := l.store.GetDailySpend(ctx, projectID)
    if daily >= l.dailyBudget {
        return fmt.Errorf("AI-RT-001: daily AI budget exceeded ($%.2f/$%.2f)", daily, l.dailyBudget)
    }
    
    monthly, _ := l.store.GetMonthlySpend(ctx, projectID)
    if monthly >= l.monthlyBudget {
        return fmt.Errorf("AI-RT-001: monthly AI budget exceeded ($%.2f/$%.2f)", monthly, l.monthlyBudget)
    }
    
    return nil
}
```

**默认配置**：
```yaml
ai:
  budget:
    daily_limit: 10.0    # USD
    monthly_limit: 200.0  # USD
    warn_threshold: 0.8   # 80% 时告警
```

### 6.3 成本监控 API

```
GET /api/v1/monitoring/ai/cost?period=daily
```

```json
{
  "code": 0,
  "data": {
    "period": "2025-05-15",
    "total_cost": 3.45,
    "budget_limit": 10.0,
    "usage_percent": 34.5,
    "by_provider": {
      "claude": {"tokens": 50000, "cost": 2.50},
      "openai": {"tokens": 20000, "cost": 0.95}
    },
    "by_project": [
      {"project_id": "xxx", "project_name": "Heka", "tokens": 70000, "cost": 3.45}
    ]
  }
}
```

---

## 7. AI 响应缓存

### 7.1 缓存策略

对**确定性 Prompt**（相同输入 → 相同输出）启用缓存：

```go
// internal/infrastructure/ai/cache.go
package ai

type ResponseCache struct {
    redis *redis.Client
    ttl   time.Duration
}

func NewResponseCache(redis *redis.Client) *ResponseCache {
    return &ResponseCache{
        redis: redis,
        ttl:   1 * time.Hour,
    }
}

// Get 获取缓存的 AI 响应
func (c *ResponseCache) Get(ctx context.Context, prompt string) (*GenerateResponse, bool) {
    key := c.cacheKey(prompt)
    data, err := c.redis.Get(ctx, key).Bytes()
    if err != nil {
        return nil, false
    }
    
    var resp GenerateResponse
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, false
    }
    return &resp, true
}

// Set 缓存 AI 响应
func (c *ResponseCache) Set(ctx context.Context, prompt string, resp *GenerateResponse) error {
    key := c.cacheKey(prompt)
    data, _ := json.Marshal(resp)
    return c.redis.Set(ctx, key, data, c.ttl).Err()
}

func (c *ResponseCache) cacheKey(prompt string) string {
    hash := sha256.Sum256([]byte(prompt))
    return fmt.Sprintf("ai:cache:%x", hash)
}

// FindSimilar 查找相似 Prompt 的缓存结果（降级时使用）
func (c *ResponseCache) FindSimilar(ctx context.Context, prompt string) (*GenerateResponse, bool) {
    // 简化实现：精确匹配
    // 生产环境可使用向量相似度匹配
    return c.Get(ctx, prompt)
}
```

### 7.2 缓存使用规则

| 场景 | 是否缓存 | TTL | 说明 |
|------|---------|-----|------|
| AI 用例生成（基于文档） | ✅ | 1h | 同一文档+查询生成结果一致 |
| AI 智能分析（基于代码变更） | ❌ | - | 每次变更不同，不应缓存 |
| Embedding 生成 | ✅ | 24h | 相同文本生成相同向量 |
| RAG 查询结果 | ✅ | 10min | 热数据可适当延长 |

---

## 8. Streaming 实现

### 8.1 SSE 流式响应

```go
// internal/infrastructure/ai/streaming.go
package ai

type StreamEvent struct {
    Type    string `json:"type"`    // "token", "done", "error"
    Content string `json:"content"`
}

// StreamGenerate 流式生成
func (c *ClaudeClient) StreamGenerate(ctx context.Context, prompt string) (<-chan StreamEvent, error) {
    ch := make(chan StreamEvent, 100)
    
    go func() {
        defer close(ch)
        
        reqBody := ClaudeStreamRequest{
            Model:     c.model,
            MaxTokens: 4096,
            Stream:    true,
            Messages:  []Message{{Role: "user", Content: prompt}},
        }
        
        body, _ := json.Marshal(reqBody)
        req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
        req.Header.Set("x-api-key", c.apiKey)
        req.Header.Set("anthropic-version", "2023-06-01")
        req.Header.Set("content-type", "application/json")
        
        resp, err := c.client.Do(req)
        if err != nil {
            ch <- StreamEvent{Type: "error", Content: err.Error()}
            return
        }
        defer resp.Body.Close()
        
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            if !strings.HasPrefix(line, "data: ") {
                continue
            }
            
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                ch <- StreamEvent{Type: "done"}
                return
            }
            
            var event map[string]interface{}
            json.Unmarshal([]byte(data), &event)
            
            if delta, ok := event["delta"].(map[string]interface{}); ok {
                if text, ok := delta["text"].(string); ok {
                    ch <- StreamEvent{Type: "token", Content: text}
                }
            }
        }
        
        ch <- StreamEvent{Type: "done"}
    }()
    
    return ch, nil
}
```

### 8.2 Handler 层 SSE 处理

```go
// SSE Handler
func (h *AIHandler) StreamGenerate(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        response.Error(w, http.StatusInternalServerError, "streaming not supported")
        return
    }
    
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    stream, err := h.service.StreamGenerate(r.Context(), req)
    if err != nil {
        fmt.Fprintf(w, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
        flusher.Flush()
        return
    }
    
    for event := range stream {
        data, _ := json.Marshal(event)
        fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
        flusher.Flush()
    }
}
```
