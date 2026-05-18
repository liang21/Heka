# Heka 性能瓶颈分析

> 基于当前部署架构的性能瓶颈识别和优先级排序
> 版本：v1.0
> 日期：2025-05-15

---

## 瓶颈排序总览

| 排名 | 瓶颈名称 | 严重程度 | 触发条件 | 一期处理 | 影响范围 |
|------|---------|---------|----------|----------|----------|
| 1 | AI 调用瓶颈 | 🔴 严重 | 多用户并发AI生成 | ✅ 必须 | 核心功能 |
| 2 | 文件解析瓶颈 | 🔴 严重 | 大文件上传、批量上传 | ✅ 必须 | 核心功能 |
| 3 | Milvus 检索瓶颈 | 🟠 中等 | 10万+ 文档块、并发检索 | ✅ 建议 | RAG 功能 |
| 4 | PostgreSQL 查询瓶颈 | 🟠 中等 | 100万+ 执行记录、复杂报表 | ⚠️ 可选 | 报表功能 |
| 5 | 内存资源瓶颈 | 🟡 轻度 | 所有服务同时运行、缓存增长 | ⚠️ 可选 | 整体稳定性 |
| 6 | 网络带宽瓶颈 | 🟡 轻度 | 多用户同时上传文件 | ❌ 不处理 | 文件上传 |

---

## 1. AI 调用瓶颈（最严重）

### 问题描述

**严重程度**：🔴 严重

**核心问题**：
- AI 调用响应时间长（10-60s）
- 外部 API 不稳定（5-20% 失败率）
- 多用户并发时 Worker Pool 容易满载
- 成本高昂（按 Token 计费）

### 触发条件

```
场景 1：团队会议后集体使用 AI 生成
- 5 个用户同时点击"AI 生成用例"
- 每个请求需要 30-60s
- Worker Pool（10 个 worker）满载
- 队列积压，后续请求等待

场景 2：大文档生成
- 单个 50 页需求文档
- 生成 100+ 测试用例
- 需要多次 AI 调用（分块生成）
- 总耗时 2-5 分钟

场景 3：AI 智能分析高频使用
- 代码变更频繁（每天多次）
- 每次分析需要调用 LLM
- 累积调用次数快速增长
```

### 性能指标

```yaml
当前配置:
  pool_max_workers: 10        # 最多 10 个并发
  pool_queue_size: 100        # 队列长度 100
  timeout_request: 60s        # 请求超时 60s

瓶颈阈值:
  并发用户: > 5              # 超过 5 人同时使用
  队列深度: > 20             # 队列积压超过 20
  响应时间: > 30s            # 用户感知卡顿
  失败率: > 5%               # 用户体验明显下降
```

### 一期处理方案（必须处理）

**方案 1：任务队列 + 轮询模式**

```go
// ===== 当前方案（同步等待） =====
// 用户点击 AI 生成 → 阻塞等待 30-60s → 返回结果

// ===== 改进方案（异步任务） =====
// 1. 用户点击 AI 生成 → 立即返回 task_id
// 2. 后台 Worker Pool 处理
// 3. 前端轮询任务状态

// 伪代码
func (h *Handler) GenerateByAI(w http.ResponseWriter, r *http.Request) {
    // 1. 创建任务
    task := &AITask{
        ID:     uuid.New().String(),
        Status: "pending",
        UserID: getUserID(r),
    }
    h.taskRepo.Create(r.Context(), task)
    
    // 2. 提交到队列（立即返回）
    h.aiQueue.Publish(task)
    
    // 3. 立即返回 task_id
    response.JSON(w, http.StatusAccepted, map[string]string{
        "task_id": task.ID,
        "status":  "pending",
        "message": "AI 生成任务已创建，请稍后查看结果",
    })
}

// 4. 前端轮询
// GET /api/ai/tasks/{task_id}
// 返回: { status: "processing|completed|failed", result: [...] }
```

**方案 2：WebSocket 实时推送（可选）**

```go
// ===== WebSocket 推送方案 =====
// 避免轮询，服务端主动推送进度

func (h *Handler) GenerateByAIWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. 升级到 WebSocket
    conn, _ := upgrader.Upgrade(w, r, nil)
    
    // 2. 创建任务
    task := createTask()
    
    // 3. 后台处理，实时推送进度
    go func() {
        updateProgress(conn, task.ID, 0, "开始分析文档...")
        
        chunks := retrieveChunks(task.FileID)
        updateProgress(conn, task.ID, 30, "检索到相关内容")
        
        testCases := generateTestCases(chunks)
        updateProgress(conn, task.ID, 80, "生成测试用例")
        
        saveResults(task.ID, testCases)
        updateProgress(conn, task.ID, 100, "完成")
    }()
}
```

**方案 3：流式响应（适用于单个生成）**

```go
// ===== 流式响应方案 =====
// 边生成边返回，用户感知更快

func (h *Handler) GenerateByAIStream(w http.ResponseWriter, r *http.Request) {
    // 1. 设置流式响应
    flusher, _ := w.(http.Flusher)
    w.Header().Set("Content-Type", "text/event-stream")
    
    // 2. 调用 AI（流式）
    for chunk := range aiClient.StreamGenerate(ctx, prompt) {
        // 3. 逐块返回
        fmt.Fprintf(w, "data: %s\n\n", chunk)
        flusher.Flush()
    }
}
```

**方案 4：降级策略**

```go
// ===== 降级策略 =====
// 高峰期自动降级，保证核心功能可用

func (s *Service) GenerateTestCases(ctx context.Context, req *GenerateRequest) error {
    // 1. 检查队列长度
    queueLen := s.aiPool.QueueLength()
    
    if queueLen > 50 {
        // 队列积压，触发降级
        if req.Priority == "low" {
            // 低优先级任务拒绝
            return ErrTooBusy
        }
        
        // 高优先级任务继续，但使用更快的模型
        req.Model = "gpt-3.5-turbo"  // 替代 gpt-4
        req.MaxTokens = 2048         // 减少生成数量
    }
    
    // 2. 执行生成
    return s.aiClient.Generate(ctx, req)
}
```

### 监控指标

```go
// ===== 需要监控的指标 =====

// 1. Worker Pool 状态
type PoolMetrics struct {
    ActiveWorkers   int           `json:"active_workers"`    // 活跃 Worker
    QueueLength     int           `json:"queue_length"`      // 队列长度
    AvgWaitTime     time.Duration `json:"avg_wait_time"`      // 平均等待时间
    AvgProcessTime  time.Duration `json:"avg_process_time"`   // 平均处理时间
}

// 2. AI Provider 状态
type ProviderMetrics struct {
    Provider       string    `json:"provider"`
    SuccessRate    float64   `json:"success_rate"`     // 成功率
    AvgLatency     time.Duration `json:"avg_latency"` // 平均延迟
    CircuitState   string    `json:"circuit_state"`    // 熔断状态
    TotalCost      float64   `json:"total_cost"`       // 累计成本（美元）
}

// 3. 任务队列状态
type QueueMetrics struct {
    Pending        int       `json:"pending"`         // 待处理
    Processing     int       `json:"processing"`      // 处理中
    Completed      int       `json:"completed"`       // 今日完成
    Failed         int       `json:"failed"`          // 今日失败
    AvgQueueTime   time.Duration `json:"avg_queue_time"` // 平均排队时间
}
```

### 一期实施优先级

| 优先级 | 方案 | 实施难度 | 收益 | 是否一期 |
|--------|------|---------|------|----------|
| P0 | 异步任务 + 轮询 | 中 | 高 | ✅ 必须 |
| P1 | WebSocket 推送 | 高 | 中 | ⚠️ 可选 |
| P2 | 流式响应 | 低 | 中 | ⚠️ 可选 |
| P3 | 降级策略 | 中 | 高 | ⚠️ 可选 |
| P4 | 本地模型兜底 | 高 | 低 | ❌ 不处理 |

**一期最小方案**：异步任务 + 轮询模式

---

## 2. 文件解析瓶颈（严重）

### 问题描述

**严重程度**：🔴 严重

**核心问题**：
- 大文件解析耗时（PDF/Word 解析慢）
- OCR 图片识别非常慢（单张 5-30s）
- 批量上传时解析任务积压
- 占用大量 CPU 和内存资源

### 触发条件

```
场景 1：大 PDF 上传
- 100 页 PDF 文档
- 文本提取：10-30s
- 如果需要 OCR：2-5 分钟

场景 2：批量上传
- 一次上传 20 个 Word 文档
- 顺序解析：20-60 秒
- 阻塞其他任务

场景 3：图片 OCR
- 50 张 UI 截图
- 每张 5-10s
- 总耗时：4-8 分钟
```

### 性能分析

```yaml
文件类型 | 解析耗时 | OCR 耗时 | CPU 占用 | 内存占用
---------|---------|---------|----------|----------
PDF      | 10-30s  | 2-5min  | 80-100%  | 500MB-2GB
Word     | 5-15s   | 1-3min  | 50-80%   | 200MB-1GB
Excel    | 3-10s   | N/A     | 30-50%   | 100MB-500MB
Image    | 0.1s    | 5-30s   | 60-80%   | 300MB-1GB
Figma    | 2-5s    | N/A     | 10-20%   | 50MB-200MB

瓶颈阈值:
  单文件大小: > 10MB        # 解析时间明显增加
  批量数量: > 5 个          # 开始积压
  OCR 页数: > 10 页        # 耗时过长
```

### 一期处理方案（必须处理）

**方案 1：异步解析 + 进度反馈**

```go
// ===== 文件上传流程改进 =====

// 1. 上传接口（立即返回）
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
    // 1. 保存文件
    file := saveUploadedFile(r)
    
    // 2. 创建文件记录（状态：pending）
    fileRecord := &File{
        ID:         uuid.New().String(),
        Name:       file.Name,
        Size:       file.Size,
        Path:       file.Path,
        Status:     "pending",
        IndexStatus: "pending",
    }
    h.fileRepo.Create(r.Context(), fileRecord)
    
    // 3. 提交解析任务到队列
    h.parseQueue.Publish(&ParseTask{
        FileID: fileRecord.ID,
        Type:   file.Type,
    })
    
    // 4. 立即返回
    response.JSON(w, http.StatusCreated, map[string]string{
        "file_id": fileRecord.ID,
        "status":  "pending",
        "message": "文件上传成功，正在解析中...",
    })
}

// 2. 解析 Worker
func (w *ParseWorker) Process(task *ParseTask) error {
    // 1. 更新状态：processing
    w.fileRepo.UpdateStatus(task.FileID, "processing", 0, nil)
    
    // 2. 读取文件
    file := w.fileRepo.FindByID(task.FileID)
    
    // 3. 根据类型解析
    var progress int
    switch task.Type {
    case "pdf":
        progress = w.parsePDF(file)
    case "docx":
        progress = w.parseWord(file)
    case "image":
        progress = w.parseImage(file)  // OCR
    }
    
    // 4. 更新状态：completed
    w.fileRepo.UpdateStatus(task.FileID, "completed", 100, nil)
    
    // 5. 触发后续任务（RAG 索引）
    w.ragQueue.Publish(&RAGTask{FileID: task.FileID})
    
    return nil
}

// 3. 进度查询接口
// GET /api/files/{file_id}/parse-status
func (h *Handler) GetParseStatus(w http.ResponseWriter, r *http.Request) {
    file := h.fileRepo.FindByID(getID(r))
    
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "file_id":       file.ID,
        "status":        file.IndexStatus,
        "progress":      file.IndexProgress,
        "error":         file.IndexError,
        "parsed_at":     file.ParsedAt,
    })
}
```

**方案 2：限制并发 + 优先级队列**

```go
// ===== 并发控制 =====

// 1. 限制同时解析的文件数量
type ParseWorkerPool struct {
    maxConcurrent int           // 最大并发数（建议 2-3）
    activeJobs    int           // 当前活跃任务
    highPriority  chan *ParseTask // 高优先级队列
    lowPriority   chan *ParseTask // 低优先级队列
    mu            sync.Mutex
}

func (p *ParseWorkerPool) Submit(task *ParseTask) error {
    // 根据文件大小/优先级选择队列
    if task.Size > 10*1024*1024 {  // > 10MB
        task.Priority = "low"
        p.lowPriority <- task
    } else {
        p.highPriority <- task
    }
    return nil
}

func (p *ParseWorkerPool) Start() {
    for i := 0; i < p.maxConcurrent; i++ {
        go p.worker()
    }
}

func (p *ParseWorkerPool) worker() {
    for {
        select {
        case task := <-p.highPriority:
            p.process(task)
        case task := <-p.lowPriority:
            // 只有没有高优先级任务时才处理
            time.Sleep(1 * time.Second)
            if len(p.highPriority) == 0 {
                p.process(task)
            }
        }
    }
}
```

**方案 3：分块上传 + 断点续传（大文件）**

```go
// ===== 分块上传 =====

// 1. 初始化上传
// POST /api/files/upload/init
func (h *Handler) InitUpload(w http.ResponseWriter, r *http.Request) {
    req := &InitUploadRequest{
        FileName: "large.pdf",
        FileSize: 50 * 1024 * 1024,  // 50MB
        ChunkSize: 5 * 1024 * 1024,   // 5MB
        TotalChunks: 10,
    }
    
    uploadID := uuid.New().String()
    h.uploadCache.Set(uploadID, req, 1*time.Hour)
    
    response.JSON(w, http.StatusOK, map[string]string{
        "upload_id":    uploadID,
        "chunk_size":   "5242880",  // 5MB
        "total_chunks": "10",
    })
}

// 2. 上传分块
// POST /api/files/upload/chunk
func (h *Handler) UploadChunk(w http.ResponseWriter, r *http.Request) {
    uploadID := r.FormValue("upload_id")
    chunkIndex := r.FormValue("chunk_index")
    file := r.FormFile("file")
    
    // 保存分块
    chunkPath := fmt.Sprintf("/tmp/uploads/%s/chunk_%s", uploadID, chunkIndex)
    saveChunk(chunkPath, file)
    
    // 检查是否所有分块都已上传
    if allChunksUploaded(uploadID) {
        // 合并分块
        mergeChunks(uploadID)
        
        // 触发解析任务
        h.parseQueue.Publish(&ParseTask{UploadID: uploadID})
    }
    
    response.JSON(w, http.StatusOK, map[string]string{
        "status": "chunk_uploaded",
    })
}
```

**方案 4：限制文件大小和类型**

```yaml
# ===== 文件限制配置 =====

file_upload:
  max_size: 104857600        # 100MB（硬限制）
  
  # 按类型限制
  type_limits:
    pdf:     50MB
    docx:    20MB
    xlsx:    10MB
    image:   5MB
    figma:   N/A  # 链接，不限制
  
  # OCR 限制
  ocr:
    max_images: 20            # 单次最多 OCR 20 张图片
    max_size_per_image: 5MB  # 单张图片最大 5MB
    timeout_per_image: 30s   # 单张图片超时
```

### 监控指标

```go
// ===== 文件解析监控 =====

type ParseMetrics struct {
    QueueLength      int           `json:"queue_length"`       // 待处理队列长度
    ActiveJobs       int           `json:"active_jobs"`        // 正在解析的任务数
    AvgParseTime     time.Duration `json:"avg_parse_time"`     // 平均解析时间
    AvgOCRTime       time.Duration `json:"avg_ocr_time"`       // 平均 OCR 时间
    FailedCount      int           `json:"failed_count"`       // 失败数量
    TimeoutCount     int           `json:"timeout_count"`      // 超时数量
}

// 告警阈值
const (
    AlertQueueLength   = 10   // 队列积压超过 10
    AlertAvgParseTime  = 30   // 平均解析时间超过 30s
    AlertFailedRate    = 0.1  // 失败率超过 10%
)
```

### 一期实施优先级

| 优先级 | 方案 | 实施难度 | 收益 | 是否一期 |
|--------|------|---------|------|----------|
| P0 | 异步解析 + 进度反馈 | 中 | 高 | ✅ 必须 |
| P1 | 限制并发 + 优先级队列 | 中 | 高 | ✅ 必须 |
| P2 | 限制文件大小 | 低 | 中 | ✅ 建议 |
| P3 | 分块上传 | 高 | 中 | ⚠️ 可选 |

**一期最小方案**：异步解析 + 限制并发

---

## 3. Milvus 检索瓶颈（中等）

### 问题描述

**严重程度**：🟠 中等

**核心问题**：
- 10万+ 文档块时检索延迟增加
- 并发检索时性能下降
- 大批量相似度计算耗时

### 触发条件

```
场景 1：大量文档累积
- 项目累积 100+ 需求文档
- 分块后 10万+ 文档块
- 单次检索：100-300ms
- 并发 10 个检索：500ms-2s

场景 2：高并发 RAG 查询
- 多个用户同时使用 AI 生成
- 每个生成需要检索向量
- Milvus 负载升高
- 检索延迟增加到秒级

场景 3：复杂查询
- 需要过滤条件（project_id + tags）
- 向量搜索 + 标量过滤混合
- 性能进一步下降
```

### 性能分析

```yaml
Milvus 性能特征:
  数据量       | 单次检索 | 并发 10 | 并发 50
  -------------|---------|---------|----------
  < 1万        | 10-20ms | 50-100ms| 200-500ms
  1万-10万     | 20-50ms | 100-300ms| 500ms-2s
  10万-100万   | 50-150ms| 300ms-1s| 1s-5s
  > 100万      | 100-300ms| 1s-3s  | 5s-10s

瓶颈阈值:
  文档块数量: > 10万        # 检索延迟明显
  并发检索: > 5             # 性能下降
  检索延迟: > 200ms         # 用户感知卡顿
```

### 一期处理方案（建议处理）

**方案 1：优化索引配置**

```python
# ===== Milvus 索引优化 =====

# 当前配置（可能不够优化）
index_params = {
    "index_type": "IVF_FLAT",
    "metric_type": "L2",
    "params": {"nlist": 128}
}

# 优化方案 1：调整 nlist 参数
# nlist 应该是数据量的平方根
import math

def get_nlist(num_vectors):
    return max(64, int(math.sqrt(num_vectors)))

# 优化方案 2：使用 HNSW 索引（更快的查询）
index_params = {
    "index_type": "HNSW",
    "metric_type": "L2",
    "params": {
        "M": 16,                # 每个节点的最大连接数
        "efConstruction": 200  # 构建索引时的搜索范围
    }
}

# 搜索时调整 ef 参数
search_params = {
    "metric_type": "L2",
    "params": {
        "ef": 64  # ef >= top_k * 2，越大越准确但越慢
    }
}
```

**方案 2：缓存热门查询**

```go
// ===== RAG 结果缓存 =====

func (s *Service) SearchChunks(ctx context.Context, query string, projectID string) ([]*Chunk, error) {
    // 1. 生成缓存键
    cacheKey := fmt.Sprintf("rag:search:%s:%s", projectID, hashQuery(query))
    
    // 2. 尝试从缓存获取
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil {
        return cached.([]*Chunk), nil
    }
    
    // 3. 缓存未命中，执行搜索
    chunks, err := s.milvusClient.Search(ctx, query, projectID, 5)
    if err != nil {
        return nil, err
    }
    
    // 4. 写入缓存（10 分钟）
    s.cache.Set(ctx, cacheKey, chunks, 10*time.Minute)
    
    return chunks, nil
}
```

**方案 3：限制返回数量**

```go
// ===== 限制检索数量 =====

type SearchConfig struct {
    TopK         int  `json:"top_k"`          // 返回数量，默认 5
    MaxTokens    int  `json:"max_tokens"`     // 最大 Token 数
    ScoreThreshold float64 `json:"score_threshold"` // 相似度阈值
}

// 根据场景调整 TopK
func getTopKForScenario(scenario string) int {
    switch scenario {
    case "ai_generate":
        return 3  // AI 生成只需要最相关的 3 个
    case "ai_analyze":
        return 5  // AI 分析需要更多上下文
    case "user_search":
        return 10 // 用户搜索可以多返回一些
    default:
        return 5
    }
}
```

**方案 4：定期清理旧数据**

```go
// ===== 数据清理策略 =====

// 1. 定期归档旧文档块（6 个月以上）
func (s *Service) ArchiveOldChunks() error {
    cutoff := time.Now().AddDate(0, -6, 0)
    
    // 2. 查询旧文档块
    oldChunks, err := s.chunkRepo.FindOlderThan(context.Background(), cutoff)
    if err != nil {
        return err
    }
    
    // 3. 从 Milvus 删除
    for _, chunk := range oldChunks {
        s.milvusClient.Delete(chunk.ID)
    }
    
    // 4. 从 PostgreSQL 标记为已归档
    s.chunkRepo.Archive(oldChunks)
    
    return nil
}

// 定时任务：每月执行一次
// SELECT cron.schedule('archive-old-chunks', '0 2 1 * *', 'SELECT archive_old_chunks()');
```

### 监控指标

```go
// ===== Milvus 监控 =====

type MilvusMetrics struct {
    CollectionCount  int           `json:"collection_count"`   // 文档块总数
    AvgSearchLatency time.Duration `json:"avg_search_latency"` // 平均检索延迟
    P95SearchLatency time.Duration `json:"p95_search_latency"` // P95 延迟
    QPS              int           `json:"qps"`                // 每秒查询数
    CacheHitRate     float64       `json:"cache_hit_rate"`     // 缓存命中率
}

// 告警阈值
const (
    AlertAvgLatency = 200 * time.Millisecond  // 平均延迟 > 200ms
    AlertP95Latency  = 500 * time.Millisecond  // P95 延迟 > 500ms
    AlertCacheHit    = 0.5                       // 缓存命中率 < 50%
)
```

### 一期实施优先级

| 优先级 | 方案 | 实施难度 | 收益 | 是否一期 |
|--------|------|---------|------|----------|
| P0 | 优化索引配置 | 低 | 高 | ✅ 建议 |
| P1 | 缓存热门查询 | 低 | 中 | ⚠️ 可选 |
| P2 | 限制返回数量 | 低 | 低 | ⚠️ 可选 |
| P3 | 定期清理数据 | 中 | 中 | ⚠️ 可选 |

**一期最小方案**：优化索引配置（HNSW）

---

## 4. PostgreSQL 查询瓶颈（中等）

### 问题描述

**严重程度**：🟠 中等

**核心问题**：
- `execution_results` 表增长快（100万+ 行）
- 复杂报表查询慢
- 大表 JOIN 性能差

### 触发条件

```
场景 1：执行历史查询
- 查询某个测试用例的历史执行记录
- execution_results 表 100万+ 行
- 查询时间：500ms-2s

场景 2：复杂报表
- 生成测试计划执行报告
- 需要关联多个表
- 查询时间：1-3s

场景 3：列表分页查询
- 查看所有执行记录
- 使用 OFFSET 分页
- 越往后越慢
```

### 性能分析

```yaml
PostgreSQL 表增长预估:
  表名              | 1年后   | 2年后   | 查询性能
  ------------------|---------|---------|----------
  test_cases        | 5万     | 10万    | 良好
  execution_results | 50万    | 100万   | 开始变慢
  document_chunks   | 10万    | 20万    | 良好
  files             | 1000    | 2000    | 良好

瓶颈查询:
  - 历史执行记录查询（无索引或索引不当）
  - 复杂报表（多表 JOIN）
  - 大 OFFSET 分页（OFFSET 10000+）
```

### 一期处理方案（可选处理）

**方案 1：按月分区 execution_results**

```sql
-- ===== 分区策略 =====

-- 1. 创建分区表
CREATE TABLE execution_results (
    id UUID DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL,
    test_case_id UUID NOT NULL,
    executor_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    bug_id VARCHAR(255),
    bug_url VARCHAR(500),
    notes TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (executed_at);

-- 2. 创建分区（提前创建）
CREATE TABLE execution_results_2025_01 PARTITION OF execution_results
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE execution_results_2025_02 PARTITION OF execution_results
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- 3. 自动创建分区（脚本或 pg_cron）
```

**方案 2：物化视图加速报表**

```sql
-- ===== 物化视图 =====

-- 1. 创建物化视图（常用报表）
CREATE MATERIALIZED VIEW mv_test_case_stats AS
SELECT 
    tc.id,
    tc.title,
    tc.status,
    COUNT(er.id) as execution_count,
    COUNT(CASE WHEN er.status = 'passed' THEN 1 END) as passed_count,
    COUNT(CASE WHEN er.status = 'failed' THEN 1 END) as failed_count,
    MAX(er.executed_at) as last_executed_at
FROM test_cases tc
LEFT JOIN execution_results er ON tc.id = er.test_case_id
GROUP BY tc.id, tc.title, tc.status;

-- 2. 创建索引（加速刷新）
CREATE UNIQUE INDEX ON mv_test_case_stats(id);

-- 3. 定期刷新（每小时）
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_test_case_stats;
```

**方案 3：使用游标分页**

```sql
-- ===== 游标分页 =====

-- ❌ 错误：大 OFFSET
SELECT * FROM execution_results
ORDER BY executed_at DESC
LIMIT 20 OFFSET 10000;  -- 很慢

-- ✅ 正确：游标分页
SELECT * FROM execution_results
WHERE executed_at < '2025-05-01T00:00:00Z'
ORDER BY executed_at DESC
LIMIT 20;
```

### 监控指标

```sql
-- ===== 慢查询监控 =====

-- 1. 查看慢查询
SELECT 
    query,
    calls,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 500  -- 超过 500ms
ORDER BY mean_time DESC
LIMIT 20;

-- 2. 查看表大小
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 3. 查看未使用的索引
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey';
```

### 一期实施优先级

| 优先级 | 方案 | 实施难度 | 收益 | 是否一期 |
|--------|------|---------|------|----------|
| P0 | 游标分页 | 低 | 中 | ⚠️ 可选 |
| P1 | 分区表 | 中 | 高 | ⚠️ 可选 |
| P2 | 物化视图 | 中 | 中 | ❌ 不处理 |

**一期建议**：暂不处理，数据量增长后再优化

---

## 5. 内存资源瓶颈（轻度）

### 问题描述

**严重程度**：🟡 轻度

**核心问题**：
- 8GB 内存需要运行所有服务
- 缓存增长可能导致 OOM
- AI 调用和文件解析占用大量内存

### 触发条件

```
场景 1：缓存增长
- Redis 缓存大量数据
- 用例列表缓存、模块树缓存等
- 缓存未设置过期或过期时间太长

场景 2：并发 AI 调用
- 10 个并发 AI 调用
- 每个 Worker 占用内存
- 总内存占用可能超过可用

场景 3：大文件处理
- 处理 100MB PDF 文件
- 文件加载到内存解析
- 峰值内存占用高
```

### 性能分析

```yaml
内存占用估算:
  服务          | 常驻内存 | 峰值内存 | 说明
  --------------|---------|----------|----------
  PostgreSQL    | 512MB   | 2GB      | shared_buffers=2GB
  Redis         | 128MB   | 1GB      | 缓存增长
  Milvus        | 512MB   | 2GB      | 向量索引加载
  etcd          | 64MB    | 256MB    | 元数据
  MinIO         | 128MB   | 512MB    | 对象存储
  Backend       | 256MB   | 1GB      | Go 服务
  AI Workers    | 512MB   | 2GB      | 并发调用
  --------------|---------|----------|----------
  总计          | 2.1GB   | 10.5GB   | 超过 8GB

瓶颈阈值:
  可用内存 < 1GB            # 开始 OOM
  Redis 内存 > 1GB          # 需要清理
  AI Workers 内存 > 1.5GB   # 需要限制并发
```

### 一期处理方案（可选处理）

**方案 1：限制缓存大小**

```go
// ===== Redis 缓存限制 =====

// 1. 设置最大内存
// redis.conf: maxmemory 1gb

// 2. 设置淘汰策略
// redis.conf: maxmemory-policy allkeys-lru

// 3. 监控内存使用
func (s *Service) MonitorCacheMemory() {
    for {
        info := s.redisClient.Info("memory")
        usedMemory := info["used_memory"]
        
        // 超过 800MB 时告警
        if usedMemory > 800*1024*1024 {
            log.Warn("Redis memory usage high", "used", usedMemory)
            // 可选：主动清理部分缓存
        }
        
        time.Sleep(5 * time.Minute)
    }
}
```

**方案 2：限制 AI 并发数**

```yaml
# ===== Worker Pool 配置 =====

ai:
  # 一期配置（保守）
  pool_max_workers: 5         # 降低并发
  pool_queue_size: 50         # 减少队列
  
  # 内存监控
  memory_limit: 1GB           # 软限制
  memory_warning: 800MB        # 告警阈值
```

**方案 3：定期清理过期数据**

```go
// ===== 定期清理 =====

// 1. 清理过期缓存
func (s *Service) CleanExpiredCache() {
    // 清理 7 天前的 AI 任务缓存
    pattern := "ai:task:*"
    s.redisClient.Del(pattern)
    
    // 清理 30 天前的统计缓存
    pattern = "project:*:stats"
    s.redisClient.Del(pattern)
}

// 2. 定时任务：每天凌晨 3 点执行
// SELECT cron.schedule('clean-cache', '0 3 * * *', 'SELECT clean_expired_cache()');
```

### 监控指标

```go
// ===== 内存监控 =====

type MemoryMetrics struct {
    TotalMemory     uint64  `json:"total_memory"`      // 总内存
    UsedMemory      uint64  `json:"used_memory"`       // 已使用
    AvailableMemory uint64  `json:"available_memory"`  // 可用
    UsedPercent     float64 `json:"used_percent"`      // 使用百分比
    
    RedisMemory     uint64  `json:"redis_memory"`      // Redis 内存
    BackendMemory   uint64  `json:"backend_memory"`    // Backend 内存
    WorkerMemory    uint64  `json:"worker_memory"`     // Worker 内存
}

// 告警阈值
const (
    AlertMemoryPercent = 0.8   # 内存使用 > 80%
    AlertAvailableMem  = 1 * 1024 * 1024 * 1024  # 可用内存 < 1GB
)
```

### 一期实施优先级

| 优先级 | 方案 | 实施难度 | 收益 | 是否一期 |
|--------|------|---------|------|----------|
| P0 | 限制缓存大小 | 低 | 中 | ⚠️ 可选 |
| P1 | 限制 AI 并发 | 低 | 中 | ⚠️ 可选 |
| P2 | 定期清理 | 低 | 低 | ❌ 不处理 |

**一期建议**：暂不处理，监控即可

---

## 6. 网络带宽瓶颈（轻度）

### 问题描述

**严重程度**：🟡 轻度

**核心问题**：
- 多用户同时上传文件占用带宽
- AI API 调用占用出口带宽
- 内网通信（Milvus、PostgreSQL）可能成为瓶颈

### 触发条件

```
场景 1：多人同时上传
- 5 个用户同时上传 10MB 文件
- 占用 50MB 带宽
- 如果是 100Mbps 网络：约 4 秒

场景 2：AI 调用高峰
- 大量 LLM 请求
- 每个 Prompt + Response 约 10KB
- 100 个并发 = 1MB/s
```

### 一期处理方案（不处理）

**方案 1：限制上传速度**

```go
// ===== 限速上传 =====

// 使用令牌桶算法限速
type RateLimitedWriter struct {
    writer    io.Writer
    rate      int  // 字节/秒
    burst     int  // 突发大小
    tokens    int
    lastTime  time.Time
}

func (w *RateLimitedWriter) Write(p []byte) (int, error) {
    // 限速逻辑
    // ...
}
```

**一期建议**：不处理，网络通常不是瓶颈

---

## 一期处理优先级总结

### 必须处理（影响核心功能）

| 瓶颈 | 处理方案 | 预计工作量 | 关键指标 |
|------|---------|-----------|----------|
| AI 调用 | 异步任务 + 轮询 | 3-5 天 | 队列积压 < 10 |
| 文件解析 | 异步解析 + 限流 | 2-3 天 | 解析积压 < 5 |

### 建议处理（提升体验）

| 瓶颈 | 处理方案 | 预计工作量 | 关键指标 |
|------|---------|-----------|----------|
| Milvus | 优化索引 | 1 天 | 检索延迟 < 200ms |
| PostgreSQL | 游标分页 | 1 天 | 分页响应 < 100ms |

### 可选处理（未来优化）

| 瓶颈 | 处理方案 | 预计工作量 | 触发条件 |
|------|---------|-----------|----------|
| 内存 | 限制缓存 + 监控 | 2-3 天 | 内存使用 > 80% |
| 网络 | 限速上传 | 1-2 天 | 带宽占用 > 80% |

---

## 监控和告警

### 核心监控指标

```yaml
# ===== 核心监控指标 =====

系统级:
  - CPU 使用率
  - 内存使用率
  - 磁盘使用率
  - 网络带宽

应用级:
  - AI Worker Pool 状态
  - 文件解析队列长度
  - Milvus 检索延迟
  - PostgreSQL 慢查询数
  - Redis 命中率

业务级:
  - AI 任务成功率
  - 文件解析成功率
  - API 响应时间（P50, P95, P99）
  - 错误率
```

### 告警规则

```yaml
# ===== 告警规则 =====

严重告警（立即处理）:
  - AI 队列积压 > 50
  - 文件解析失败率 > 20%
  - API 错误率 > 5%
  - 内存使用率 > 90%
  - 磁盘使用率 > 85%

警告告警（关注）:
  - AI 队列积压 > 20
  - Milvus 检索延迟 > 500ms
  - API P95 响应 > 1s
  - Redis 内存使用 > 1GB
```

---

## 7. 性能测试方案

### 7.1 压测工具选型

使用 [k6](https://k6.io/) 进行负载测试：

```bash
# 安装
brew install k6   # macOS
# 或
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C547657E45
sudo add-apt-repository 'deb https://dl.k6.io/deb stable main'
sudo apt-get install k6
```

### 7.2 基准场景定义

```javascript
// scripts/load-test/heka-baseline.js
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_TOKEN = __ENV.API_TOKEN || '';

export const options = {
    scenarios: {
        // 场景 1：正常负载
        normal_load: {
            executor: 'constant-arrival-rate',
            rate: 10,           // 10 req/s
            timeUnit: '1s',
            duration: '5m',
            preAllocatedVUs: 20,
        },
        // 场景 2：峰值负载
        peak_load: {
            executor: 'ramping-arrival-rate',
            startRate: 10,
            timeUnit: '1s',
            stages: [
                { duration: '2m', target: 50 },   // 爬坡到 50 req/s
                { duration: '3m', target: 50 },   // 维持 50 req/s
                { duration: '1m', target: 10 },   // 降回正常
            ],
            preAllocatedVUs: 50,
        },
    },
    thresholds: {
        http_req_duration: ['p(50)<200', 'p(95)<1000', 'p(99)<3000'],
        http_req_failed: ['rate<0.05'],
    },
};

const params = {
    headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${API_TOKEN}`,
    },
};

export default function () {
    // 用例列表查询
    const listRes = http.get(`${BASE_URL}/api/v1/testcases?project_id=test-project&page=1&page_size=20`, params);
    check(listRes, { 'list status 200': (r) => r.status === 200 });
    sleep(1);
}
```

### 7.3 关键场景测试脚本

```javascript
// scripts/load-test/ai-generate.js
// AI 生成场景（低并发，长耗时）
export const options = {
    scenarios: {
        ai_generate: {
            executor: 'constant-arrival-rate',
            rate: 2,            // 2 req/s（AI 限流）
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 10,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<60000'],  // AI 调用 P95 < 60s
    },
};
```

### 7.4 执行命令

```bash
# 基准测试
k6 run -e BASE_URL=http://localhost:8080 -e API_TOKEN=xxx scripts/load-test/heka-baseline.js

# AI 专项测试
k6 run -e BASE_URL=http://localhost:8080 -e API_TOKEN=xxx scripts/load-test/ai-generate.js

# 输出报告
k6 run --out json=results.json scripts/load-test/heka-baseline.js
```

---

## 8. 性能基线指标

### 8.1 API 响应时间目标

| API 端点 | P50 | P95 | P99 | 说明 |
|----------|-----|-----|-----|------|
| `GET /api/v1/testcases` | < 100ms | < 300ms | < 500ms | 列表查询（含缓存） |
| `GET /api/v1/testcases/:id` | < 50ms | < 100ms | < 200ms | 详情查询（含缓存） |
| `POST /api/v1/testcases` | < 200ms | < 500ms | < 1s | 创建用例 |
| `PUT /api/v1/testcases/:id` | < 200ms | < 500ms | < 1s | 更新用例 |
| `GET /api/v1/testplans/:id` | < 100ms | < 300ms | < 500ms | 计划详情 |
| `POST /api/v1/executions/:id/results` | < 100ms | < 300ms | < 500ms | 提交执行结果 |
| `POST /api/v1/files/upload` | < 1s | < 5s | < 10s | 文件上传（取决于大小） |
| `POST /api/v1/ai/generate-testcases` | < 15s | < 30s | < 60s | AI 生成（不含排队） |
| `POST /api/v1/ai/analyze` | < 10s | < 20s | < 45s | AI 分析 |
| `GET /api/v1/reports/*` | < 500ms | < 2s | < 5s | 报表查询 |

### 8.2 系统资源基线

| 指标 | 正常值 | 告警阈值 | 危险阈值 |
|------|--------|----------|----------|
| CPU 使用率 | < 40% | > 70% | > 90% |
| 内存使用率 | < 60% | > 80% | > 90% |
| PostgreSQL 连接数 | < 20 | > 40 | > 45 |
| Redis 内存 | < 200MB | > 400MB | > 450MB |
| Milvus 查询延迟 | < 50ms | > 200ms | > 500ms |
| 磁盘使用率 | < 50% | > 75% | > 85% |

### 8.3 AI Worker 吞吐评估

**5 Worker 配置**（MVP）：

| 场景 | 单次耗时 | 并发吞吐 | 排队延迟 |
|------|---------|---------|----------|
| AI 用例生成 | 15-30s | 10-20 请求/分钟 | 0-30s |
| AI 智能分析 | 10-20s | 15-30 请求/分钟 | 0-20s |
| Embedding 生成 | 2-5s | 60-150 请求/分钟 | 0-5s |

**排队延迟估算**：
- 5 Worker 满载时，第 6 个请求开始排队
- 平均等待时间 = 队列长度 x 平均处理时间 / Worker 数
- 建议队列长度不超过 50，防止内存溢出

### 8.4 缓存 TTL 优化

| 数据类型 | 当前 TTL | 优化后 TTL | 原因 |
|----------|---------|-----------|------|
| 用户信息 | 1h | 1h | 不变 |
| 项目信息 | 30min | 30min | 不变 |
| 用例列表 | 5min | 5min | 不变 |
| 用例详情 | 10min | 10min | 不变 |
| 模块树 | 10min | 10min | 不变 |
| RAG 查询结果 | 10min | **1h**（热）/ 10min（冷） | 区分冷热数据 |
| AI 生成结果 | - | **1h** | 新增：确定性 Prompt 缓存 |
| Embedding 结果 | - | **24h** | 新增：相同文本结果不变 |

### 8.5 性能回归检测

```bash
# 每次发版前执行基准测试，对比上次结果
k6 run --out json=results_v1.1.json scripts/load-test/heka-baseline.js

# 对比结果（使用 k6 报告或自定义脚本）
# 关注 P95 退化超过 20% 的端点
```

---

## 9. MVP 性能优化行动清单

| 优先级 | 优化项 | 预期效果 | 实施复杂度 |
|--------|--------|----------|-----------|
| P0 | AI Worker 降到 5 | 内存可控 | 低 |
| P0 | RAG 缓存 TTL 分级 | 减少重复检索 | 低 |
| P1 | API 热点查询加缓存 | P95 降低 50% | 中 |
| P1 | execution_results 按月分区 | 查询性能稳定 | 中 |
| P2 | 用例列表覆盖索引 | 避免回表 | 低 |
| P2 | 文件解析异步化 | 上传不阻塞 | 中 |

---

**文档版本**：v1.0
**最后更新**：2025-05-15
