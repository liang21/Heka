# Heka AI 测试管理平台

> 简化版 AI 测试管理平台，基于 MeterSphere 理念但大幅简化
> 目标用户：20-50 人内部团队，本地部署

---

## 项目概述

Heka 是一个**单体架构**的测试管理平台，核心特点是：
- **简化**：聚焦核心测试管理功能，去除冗余
- **AI 赋能**：AI 辅助用例生成、智能分析、RAG 检索
- **一人可维护**：清晰的代码组织，严格的分层架构
- **本地部署**：数据安全，响应快速

**核心功能**：
- 测试用例管理（模块、用例、步骤、标签）
- 测试计划与执行（迭代执行、结果记录）
- 文件管理 + RAG（PDF/Word/Excel/Image/Figma）
- AI 用例生成（基于需求文档）
- AI 智能分析（变更影响分析、回归推荐）

**技术栈**：
- 前端：React 18 + TypeScript + Tailwind CSS
- 后端：Go 1.21+ + GORM
- 数据库：PostgreSQL 15+ + Redis 7+ + Milvus 2.3+
- AI：Claude / OpenAI / Gemini / Ollama（多 Provider 故障转移）

---

## 架构原则

### 分层架构（严格遵守）

```
Interface 层 (HTTP Handler)
    ↓ 调用
Application 层 (Service 编排)
    ↓ 调用
Domain 层 (Entity + Repository 接口)
    ↑ 实现
Infrastructure 层 (PostgreSQL/Redis/Milvus 实现)
```

**依赖规则**：
- ✅ Interface → Application → Domain
- ✅ Infrastructure → Domain（实现接口）
- ❌ Domain 不依赖任何其他层
- ❌ Application 不依赖 Infrastructure 实现
- ❌ Interface 不跳过 Application 直接调用 Domain

### 领域划分

| 领域 | 职责 | 核心实体 |
|------|------|----------|
| `testcase` | 测试用例管理 | TestCase, Step, Module |
| `plan` | 测试计划管理 | TestPlan, PlanTestCase |
| `execution` | 执行记录管理 | TestExecution, ExecutionResult |
| `file` | 文件管理 | File, FileVersion |
| `rag` | RAG 向量检索 | Chunk, Embedding |
| `user` | 用户管理 | User |
| `project` | 项目管理 | Project, ProjectMember |
| `shared` | 共享概念 | ID, Error, Timestamp |

---

## 代码组织规范

### 目录结构

```
heka-backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口，依赖注入
├── internal/
│   ├── domain/                     # 领域层：核心业务逻辑
│   │   ├── testcase/               #   entity.go, valueobject.go, repository.go, service.go
│   │   ├── plan/
│   │   ├── execution/
│   │   ├── file/
│   │   ├── rag/
│   │   ├── user/
│   │   ├── project/
│   │   └── shared/                 #   errors.go, types.go
│   ├── application/                # 应用层：业务编排
│   │   ├── testcase/               #   service.go, dto.go
│   │   ├── plan/
│   │   ├── execution/
│   │   ├── file/
│   │   ├── rag/
│   │   ├── ai/                     #   AI 编排服务
│   │   ├── user/
│   │   └── project/
│   ├── infrastructure/             # 基础设施层：技术实现
│   │   ├── persistence/
│   │   │   ├── postgres/           #   各 Repository 实现
│   │   │   └── milvus/             #   VectorRepository 实现
│   │   ├── cache/
│   │   │   └── redis.go
│   │   ├── ai/                     #   AI 客户端实现
│   │   ├── storage/
│   │   └── figma/
│   ├── interface/                  # 接口层：对外暴露
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   ├── handler/
│   │   │   └── response/
│   │   └── dto/
│   └── shared/                     # 内部共享工具
│       ├── config/
│       ├── logger/
│       └── validator/
├── pkg/                            # 可对外暴露的包
└── scripts/
    ├── migration/
    └── deploy/
```

### 文件命名约定

| 文件类型 | 命名 | 示例 |
|---------|------|------|
| 实体定义 | `entity.go` | `testcase/entity.go` |
| 值对象 | `valueobject.go` | `testcase/valueobject.go` |
| 仓储接口 | `repository.go` | `testcase/repository.go` |
| 应用服务 | `service.go` | `application/testcase/service.go` |
| DTO | `dto.go` | `application/testcase/dto.go` |
| HTTP Handler | `handler.go` 或 `{resource}.go` | `interface/http/handler/testcase.go` |
| 仓储实现 | `{tech}{resource}.go` | `postgres/testcase.go` |

### 包命名约定

- Domain 层：`internal/domain/<module>`
- Application 层：`internal/application/<module>`
- Infrastructure 层：`internal/infrastructure/<category>/<tech>`
- Interface 层：`internal/interface/http/<category>`

### 依赖注入示例

```go
// cmd/server/main.go
func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化基础设施
    db := postgres.NewDB(cfg.Database.DSN)
    redisClient := redis.NewClient(cfg.Redis.URL)
    milvusClient := milvus.NewClient(cfg.Milvus.Address)

    // 3. 初始化仓储（Domain 接口实现）
    tcRepo := postgres.NewTestCaseRepository(db)
    fileRepo := postgres.NewFileRepository(db)
    ragRepo := milvus.NewVectorRepository(milvusClient)

    // 4. 初始化应用服务
    tcSvc := testcase.NewService(tcRepo, ragRepo)
    fileSvc := file.NewService(fileRepo, ...)
    aiSvc := ai.NewService(fileSvc, ragRepo, tcSvc, ...)

    // 5. 初始化 HTTP 接口
    router := http.NewRouter(tcSvc, aiSvc, ...)

    // 6. 启动服务
    server := &http.Server{
        Addr:    ":8080",
        Handler: router.Setup(),
    }
    log.Fatal(server.ListenAndServe())
}
```

---

## 编码规范（基于 Uber Go Style Guide）

> 完整规范请参考：`specs/uber-go-style-guide.md`

### 命名规范

#### 1. 包命名

```go
// ✅ 全小写、单数、简短
package testcase
package file
package user

// ❌ 避免
package testCases     // 不要驼峰
package utils         // 不要无意义名称
package common        // 不要通用名称
package lib           // 不要 lib
```

#### 2. 错误命名

```go
// ✅ 导出错误用 Err 前缀
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)

// ✅ 未导出错误用 err 前缀（无需下划线）
var errNotFound = errors.New("not found")

// ✅ 自定义错误类型用 Error 后缀
type NotFoundError struct {
    File string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("file %q not found", e.File)
}
```

#### 3. 避免遮蔽内置名称

```go
// ❌ 遮蔽内置类型
type Foo struct {
    error  error  // 与 builtin error 冲突
    string string // 与 builtin string 冲突
}

// ✅ 使用描述性名称
type Foo struct {
    err error
    msg string
}
```

#### 4. 未导出全局变量使用 _ 前缀

```go
// ✅ 未导出全局变量用 _ 前缀
var (
    _defaultTimeout = 30 * time.Second
    _maxRetries     = 3
)

// ❌ 容易误用
var (
    defaultTimeout = 30 * time.Second
    maxRetries     = 3
)
```

#### 5. 接口实现编译验证

```go
// ✅ 编译时验证接口实现
var _ http.Handler = (*Handler)(nil)

type Handler struct {
    // ...
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### 错误处理

#### 6. 错误只处理一次

```go
// ❌ 既记录日志又返回错误（重复处理）
user, err := getUser(id)
if err != nil {
    log.Printf("failed to get user: %v", err) // 多余的日志
    return err
}

// ✅ 只返回错误，让调用者处理
user, err := getUser(id)
if err != nil {
    return fmt.Errorf("get user %q: %w", id, err)
}

// ✅ 在最外层记录日志并退出
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
```

#### 7. 错误包装使用 %w

```go
// ✅ 使用 %w 保留底层错误类型（支持 errors.Is/As）
return fmt.Errorf("open file: %w", err)

// ❌ 使用 %v 会丢失底层错误类型
return fmt.Errorf("open file: %v", err)

// 使用示例
if errors.Is(err, os.ErrNotExist) {
    // 文件不存在
}
```

#### 8. 错误信息简短

```go
// ❌ 冗余的错误信息
return fmt.Errorf("failed to open file: %w", err)
return fmt.Errorf("failed to create user: %w", err)

// ✅ 简洁的错误信息
return fmt.Errorf("open file: %w", err)
return fmt.Errorf("create user: %w", err)
```

#### 9. 生产代码不要 Panic

```go
// ❌ 生产代码使用 panic
func run(args []string) {
    if len(args) == 0 {
        panic("an argument is required")
    }
}

// ✅ 返回错误
func run(args []string) error {
    if len(args) == 0 {
        return errors.New("an argument is required")
    }
    return nil
}

// ✅ 只在 main() 中处理退出
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
```

#### 10. 类型断言使用 comma ok

```go
// ❌ 单返回值会 panic
t := i.(string)

// ✅ 使用 comma ok
t, ok := i.(string)
if !ok {
    return fmt.Errorf("expected string, got %T", i)
}
```

#### 11. 只在 main() 中 Exit

```go
// ❌ 在普通函数中直接退出
func readFile(path string) string {
    f, err := os.Open(path)
    if err != nil {
        log.Fatal(err) // 无法测试，跳过 defer
    }
    // ...
}

// ✅ 返回错误
func readFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()
    // ...
}
```

### 并发

#### 12. Mutex 零值有效

```go
// ✅ Mutex 零值有效，无需指针
type SafeMap struct {
    mu   sync.Mutex
    data map[string]string
}

func (m *SafeMap) Get(key string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[key]
}

// ❌ 不要用指针
mu := new(sync.Mutex)
mu.Lock()
```

#### 13. 不要内嵌 Mutex

```go
// ❌ 内嵌会泄露方法
type SMap struct {
    sync.Mutex  // SMap.Lock/Unlock 意外暴露
    data        map[string]string
}

// ✅ 使用独立字段
type SMap struct {
    mu   sync.Mutex
    data map[string]string
}
```

#### 14. Channel 大小为 1 或 0

```go
// ✅ 无缓冲 channel（默认，size=0）
ch := make(chan int)

// ✅ 缓冲为 1
ch := make(chan int, 1)

// ❌ 避免任意大小的缓冲
ch := make(chan int, 64) // 除非有明确理由
```

#### 15. Goroutine 必须可停止

```go
// ❌ 无法停止的 goroutine
go func() {
    for {
        flush()
        time.Sleep(delay)
    }
}()

// ✅ 可停止的 goroutine
var (
    stop = make(chan struct{})
    done = make(chan struct{})
)

go func() {
    defer close(done)
    ticker := time.NewTicker(delay)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            flush()
        case <-stop:
            return
        }
    }
}()

// 停止
close(stop)
<-done
```

#### 16. 等待 Goroutine 退出

```go
// ✅ 单个 goroutine 使用 done channel
done := make(chan struct{})
go func() {
    defer close(done)
    // ...
}()
<-done

// ✅ 多个 goroutine 使用 WaitGroup
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // ...
    }()
}
wg.Wait()
```

#### 17. 禁止在 init() 中启动 Goroutine

```go
// ❌ init() 中启动 goroutine
func init() {
    go doWork() // 无法控制生命周期
}

// ✅ 显式管理 goroutine 生命周期
type Worker struct {
    stop chan struct{}
    done chan struct{}
}

func NewWorker() *Worker {
    w := &Worker{
        stop: make(chan struct{}),
        done: make(chan struct{}),
    }
    go w.run()
    return w
}

func (w *Worker) Shutdown() {
    close(w.stop)
    <-w.done
}
```

#### 18. 使用 Context 取消

```go
// ✅ 使用 context 控制超时和取消
func fetch(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// 使用
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
data, err := fetch(ctx, "https://example.com")
```

### 日志与资源管理

#### 19. 关键路径必须记录日志

```go
// ✅ 记录关键操作
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        s.logger.Error("create user failed",
            zap.String("email", req.Email),
            zap.Error(err))
        return err
    }

    s.logger.Info("user created",
        zap.String("user_id", user.ID),
        zap.String("email", user.Email))
    return nil
}
```

#### 20. 使用 defer 清理资源

```go
// ✅ 始终使用 defer 清理
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // 确保关闭

    // 使用文件
    return nil
}

// ✅ 多个资源按相反顺序 defer
func process() error {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        return err
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // 先执行

    // ...
    return tx.Commit() // Commit 后 Rollback 无效果
}
```

### 性能优化

#### 21. 指定容器容量

```go
// ✅ 指定 map 容量
m := make(map[string]string, len(files))

// ✅ 指定 slice 容量
data := make([]int, 0, size)

// 性能差异：有预分配 vs 无预分配可差 10 倍
```

#### 22. 边界复制 Slice 和 Map

```go
// ✅ 接收参数时复制
func (d *Driver) SetTrips(trips []Trip) {
    d.trips = make([]Trip, len(trips))
    copy(d.trips, trips)
}

// ✅ 返回结果时复制
func (s *Stats) Snapshot() map[string]int {
    s.mu.Lock()
    defer s.mu.Unlock()

    result := make(map[string]int, len(s.counters))
    for k, v := range s.counters {
        result[k] = v
    }
    return result
}
```

#### 23. 使用 strconv 代替 fmt

```go
// ✅ 更快的类型转换
s := strconv.Itoa(i)
i, err := strconv.Atoi(s)

// ❌ 较慢
s := fmt.Sprint(i)
i, err := fmt.Sscanf(s, "%d")
```

### 其他重要规范

#### 24. 避免在公共结构体中内嵌

```go
// ❌ 内嵌泄露实现细节
type ConcreteList struct {
    *AbstractList
}

// ✅ 使用字段委托
type ConcreteList struct {
    list *AbstractList
}

func (c *ConcreteList) Add(e Entity) {
    c.list.Add(e)
}
```

#### 25. 避免可变全局状态

```go
// ❌ 全局可变状态
var _timeNow = time.Now

func sign(msg string) string {
    now := _timeNow()
    return signWithTime(msg, now)
}

// ✅ 依赖注入
type signer struct {
    now func() time.Time
}

func newSigner() *signer {
    return &signer{
        now: time.Now,
    }
}

func (s *signer) Sign(msg string) string {
    now := s.now()
    return signWithTime(msg, now)
}
```

### 编码规范快速检查清单

在提交代码前，确保：

- [ ] 包名全小写、单数
- [ ] 错误命名使用 Err/err/Error 后缀
- [ ] 错误只处理一次（不既 log 又 return）
- [ ] 生产代码无 panic
- [ ] 类型断言使用 comma ok
- [ ] Mutex 使用零值，不内嵌
- [ ] Channel 大小为 0 或 1
- [ ] 每个 goroutine 可停止
- [ ] 资源清理使用 defer
- [ ] 公共结构体不内嵌类型

---

## 数据库规范

### PostgreSQL 规范

**重要**：所有建表和查询操作必须遵循 `specs/database-performance-spec.md` 规范。

**核心原则**：

1. **通用字段约定**：
   - 所有表必须包含：`id` (UUID PK), `project_id`, `created_by`, `created_at`
   - 可选字段：`updated_by`, `updated_at`, `deleted_at`, `version`

2. **索引设计**：
   - 外键列必须有索引
   - 高频查询条件必须有索引
   - 复合索引遵循最左前缀原则
   - 命名规范：`idx_{table}_{column}_{suffix}`

3. **分页查询**：
   - 大数据量使用游标分页（WHERE id > ?）
   - 避免大 Offset（OFFSET 10000）
   - 只查询需要的列

4. **大表应对**：
   - `test_cases`：10万+ 行，考虑分区
   - `execution_results`：100万+ 行，按月分区 + 6个月归档
   - `document_chunks`：10万+ 行，冷热分离

### Milvus 规范

**Collection 设计**：

```python
# 字段模板
document_chunks_schema = CollectionSchema([
    FieldSchema(name="id", dtype=DataType.VARCHAR, max_length=36, is_primary=True),
    FieldSchema(name="chunk_id", dtype=DataType.VARCHAR, max_length=36),
    FieldSchema(name="project_id", dtype=DataType.VARCHAR, max_length=36),
    FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=65535),
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=1536),
])
```

**索引选择**：
- < 10万：IVF_FLAT (nlist=64)
- 10万-100万：IVF_PQ (nlist=128, m=8)
- > 100万：HNSW (M=16)

### 跨数据库事务

Milvus 不支持 ACID 事务，采用**最终一致性**：

1. 先写入 PostgreSQL（事务保证）
2. 创建待处理记录（index_task 表）
3. 后台 Worker 异步处理 Milvus 索引
4. 定时补偿机制处理失败任务

---

## AI 调用规范

### 多 Provider 故障转移

遵循 `specs/llm-api-calling-strategy.md` 规范。

**核心机制**：
- Worker Pool：控制并发（默认 10 worker）
- 熔断器：连续失败自动熔断
- 指数退避重试：1s → 2s → 4s → 8s（最大 30s）
- 分级超时：Dial 10s, TLS 5s, Response 30s, Request 60s, Generation 55s

**Provider 优先级**：
1. Claude (Priority 1)
2. OpenAI (Priority 2)
3. Gemini (Priority 3)
4. Ollama (Priority 4, 本地兜底)

### 容错配置

```yaml
ai:
  pool_max_workers: 10
  pool_queue_size: 100
  retry_max_attempts: 3
  retry_base_delay: 1s
  retry_max_delay: 30s
  timeout_request: 60s
  timeout_generation: 55s
```

---

## 开发工作流程

### 添加新功能流程

1. **Domain 层**：
   - 定义 Entity（实体）
   - 定义 Value Object（值对象）
   - 定义 Repository 接口

2. **Infrastructure 层**：
   - 实现 Repository 接口（PostgreSQL/Milvus）
   - 编写数据库迁移脚本

3. **Application 层**：
   - 定义 DTO（请求/响应）
   - 实现 Service（业务编排）

4. **Interface 层**：
   - 实现 Handler
   - 注册路由

5. **测试**：
   - 单元测试（Service 层）
   - 集成测试（API 层）

### 命名规范（简化版）

| 类型 | 约定 | 示例 |
|------|------|------|
| 接口 | 名词或 `I` 结尾 | `Repository`, `Service` |
| 实现 | 小写开头 | `testCaseRepository`, `testCaseService` |
| 变量 | 驼峰命名 | `tcSvc`, `tcRepo`, `ctx` |
| 常量 | 大写下划线 | `MAX_UPLOAD_SIZE`, `DEFAULT_TIMEOUT` |
| 私有方法 | 小写开头 | `validateRequest`, `buildPrompt` |

> **详细命名规范**：参见上文"编码规范"章节

### 错误处理（简化版）

```go
// Domain 层错误定义
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized")
)

// 使用示例
func (s *Service) GetTestCase(ctx context.Context, id ID) (*TestCase, error) {
    tc, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, shared.ErrNotFound) {
            return nil, testcase.ErrNotFound
        }
        return nil, shared.Wrap(err, "failed to find test case")
    }
    return tc, nil
}
```

> **详细错误处理规范**：参见上文"编码规范"章节

---

## 关键约定和禁忌

### ✅ 必须遵守

1. **分层架构**：严格遵守依赖方向，禁止跨层调用
2. **数据库规范**：所有建表和查询必须遵循性能规范
3. **错误处理**：使用 domain 层定义的错误，不要裸返回 err
4. **日志记录**：关键操作必须记录日志（创建、更新、删除、AI 调用）
5. **依赖注入**：通过构造函数注入依赖，禁止硬编码
6. **配置管理**：敏感信息使用环境变量
7. **编码规范**：遵守 Uber Go Style Guide（见上文"编码规范"章节）

### ❌ 严格禁止

1. **禁止** Interface 层直接调用 Infrastructure 层
2. **禁止** Domain 层依赖具体技术实现
3. **禁止** Application 层包含业务规则（应在 Domain 层）
4. **禁止** 裸 SQL 字符串拼接（使用参数化查询）
5. **禁止** 忽略错误（所有 err 必须处理）
6. **禁止** 在循环中执行数据库查询（使用批量操作）
7. **禁止** SELECT * 查询（只查询需要的列）
8. **禁止** 大 Offset 分页（使用游标分页）
9. **禁止** 忽略索引检查（建表后必须检查索引）
10. **禁止** 跳过测试（关键路径必须有测试）
11. **禁止** 生产代码使用 panic（返回 error）
12. **禁止** 既记录日志又返回错误（错误只处理一次）

### ⚠️ 需要注意

1. **乐观锁**：更新操作使用 `version` 字段防止并发冲突
2. **软删除**：重要数据使用 `deleted_at` 软删除
3. **缓存策略**：读多写少数据使用 Cache-Aside
4. **异步处理**：耗时操作（文件解析、AI 调用）异步执行
5. **限流保护**：AI 调用使用 Worker Pool 限流
6. **资源清理**：所有资源必须使用 defer 清理
7. **并发控制**：每个 goroutine 必须可停止

---

## 测试要求

### 单元测试

- **覆盖范围**：Application 层 Service 必须 >= 80%
- **Mock 使用**：使用 mock 替代 Infrastructure 层
- **测试框架**：使用 `testify/assert`

```go
func TestService_CreateTestCase(t *testing.T) {
    // Arrange
    mockRepo := new(MockTestCaseRepository)
    svc := testcase.NewService(mockRepo)

    // Act
    result, err := svc.CreateTestCase(ctx, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 集成测试

- **覆盖范围**：关键 API 端点必须有测试
- **测试数据库**：使用 testcontainers 或独立测试数据库
- **清理**：每个测试后清理数据

---

## 部署相关

### 本地开发

```bash
# 启动基础设施
docker-compose up -d postgres redis milvus

# 运行后端
cd backend
go run cmd/server/main.go

# 运行前端
cd frontend
npm run dev
```

### 环境变量

```bash
# .env
POSTGRES_PASSWORD=your_password
JWT_SECRET=your_jwt_secret
CLAUDE_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
```

### 健康检查

```bash
# 健康检查端点
curl http://localhost:8080/api/health

# AI 服务状态
curl http://localhost:8080/api/monitoring/ai
```

---

## 相关文档

详细规范文档请查看 `specs/` 目录：

- `specs/heka-design-doc.md` - 完整设计文档
- `specs/backend-code-organization.md` - 后端代码组织规范
- `specs/uber-go-style-guide.md` - **Uber Go Style Guide 编码规范（25 条关键规则）**
- `specs/llm-api-calling-strategy.md` - LLM 调用技术方案
- `specs/deployment-architecture.md` - 部署架构设计
- `specs/database-performance-spec.md` - 数据库性能规范

---

## 快速参考

### 常用命令

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./internal/application/testcase/

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...

# 构建
go build -o bin/server cmd/server/main.go

# 数据库迁移
go run scripts/migrate/main.go up
```

### 常用查询

```sql
-- 查看表大小
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 查看慢查询
SELECT query, calls, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

---

**最后更新**：2026-05-15
**维护者**：Heka Team
