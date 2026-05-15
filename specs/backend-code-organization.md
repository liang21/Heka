# Heka 后端代码组织规范

> 单体架构，严格分层，模块解耦。
> 技术栈：Go + GORM + PostgreSQL + Redis + Milvus

## 1. 目录结构（完整）

```
heka-backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/
│   ├── domain/                     # 领域层：核心业务逻辑
│   │   ├── testcase/               # 测试用例领域
│   │   │   ├── entity.go           # 实体定义
│   │   │   ├── valueobject.go      # 值对象（Priority、Status 等）
│   │   │   ├── repository.go       # 仓储接口（只定义接口）
│   │   │   └── service.go          # 领域服务（复杂业务逻辑）
│   │   ├── plan/                   # 测试计划领域
│   │   ├── execution/              # 执行记录领域
│   │   ├── file/                   # 文件管理领域
│   │   ├── rag/                    # RAG 领域
│   │   ├── user/                   # 用户领域
│   │   ├── project/                # 项目领域
│   │   └── shared/                 # 共享领域概念
│   │       ├── errors.go           # 领域错误定义
│   │       └── types.go            # 共享类型（ID、Timestamp 等）
│   ├── application/                # 应用层：编排业务流程
│   │   ├── testcase/               # 测试用例应用服务
│   │   │   ├── service.go          # 应用服务实现
│   │   │   └── dto.go              # 请求/响应 DTO
│   │   ├── plan/                   # 测试计划应用服务
│   │   ├── execution/              # 执行记录应用服务
│   │   ├── file/                   # 文件应用服务
│   │   ├── rag/                    # RAG 应用服务
│   │   ├── ai/                     # AI 应用服务（编排生成、分析）
│   │   ├── user/                   # 用户应用服务
│   │   └── project/                # 项目应用服务
│   ├── infrastructure/             # 基础设施层：技术实现
│   │   ├── persistence/            # 持久化
│   │   │   ├── postgres/           # PostgreSQL 实现
│   │   │   │   ├── testcase.go     # TestCaseRepository 实现
│   │   │   │   ├── plan.go
│   │   │   │   ├── execution.go
│   │   │   │   ├── file.go
│   │   │   │   ├── user.go
│   │   │   │   └── project.go
│   │   │   └── milvus/             # Milvus 实现
│   │   │       ├── vector.go       # VectorRepository 实现
│   │   │       └── collection.go   # 集合管理
│   │   ├── cache/                  # 缓存
│   │   │   └── redis.go            # Redis 缓存实现
│   │   ├── ai/                     # AI 服务
│   │   │   ├── client.go           # AI 客户端接口
│   │   │   ├── claude.go           # Claude 实现
│   │   │   ├── openai.go           # OpenAI 实现
│   │   │   └── local.go            # 本地模型实现
│   │   ├── storage/                # 文件存储
│   │   │   └── local.go            # 本地文件存储
│   │   └── figma/                  # Figma 集成
│   │       └── client.go           # Figma API 客户端
│   ├── interface/                  # 接口层：对外暴露
│   │   ├── http/                   # HTTP 接口
│   │   │   ├── router.go           # 路由注册
│   │   │   ├── middleware/         # 中间件
│   │   │   │   ├── auth.go         # 认证
│   │   │   │   ├── cors.go         # 跨域
│   │   │   │   └── logger.go       # 日志
│   │   │   ├── handler/            # HTTP 处理器
│   │   │   │   ├── testcase.go
│   │   │   │   ├── plan.go
│   │   │   │   ├── execution.go
│   │   │   │   ├── file.go
│   │   │   │   ├── ai.go
│   │   │   │   ├── user.go
│   │   │   │   └── project.go
│   │   │   └── response/           # 统一响应格式
│   │   │       └── response.go
│   │   └── dto/                    # 接口层 DTO（请求/响应结构）
│   │       └── ...
│   └── shared/                     # 内部共享工具
│       ├── config/                 # 配置加载
│       │   └── config.go
│       ├── logger/                 # 日志封装
│       │   └── logger.go
│       └── validator/              # 验证器
│           └── validator.go
├── pkg/                            # 可对外暴露的包
│   ├── id/                         # ID 生成器
│   └── errors/                     # 错误处理
├── api/                            # API 协议文件（可选）
│   └── openapi/                    # OpenAPI 规范
│       ├── openapi.yaml            # API 规范定义
│       └── schemas/                # JSON Schema
├── configs/                        # 配置文件模板（可选）
│   ├── config.example.yaml         # 配置示例
│   ├── config.dev.yaml             # 开发环境配置
│   └── config.prod.yaml            # 生产环境配置
├── test/                           # 额外测试数据（可选）
│   ├── fixtures/                   # 测试 fixtures
│   ├── testdata/                   # 测试数据文件
│   └── integration/                # 集成测试辅助
└── scripts/                        # 脚本
    ├── migration/                  # 数据库迁移
    └── deploy/                     # 部署脚本
```

### Standard Go Project Layout 可选目录说明

| 目录 | 用途 | 是否必需 |
|------|------|---------|
| `/api` | OpenAPI/Swagger 规范、JSON Schema、Protocol Buffers 文件 | ⚪ 可选 |
| `/configs` | 配置文件模板或默认配置 | ⚪ 可选 |
| `/test` | 额外的外部测试数据和测试工具（与 `_test.go` 分离） | ⚪ 可选 |
| `/docs` | 项目文档、设计文档、用户指南等 | ⚪ 可选 |

**建议**：
- `/api`：如果有 REST API，建议添加 OpenAPI 规范，便于前后端协作和 API 文档生成
- `/configs`：提供配置示例，帮助新开发者快速上手
- `/test`：存放测试 fixtures 和测试数据，保持测试代码整洁
- `/docs`：存放设计文档、架构图、API 文档等（也可使用 `/specs` 目录）

## 2. 分层职责边界

### 2.1 Domain 层（领域层）

**职责**：核心业务逻辑，与技术无关。

**包含**：
- Entity（实体）：具有身份标识的业务对象
- Value Object（值对象）：无身份标识的值类型
- Repository Interface（仓储接口）：定义数据访问契约
- Domain Service（领域服务）：复杂业务规则

**规则**：
- ❌ 不依赖任何外部技术（数据库、HTTP、缓存等）
- ❌ 不依赖 Application、Infrastructure、Interface 层
- ✅ 只能依赖同层的 Shared 模块

**示例 - Entity**：

```go
// internal/domain/testcase/entity.go
package testcase

import (
    "time"
    "heka-backend/internal/domain/shared"
)

type TestCase struct {
    ID          shared.ID
    ProjectID   shared.ID
    Title       string
    Description string
    Steps       []Step
    Priority    Priority
    Status      Status
    Tags        []string
    CreatedBy   shared.ID
    CreatedAt   time.Time
    UpdatedBy   shared.ID
    UpdatedAt   time.Time
    Version     int // 乐观锁
}

type Step struct {
    Number      int
    Action      string
    Expected    string
}

// 领域行为：业务规则封装
func (tc *TestCase) CanExecute() error {
    if tc.Status != StatusReady {
        return ErrTestCaseNotReady
    }
    return nil
}

func (tc *TestCase) AddStep(action, expected string) error {
    if action == "" {
        return ErrInvalidStep
    }
    tc.Steps = append(tc.Steps, Step{
        Number:   len(tc.Steps) + 1,
        Action:   action,
        Expected: expected,
    })
    return nil
}
```

**示例 - Value Object**：

```go
// internal/domain/testcase/valueobject.go
package testcase

import "errors"

type Priority int

const (
    PriorityLow Priority = iota
    PriorityMedium
    PriorityHigh
    PriorityCritical
)

func (p Priority) String() string {
    switch p {
    case PriorityLow:
        return "low"
    case PriorityMedium:
        return "medium"
    case PriorityHigh:
        return "high"
    case PriorityCritical:
        return "critical"
    default:
        return "unknown"
    }
}

func ParsePriority(s string) (Priority, error) {
    switch s {
    case "low":
        return PriorityLow, nil
    case "medium":
        return PriorityMedium, nil
    case "high":
        return PriorityHigh, nil
    case "critical":
        return PriorityCritical, nil
    default:
        return PriorityLow, errors.New("invalid priority")
    }
}

type Status string

const (
    StatusDraft     Status = "draft"
    StatusReady     Status = "ready"
    StatusArchived  Status = "archived"
)

func (s Status) CanTransitionTo(new Status) bool {
    // 状态转换规则
    switch s {
    case StatusDraft:
        return new == StatusReady || new == StatusArchived
    case StatusReady:
        return new == StatusArchived
    default:
        return false
    }
}
```

**示例 - Repository Interface**：

```go
// internal/domain/testcase/repository.go
package testcase

import "context"

type Repository interface {
    Save(ctx context.Context, tc *TestCase) error
    FindByID(ctx context.Context, id shared.ID) (*TestCase, error)
    FindByProject(ctx context.Context, projectID shared.ID, opts FindOptions) ([]*TestCase, int64, error)
    Delete(ctx context.Context, id shared.ID) error
    UpdateStatus(ctx context.Context, id shared.ID, status Status) error
}

type FindOptions struct {
    Page     int
    PageSize int
    Status   Status
    Priority Priority
    Keyword  string
    Tags     []string
    SortBy   string // "created_at", "updated_at", "priority"
    SortDesc bool
}
```

**示例 - Domain Service**：

```go
// internal/domain/testcase/service.go
package testcase

import (
    "context"
    "heka-backend/internal/domain/shared"
)

// Service 处理复杂的业务规则
type Service struct {
    tcRepo Repository
}

func NewService(tcRepo Repository) *Service {
    return &Service{tcRepo: tcRepo}
}

// ValidateTransition 验证状态转换是否符合业务规则
func (s *Service) ValidateTransition(ctx context.Context, tcID shared.ID, newStatus Status) error {
    tc, err := s.tcRepo.FindByID(ctx, tcID)
    if err != nil {
        return err
    }

    if !tc.Status.CanTransitionTo(newStatus) {
        return ErrInvalidStatusTransition
    }

    return nil
}

// CalculatePriority 根据标签和历史数据计算建议优先级
func (s *Service) CalculatePriority(ctx context.Context, tags []string, history []*TestCase) Priority {
    // 复杂业务逻辑
    return PriorityMedium
}
```

### 2.2 Application 层（应用层）

**职责**：编排业务流程，协调多个领域对象。

**包含**：
- Application Service（应用服务）：用例编排
- DTO（数据传输对象）：跨层数据传递

**规则**：
- ✅ 依赖 Domain 层
- ✅ 依赖 Infrastructure 层接口（通过依赖注入）
- ❌ 不直接依赖 Infrastructure 实现细节
- ❌ 不依赖 Interface 层
- ❌ 不包含业务规则（业务规则在 Domain 层）

**示例 - Application Service**：

```go
// internal/application/testcase/service.go
package testcase

import (
    "context"
    "time"
    "heka-backend/internal/application/testcase/dto"
    "heka-backend/internal/domain/file"
    "heka-backend/internal/domain/rag"
    "heka-backend/internal/domain/shared"
    "heka-backend/internal/domain/testcase"
)

type Service struct {
    tcRepo   testcase.Repository
    fileRepo file.Repository
    ragRepo  rag.Repository
    aiClient ai.Client
}

func NewService(
    tcRepo testcase.Repository,
    fileRepo file.Repository,
    ragRepo rag.Repository,
    aiClient ai.Client,
) *Service {
    return &Service{
        tcRepo:   tcRepo,
        fileRepo: fileRepo,
        ragRepo:  ragRepo,
        aiClient: aiClient,
    }
}

// CreateTestCase 创建测试用例（应用服务编排）
func (s *Service) CreateTestCase(ctx context.Context, req dto.CreateTestCaseRequest) (*dto.CreateTestCaseResponse, error) {
    // 1. 验证输入
    if err := s.validateCreateRequest(req); err != nil {
        return nil, err
    }

    // 2. 构建领域对象
    tc := &testcase.TestCase{
        ID:          shared.NewID(),
        ProjectID:   req.ProjectID,
        Title:       req.Title,
        Description: req.Description,
        Steps:       s.toDomainSteps(req.Steps),
        Priority:    testcase.Priority(req.Priority),
        Status:      testcase.StatusDraft,
        Tags:        req.Tags,
        CreatedBy:   req.CreatorID,
        CreatedAt:   time.Now(),
        UpdatedBy:   req.CreatorID,
        UpdatedAt:   time.Now(),
        Version:     0,
    }

    // 3. 调用领域仓储
    if err := s.tcRepo.Save(ctx, tc); err != nil {
        return nil, err
    }

    // 4. 返回 DTO
    return &dto.CreateTestCaseResponse{
        ID:        tc.ID,
        CreatedAt: tc.CreatedAt,
    }, nil
}

// GenerateByAI AI 生成用例（编排 RAG + AI 生成）
func (s *Service) GenerateByAI(ctx context.Context, req dto.GenerateByAIRequest) (*dto.GenerateByAIResponse, error) {
    // 1. 获取文件
    f, err := s.fileRepo.FindByID(ctx, req.FileID)
    if err != nil {
        return nil, err
    }

    // 2. 检查向量是否已生成
    chunks, err := s.ragRepo.FindByFileID(ctx, f.ID)
    if err != nil || len(chunks) == 0 {
        // 首次处理，需要分块和向量化
        chunks, err = s.processAndIndexFile(ctx, f)
        if err != nil {
            return nil, err
        }
    }

    // 3. 检索相关内容
    relevantChunks, err := s.ragRepo.Search(ctx, req.Query, 5)
    if err != nil {
        return nil, err
    }

    // 4. 构建提示词
    prompt := s.buildTestGenerationPrompt(relevantChunks, req.Query)

    // 5. 调用 AI 生成
    aiResponse, err := s.aiClient.GenerateTestCases(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 6. 解析并创建用例
    var createdIDs []shared.ID
    for _, tc := range aiResponse.TestCases {
        req := dto.CreateTestCaseRequest{
            ProjectID:   req.ProjectID,
            Title:       tc.Title,
            Description: tc.Description,
            Steps:       tc.Steps,
            Priority:    tc.Priority,
            Tags:        tc.Tags,
            CreatorID:   req.CreatorID,
        }
        res, err := s.CreateTestCase(ctx, req)
        if err != nil {
            return nil, err
        }
        createdIDs = append(createdIDs, res.ID)
    }

    return &dto.GenerateByAIResponse{
        TestCaseIDs: createdIDs,
        Count:       len(createdIDs),
    }, nil
}

// 私有辅助方法
func (s *Service) validateCreateRequest(req dto.CreateTestCaseRequest) error {
    if req.Title == "" {
        return ErrTitleRequired
    }
    if len(req.Steps) == 0 {
        return ErrAtLeastOneStep
    }
    return nil
}

func (s *Service) toDomainSteps(steps []dto.StepDTO) []testcase.Step {
    result := make([]testcase.Step, len(steps))
    for i, s := range steps {
        result[i] = testcase.Step{
            Number:   i + 1,
            Action:   s.Action,
            Expected: s.Expected,
        }
    }
    return result
}

func (s *Service) processAndIndexFile(ctx context.Context, f *file.File) ([]*rag.Chunk, error) {
    // 文件解析、分块、向量化逻辑
    return nil, nil
}

func (s *Service) buildTestGenerationPrompt(chunks []*rag.Chunk, query string) string {
    return ""
}
```

**示例 - DTO**：

```go
// internal/application/testcase/dto.go
package testcase

import "heka-backend/internal/domain/shared"

// CreateTestCaseRequest 创建测试用例请求
type CreateTestCaseRequest struct {
    ProjectID   shared.ID  `json:"project_id"`
    Title       string     `json:"title" validate:"required"`
    Description string     `json:"description"`
    Steps       []StepDTO  `json:"steps" validate:"required,min=1"`
    Priority    int        `json:"priority" validate:"min=0,max=3"`
    Tags        []string   `json:"tags"`
    CreatorID   shared.ID  `json:"-"` // 从上下文获取
}

// StepDTO 测试步骤 DTO
type StepDTO struct {
    Action   string `json:"action" validate:"required"`
    Expected string `json:"expected" validate:"required"`
}

// CreateTestCaseResponse 创建测试用例响应
type CreateTestCaseResponse struct {
    ID        shared.ID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}

// GenerateByAIRequest AI 生成用例请求
type GenerateByAIRequest struct {
    FileID    shared.ID  `json:"file_id" validate:"required"`
    ProjectID shared.ID  `json:"project_id" validate:"required"`
    Query     string     `json:"query" validate:"required"`
    CreatorID shared.ID  `json:"-"`
}

// GenerateByAIResponse AI 生成用例响应
type GenerateByAIResponse struct {
    TestCaseIDs []shared.ID `json:"test_case_ids"`
    Count       int         `json:"count"`
}

// AITestCaseDTO AI 生成的测试用例结构
type AITestCaseDTO struct {
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Steps       []StepDTO  `json:"steps"`
    Priority    int        `json:"priority"`
    Tags        []string   `json:"tags"`
}

// ListTestCasesRequest 查询测试用列表请求
type ListTestCasesRequest struct {
    ProjectID shared.ID  `json:"project_id" validate:"required"`
    Page      int        `json:"page" validate:"min=1"`
    PageSize  int        `json:"page_size" validate:"min=1,max=100"`
    Status    string     `json:"status"`
    Priority  string     `json:"priority"`
    Keyword   string     `json:"keyword"`
    Tags      []string   `json:"tags"`
    SortBy    string     `json:"sort_by"`
    SortDesc  bool       `json:"sort_desc"`
}

// ListTestCasesResponse 查询测试用例列表响应
type ListTestCasesResponse struct {
    TestCases []TestCaseDTO `json:"test_cases"`
    Total     int64         `json:"total"`
    Page      int           `json:"page"`
    PageSize  int           `json:"page_size"`
}

// TestCaseDTO 测试用例 DTO
type TestCaseDTO struct {
    ID          shared.ID    `json:"id"`
    ProjectID   shared.ID    `json:"project_id"`
    Title       string       `json:"title"`
    Description string       `json:"description"`
    Steps       []StepDTO    `json:"steps"`
    Priority    int          `json:"priority"`
    Status      string       `json:"status"`
    Tags        []string     `json:"tags"`
    CreatedBy   shared.ID    `json:"created_by"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedBy   shared.ID    `json:"updated_by"`
    UpdatedAt   time.Time    `json:"updated_at"`
}
```

### 2.3 Infrastructure 层（基础设施层）

**职责**：技术实现细节。

**包含**：
- Persistence：数据库访问（Repository 实现）
- Cache：缓存
- AI：AI 客户端
- Storage：文件存储
- External：外部服务集成

**规则**：
- ✅ 实现 Domain 层定义的接口
- ✅ 依赖 Domain 层
- ❌ 不依赖 Application、Interface 层
- ❌ 不包含业务逻辑

**示例 - PostgreSQL Repository**：

```go
// internal/infrastructure/persistence/postgres/testcase.go
package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
    "heka-backend/internal/domain/shared"
    "heka-backend/internal/domain/testcase"
)

type testCaseRepository struct {
    db *sql.DB
}

func NewTestCaseRepository(db *sql.DB) testcase.Repository {
    return &testCaseRepository{db: db}
}

func (r *testCaseRepository) Save(ctx context.Context, tc *testcase.TestCase) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 序列化 Steps
    stepsJSON, err := json.Marshal(tc.Steps)
    if err != nil {
        return err
    }

    // 序列化 Tags
    tagsJSON, err := json.Marshal(tc.Tags)
    if err != nil {
        return err
    }

    query := `
        INSERT INTO test_cases (
            id, project_id, title, description, steps, priority, status,
            tags, created_by, created_at, updated_by, updated_at, version
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (id) DO UPDATE SET
            title = EXCLUDED.title,
            description = EXCLUDED.description,
            steps = EXCLUDED.steps,
            priority = EXCLUDED.priority,
            status = EXCLUDED.status,
            tags = EXCLUDED.tags,
            updated_by = EXCLUDED.updated_by,
            updated_at = EXCLUDED.updated_at,
            version = test_cases.version + 1
        WHERE test_cases.version = $14
    `

    result, err := tx.ExecContext(
        ctx,
        query,
        tc.ID,
        tc.ProjectID,
        tc.Title,
        tc.Description,
        stepsJSON,
        int(tc.Priority),
        string(tc.Status),
        tagsJSON,
        tc.CreatedBy,
        tc.CreatedAt,
        tc.UpdatedBy,
        tc.UpdatedAt,
        tc.Version,
        tc.Version, // 乐观锁
    )
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return testcase.ErrVersionConflict
    }

    return tx.Commit()
}

func (r *testCaseRepository) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
    var tc testcase.TestCase
    var stepsJSON, tagsJSON []byte

    query := `
        SELECT id, project_id, title, description, steps, priority, status,
               tags, created_by, created_at, updated_by, updated_at, version
        FROM test_cases
        WHERE id = $1
    `

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &tc.ID,
        &tc.ProjectID,
        &tc.Title,
        &tc.Description,
        &stepsJSON,
        &tc.Priority,
        &tc.Status,
        &tagsJSON,
        &tc.CreatedBy,
        &tc.CreatedAt,
        &tc.UpdatedBy,
        &tc.UpdatedAt,
        &tc.Version,
    )
    if err == sql.ErrNoRows {
        return nil, testcase.ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    // 反序列化
    if err := json.Unmarshal(stepsJSON, &tc.Steps); err != nil {
        return nil, err
    }
    if err := json.Unmarshal(tagsJSON, &tc.Tags); err != nil {
        return nil, err
    }

    return &tc, nil
}

func (r *testCaseRepository) FindByProject(
    ctx context.Context,
    projectID shared.ID,
    opts testcase.FindOptions,
) ([]*testcase.TestCase, int64, error) {
    // 构建查询
    query := `
        SELECT id, project_id, title, description, steps, priority, status,
               tags, created_by, created_at, updated_by, updated_at, version
        FROM test_cases
        WHERE project_id = $1
    `
    args := []interface{}{projectID}
    argIdx := 2

    // 动态条件
    if opts.Status != "" {
        query += fmt.Sprintf(" AND status = $%d", argIdx)
        args = append(args, opts.Status)
        argIdx++
    }
    if opts.Priority != testcase.Priority(0) {
        query += fmt.Sprintf(" AND priority = $%d", argIdx)
        args = append(args, int(opts.Priority))
        argIdx++
    }
    if opts.Keyword != "" {
        query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
        args = append(args, "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
        argIdx++
    }
    if len(opts.Tags) > 0 {
        query += fmt.Sprintf(" AND tags && $%d", argIdx)
        args = append(args, opts.Tags)
        argIdx++
    }

    // 排序
    if opts.SortBy == "" {
        opts.SortBy = "created_at"
    }
    order := "ASC"
    if opts.SortDesc {
        order = "DESC"
    }
    query += fmt.Sprintf(" ORDER BY %s %s", opts.SortBy, order)

    // 分页
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
    args = append(args, opts.PageSize, (opts.Page-1)*opts.PageSize)

    // 执行查询
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var testcases []*testcase.TestCase
    for rows.Next() {
        var tc testcase.TestCase
        var stepsJSON, tagsJSON []byte

        err := rows.Scan(
            &tc.ID,
            &tc.ProjectID,
            &tc.Title,
            &tc.Description,
            &stepsJSON,
            &tc.Priority,
            &tc.Status,
            &tagsJSON,
            &tc.CreatedBy,
            &tc.CreatedAt,
            &tc.UpdatedBy,
            &tc.UpdatedAt,
            &tc.Version,
        )
        if err != nil {
            return nil, 0, err
        }

        json.Unmarshal(stepsJSON, &tc.Steps)
        json.Unmarshal(tagsJSON, &tc.Tags)

        testcases = append(testcases, &tc)
    }

    // 查询总数
    countQuery := `
        SELECT COUNT(*)
        FROM test_cases
        WHERE project_id = $1
    ``
    // （省略 count 查询的条件构建）

    var total int64
    err = r.db.QueryRowContext(ctx, countQuery, projectID).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    return testcases, total, nil
}

func (r *testCaseRepository) Delete(ctx context.Context, id shared.ID) error {
    query := `DELETE FROM test_cases WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return testcase.ErrNotFound
    }

    return nil
}

func (r *testCaseRepository) UpdateStatus(ctx context.Context, id shared.ID, status testcase.Status) error {
    query := `
        UPDATE test_cases
        SET status = $1, updated_at = $2
        WHERE id = $3
    `
    result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return testcase.ErrNotFound
    }

    return nil
}
```

**示例 - Milvus Vector Repository**：

```go
// internal/infrastructure/persistence/milvus/vector.go
package milvus

import (
    "context"
    "heka-backend/internal/domain/file"
    "heka-backend/internal/domain/rag"
    "heka-backend/internal/domain/shared"
)

type vectorRepository struct {
    client *milvus.Client
}

func NewVectorRepository(client *milvus.Client) rag.VectorRepository {
    return &vectorRepository{client: client}
}

func (r *vectorRepository) Save(ctx context.Context, chunk *rag.Chunk) error {
    // 将 chunk 存入 Milvus
    collection := r.client.GetCollection("document_chunks")

    _, err := collection.Insert(
        context.Background(),
        []string{chunk.ID.String()},
        [][]float32{chunk.Embedding},
        []string{chunk.Content},
        []string{chunk.FileID.String()},
    )

    return err
}

func (r *vectorRepository) Search(ctx context.Context, embedding []float32, topK int) ([]*rag.Chunk, error) {
    collection := r.client.GetCollection("document_chunks")

    results, err := collection.Search(
        context.Background(),
        []string{embedding}, // query vectors
        "embedding",
        []string{"content", "file_id"},
        topK,
    )

    if err != nil {
        return nil, err
    }

    var chunks []*rag.Chunk
    for _, result := range results {
        chunks = append(chunks, &rag.Chunk{
            ID:      shared.ID(result.ID.(string)),
            Content: result.Fields["content"].(string),
            FileID:  shared.ID(result.Fields["file_id"].(string)),
        })
    }

    return chunks, nil
}

func (r *vectorRepository) FindByFileID(ctx context.Context, fileID shared.ID) ([]*rag.Chunk, error) {
    collection := r.client.GetCollection("document_chunks")

    results, err := collection.Query(
        context.Background(),
        fmt.Sprintf("file_id == %s", fileID),
        []string{"id", "content", "embedding"},
    )

    if err != nil {
        return nil, err
    }

    var chunks []*rag.Chunk
    for _, result := range results {
        chunks = append(chunks, &rag.Chunk{
            ID:        shared.ID(result.ID.(string)),
            Content:   result.Fields["content"].(string),
            Embedding: result.Fields["embedding"].([]float32),
            FileID:    fileID,
        })
    }

    return chunks, nil
}
```

**示例 - AI Client**：

```go
// internal/infrastructure/ai/client.go
package ai

import "context"

type Client interface {
    GenerateTestCases(ctx context.Context, prompt string) (*TestCasesResponse, error)
    AnalyzeChanges(ctx context.Context, diff string) (*AnalysisResponse, error)
    Chat(ctx context.Context, message string, history []Message) (string, error)
}

type TestCasesResponse struct {
    TestCases []TestCase `json:"test_cases"`
}

type TestCase struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Steps       []Step   `json:"steps"`
    Priority    int      `json:"priority"`
    Tags        []string `json:"tags"`
}

type Step struct {
    Action   string `json:"action"`
    Expected string `json:"expected"`
}

type AnalysisResponse struct {
    RecommendedTestCases []shared.ID `json:"recommended_test_cases"`
    Reasoning           string      `json:"reasoning"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

```go
// internal/infrastructure/ai/claude.go
package ai

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

type ClaudeClient struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewClaudeClient(apiKey string) *ClaudeClient {
    return &ClaudeClient{
        apiKey:  apiKey,
        baseURL: "https://api.anthropic.com/v1",
        client:  http.DefaultClient,
    }
}

func (c *ClaudeClient) GenerateTestCases(ctx context.Context, prompt string) (*TestCasesResponse, error) {
    reqBody := ClaudeRequest{
        Model:     "claude-3-5-sonnet-20241022",
        MaxTokens: 4096,
        Messages: []Message{
            {
                Role:    "user",
                Content: prompt,
            },
        },
    }

    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var claudeResp ClaudeResponse
    if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
        return nil, err
    }

    // 解析 AI 返回的测试用例
    return c.parseTestCasesResponse(claudeResp.Content[0].Text)
}

func (c *ClaudeClient) AnalyzeChanges(ctx context.Context, diff string) (*AnalysisResponse, error) {
    prompt := c.buildAnalysisPrompt(diff)
    // 类似实现
    return nil, nil
}

func (c *ClaudeClient) Chat(ctx context.Context, message string, history []Message) (string, error) {
    // 实现对话
    return "", nil
}

func (c *ClaudeClient) parseTestCasesResponse(text string) (*TestCasesResponse, error) {
    // 解析 AI 返回的 JSON
    return &TestCasesResponse{}, nil
}

func (c *ClaudeClient) buildAnalysisPrompt(diff string) string {
    return ""
}

type ClaudeRequest struct {
    Model     string   `json:"model"`
    MaxTokens int      `json:"max_tokens"`
    Messages  []Message `json:"messages"`
}

type ClaudeResponse struct {
    ID      string `json:"id"`
    Type    string `json:"type"`
    Content []struct {
        Type string `json:"type"`
        Text string `json:"text"`
    } `json:"content"`
}
```

**示例 - File Storage**：

```go
// internal/infrastructure/storage/local.go
package storage

import (
    "context"
    "io"
    "os"
    "path/filepath"
)

type LocalStorage struct {
    basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
    return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Save(ctx context.Context, path string, reader io.Reader) error {
    fullPath := filepath.Join(s.basePath, path)

    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return err
    }

    file, err := os.Create(fullPath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, reader)
    return err
}

func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
    fullPath := filepath.Join(s.basePath, path)
    return os.Open(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
    fullPath := filepath.Join(s.basePath, path)
    return os.Remove(fullPath)
}
```

**示例 - Figma Client**：

```go
// internal/infrastructure/figma/client.go
package figma

import (
    "context"
    "encoding/json"
    "net/http"
)

type Client struct {
    accessToken string
    baseURL     string
    client      *http.Client
}

func NewClient(accessToken string) *Client {
    return &Client{
        accessToken: accessToken,
        baseURL:     "https://api.figma.com/v1",
        client:      http.DefaultClient,
    }
}

type File struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Nodes []Node `json:"nodes"`
}

type Node struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Type     string `json:"type"`
    Children []Node `json:"children,omitempty"`
    Content  string `json:"content,omitempty"`
}

func (c *Client) GetFile(ctx context.Context, fileKey string) (*File, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/files/"+fileKey, nil)
    req.Header.Set("X-Figma-Token", c.accessToken)

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        File *File `json:"file"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.File, nil
}

func (c *Client) GetImage(ctx context.Context, fileKey string, nodeID string) ([]byte, error) {
    // 获取图片
    return nil, nil
}
```

### 2.4 Interface 层（接口层）

**职责**：对外暴露接口，处理请求/响应。

**包含**：
- HTTP Handler：处理 HTTP 请求
- Middleware：横切关注点
- Router：路由注册

**规则**：
- ✅ 依赖 Application 层
- ❌ 不直接调用 Domain、Infrastructure 层
- ❌ 不包含业务逻辑
- ✅ 只负责参数解析、调用服务、返回响应

**示例 - HTTP Handler**：

```go
// internal/interface/http/handler/testcase.go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    "heka-backend/internal/application/testcase"
    "heka-backend/internal/domain/shared"
    "heka-backend/internal/interface/http/response"
)

type TestCaseHandler struct {
    service *testcase.Service
}

func NewTestCaseHandler(service *testcase.Service) *TestCaseHandler {
    return &TestCaseHandler{service: service}
}

// CreateTestCase 创建测试用例
// POST /api/testcases
func (h *TestCaseHandler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
    var req testcase.CreateTestCaseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // 从上下文获取用户 ID（由中间件设置）
    userID := r.Context().Value("user_id").(shared.ID)
    req.CreatorID = userID

    res, err := h.service.CreateTestCase(r.Context(), req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusCreated, res)
}

// ListTestCases 查询测试用例列表
// GET /api/testcases
func (h *TestCaseHandler) ListTestCases(w http.ResponseWriter, r *http.Request) {
    req := testcase.ListTestCasesRequest{
        ProjectID: shared.ID(r.URL.Query().Get("project_id")),
    }

    // 解析分页参数
    if page := r.URL.Query().Get("page"); page != "" {
        if p, err := strconv.Atoi(page); err == nil {
            req.Page = p
        }
    }
    if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
        if ps, err := strconv.Atoi(pageSize); err == nil {
            req.PageSize = ps
        }
    }
    if req.Page == 0 {
        req.Page = 1
    }
    if req.PageSize == 0 {
        req.PageSize = 20
    }

    // 其他过滤参数
    req.Status = r.URL.Query().Get("status")
    req.Priority = r.URL.Query().Get("priority")
    req.Keyword = r.URL.Query().Get("keyword")
    req.SortBy = r.URL.Query().Get("sort_by")
    req.SortDesc = r.URL.Query().Get("sort_desc") == "true"

    res, err := h.service.ListTestCases(r.Context(), req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, res)
}

// GetTestCase 获取单个测试用例
// GET /api/testcases/:id
func (h *TestCaseHandler) GetTestCase(w http.ResponseWriter, r *http.Request) {
    id := shared.ID(r.PathValue("id"))

    tc, err := h.service.GetTestCase(r.Context(), id)
    if err != nil {
        if err == testcase.ErrNotFound {
            response.Error(w, http.StatusNotFound, "test case not found")
            return
        }
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, tc)
}

// UpdateTestCase 更新测试用例
// PUT /api/testcases/:id
func (h *TestCaseHandler) UpdateTestCase(w http.ResponseWriter, r *http.Request) {
    id := shared.ID(r.PathValue("id"))

    var req testcase.UpdateTestCaseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    userID := r.Context().Value("user_id").(shared.ID)
    req.UpdaterID = userID

    res, err := h.service.UpdateTestCase(r.Context(), id, req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, res)
}

// DeleteTestCase 删除测试用例
// DELETE /api/testcases/:id
func (h *TestCaseHandler) DeleteTestCase(w http.ResponseWriter, r *http.Request) {
    id := shared.ID(r.PathValue("id"))

    if err := h.service.DeleteTestCase(r.Context(), id); err != nil {
        if err == testcase.ErrNotFound {
            response.Error(w, http.StatusNotFound, "test case not found")
            return
        }
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.NoContent(w)
}

// GenerateByAI AI 生成测试用例
// POST /api/testcases/ai-generate
func (h *TestCaseHandler) GenerateByAI(w http.ResponseWriter, r *http.Request) {
    var req testcase.GenerateByAIRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    userID := r.Context().Value("user_id").(shared.ID)
    req.CreatorID = userID

    res, err := h.service.GenerateByAI(r.Context(), req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, res)
}
```

```go
// internal/interface/http/handler/ai.go
package handler

import (
    "encoding/json"
    "net/http"
    "heka-backend/internal/application/ai"
    "heka-backend/internal/domain/shared"
    "heka-backend/internal/interface/http/response"
)

type AIHandler struct {
    service *ai.Service
}

func NewAIHandler(service *ai.Service) *AIHandler {
    return &AIHandler{service: service}
}

// Chat AI 助手对话
// POST /api/ai/chat
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
    var req ai.ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    userID := r.Context().Value("user_id").(shared.ID)

    res, err := h.service.Chat(r.Context(), userID, req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, res)
}

// Analyze 分析代码变更，推荐回归测试
// POST /api/ai/analyze
func (h *AIHandler) Analyze(w http.ResponseWriter, r *http.Request) {
    var req ai.AnalyzeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    res, err := h.service.Analyze(r.Context(), req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, res)
}
```

**示例 - Response**：

```go
// internal/interface/http/response/response.go
package response

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{
        Code: 0,
        Data: data,
    })
}

func Error(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{
        Code:    status,
        Message: message,
    })
}

func NoContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}
```

**示例 - Middleware**：

```go
// internal/interface/http/middleware/auth.go
package middleware

import (
    "net/http"
    "heka-backend/internal/domain/shared"
)

func Auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从 Header 获取 Token
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        // 验证 Token
        userID, err := validateToken(token)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        // 将用户 ID 放入上下文
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

```go
// internal/interface/http/middleware/cors.go
package middleware

import (
    "net/http"
)

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**示例 - Router**：

```go
// internal/interface/http/router.go
package http

import (
    "net/http"
    "heka-backend/internal/application/ai"
    "heka-backend/internal/application/testcase"
    "heka-backend/internal/interface/http/handler"
    "heka-backend/internal/interface/http/middleware"
)

type Router struct {
    tcHandler *handler.TestCaseHandler
    aiHandler *handler.AIHandler
}

func NewRouter(
    tcSvc *testcase.Service,
    aiSvc *ai.Service,
) *Router {
    return &Router{
        tcHandler: handler.NewTestCaseHandler(tcSvc),
        aiHandler: handler.NewAIHandler(aiSvc),
    }
}

func (r *Router) Setup() http.Handler {
    mux := http.NewServeMux()

    // 测试用例路由
    mux.HandleFunc("POST /api/testcases", r.tcHandler.CreateTestCase)
    mux.HandleFunc("GET /api/testcases", r.tcHandler.ListTestCases)
    mux.HandleFunc("GET /api/testcases/{id}", r.tcHandler.GetTestCase)
    mux.HandleFunc("PUT /api/testcases/{id}", r.tcHandler.UpdateTestCase)
    mux.HandleFunc("DELETE /api/testcases/{id}", r.tcHandler.DeleteTestCase)
    mux.HandleFunc("POST /api/testcases/ai-generate", r.tcHandler.GenerateByAI)

    // AI 路由
    mux.HandleFunc("POST /api/ai/chat", r.aiHandler.Chat)
    mux.HandleFunc("POST /api/ai/analyze", r.aiHandler.Analyze)

    // 应用中间件
    var handler http.Handler = mux
    handler = middleware.Auth(handler)
    handler = middleware.CORS(handler)
    handler = middleware.Logger(handler)

    return handler
}
```

## 3. 跨模块调用规则

### 3.1 依赖方向

```
Interface → Application → Domain ← Infrastructure
```

**依赖规则**：
- Interface 层 → Application 层
- Application 层 → Domain 层
- Infrastructure 层 → Domain 层（实现接口）
- Domain 层 → 无依赖（除了 shared）

### 3.2 跨模块调用示例

**场景**：AI 生成测试用例需要用到文件和 RAG

```go
// internal/application/ai/service.go
package ai

import (
    "context"
    "heka-backend/internal/application/file"
    "heka-backend/internal/application/rag"
    "heka-backend/internal/application/testcase"
    "heka-backend/internal/infrastructure/ai"
)

type Service struct {
    fileSvc *file.Service
    ragSvc  *rag.Service
    tcSvc   *testcase.Service
    aiClient ai.Client
}

func NewService(
    fileSvc *file.Service,
    ragSvc *rag.Service,
    tcSvc *testcase.Service,
    aiClient ai.Client,
) *Service {
    return &Service{
        fileSvc:  fileSvc,
        ragSvc:   ragSvc,
        tcSvc:    tcSvc,
        aiClient: aiClient,
    }
}

// GenerateTestCases AI 生成测试用例
func (s *Service) GenerateTestCases(ctx context.Context, fileID shared.ID, query string) ([]*testcase.TestCase, error) {
    // 1. 调用文件服务获取文件
    f, err := s.fileSvc.GetFile(ctx, fileID)
    if err != nil {
        return nil, err
    }

    // 2. 调用 RAG 服务检索相关内容
    ragResult, err := s.ragSvc.Retrieve(ctx, f, query)
    if err != nil {
        return nil, err
    }

    // 3. 调用 AI 客户端生成用例
    prompt := s.buildPrompt(ragResult)
    aiResponse, err := s.aiClient.GenerateTestCases(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 4. 调用用例服务创建用例
    return s.tcSvc.CreateFromAI(ctx, aiResponse)
}
```

**关键**：
- Application 层之间可以相互调用
- 通过构造函数注入依赖
- 不跨层调用（Interface 不能直接调 Infrastructure）

### 3.3 禁止的调用

```go
// ❌ 错误：Handler 直接调 Repository
func (h *Handler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
    // 不能直接调 Infrastructure
    // h.tcRepo.Save(...)
}

// ❌ 错误：Domain 层调 Infrastructure
func (tc *TestCase) Save() error {
    // Domain 不能依赖具体技术实现
}

// ❌ 错误：Application 层直接调 HTTP
func (s *Service) CreateTestCase() {
    // 应该通过 Infrastructure 的 AI Client
    // http.Get(...)
}

// ❌ 错误：Application 层调 Interface 层
func (s *Service) CreateTestCase() {
    // 不能反向依赖
    // handler.SendNotification(...)
}
```

## 3.5 Context 传递规范

### 3.5.1 标准 Context Key

```go
// internal/shared/context/keys.go
package context

type contextKey string

const (
    KeyUserID    contextKey = "user_id"
    KeyProjectID contextKey = "project_id"
    KeyRequestID contextKey = "request_id"
)

// Helper 函数
func WithUserID(ctx context.Context, userID shared.ID) context.Context {
    return context.WithValue(ctx, KeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (shared.ID, bool) {
    id, ok := ctx.Value(KeyUserID).(shared.ID)
    return id, ok
}

func WithProjectID(ctx context.Context, projectID shared.ID) context.Context {
    return context.WithValue(ctx, KeyProjectID, projectID)
}

func ProjectIDFromContext(ctx context.Context) (shared.ID, bool) {
    id, ok := ctx.Value(KeyProjectID).(shared.ID)
    return id, ok
}
```

### 3.5.2 Middleware 注入

```go
// internal/interface/http/middleware/auth.go
// Auth 中间件注入用户 ID
func Auth(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            userID, err := validateToken(token, jwtSecret)
            if err != nil {
                response.Error(w, http.StatusUnauthorized, "AUTH-AU-001")
                return
            }
            // 注入用户 ID 到 Context
            ctx := sharedctx.WithUserID(r.Context(), userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Project 中间件注入项目 ID（从 URL 或 Header 提取）
func Project() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            projectID := shared.ID(r.PathValue("project_id"))
            if projectID == "" {
                projectID = shared.ID(r.URL.Query().Get("project_id"))
            }
            if projectID != "" {
                ctx := sharedctx.WithProjectID(r.Context(), projectID)
                r = r.WithContext(ctx)
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 3.5.3 Service 层使用

```go
// Application 层从 Context 获取用户 ID
func (s *Service) CreateTestCase(ctx context.Context, req dto.CreateTestCaseRequest) (*dto.CreateTestCaseResponse, error) {
    userID, ok := sharedctx.UserIDFromContext(ctx)
    if !ok {
        return nil, shared.ErrUnauthorized
    }
    // 使用 userID 作为创建者
    tc := &testcase.TestCase{
        CreatedBy: userID,
        // ...
    }
    // ...
}
```

**规则**：
- ✅ 只使用 `sharedctx.UserIDFromContext(ctx)` 获取值
- ❌ 禁止 `ctx.Value("user_id")` 字符串 key
- ✅ Context key 定义集中在 `shared/context/keys.go`
- ❌ 禁止在 Domain 层读取 Context 业务值

## 3.6 事务管理规范

### 3.6.1 事务管理原则

- **Application 层管理事务边界**
- Domain 层不感知事务
- Infrastructure 层通过 TransactionManager 接口参与事务

### 3.6.2 TransactionManager 接口

```go
// internal/domain/shared/transaction.go
package shared

import "context"

// TransactionManager 事务管理器接口
type TransactionManager interface {
    // InTx 在事务中执行 fn，自动提交或回滚
    InTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 3.6.3 Infrastructure 层实现

```go
// internal/infrastructure/persistence/postgres/transaction.go
package postgres

import (
    "context"
    "database/sql"
    "heka-backend/internal/domain/shared"
)

type transactionManager struct {
    db *sql.DB
}

func NewTransactionManager(db *sql.DB) shared.TransactionManager {
    return &transactionManager{db: db}
}

func (tm *transactionManager) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := tm.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 将 tx 注入到 context
    txCtx := context.WithValue(ctx, txKey{}, tx)

    if err := fn(txCtx); err != nil {
        return err
    }

    return tx.Commit()
}

// DBOrTx 从 context 获取事务，如果没有事务则返回 db
func DBOrTx(ctx context.Context, db *sql.DB) DBTX {
    if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
        return tx
    }
    return db
}

type txKey struct{}

type DBTX interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
```

### 3.6.4 Application 层使用

```go
// Application 层管理事务边界
func (s *Service) CreateTestPlan(ctx context.Context, req dto.CreateTestPlanRequest) (*dto.CreateTestPlanResponse, error) {
    var plan *plan.TestPlan

    err := s.txManager.InTx(ctx, func(txCtx context.Context) error {
        // 1. 创建计划
        p := &plan.TestPlan{...}
        if err := s.planRepo.Save(txCtx, p); err != nil {
            return err
        }

        // 2. 关联用例（在同一事务中）
        for _, tcID := range req.TestCaseIDs {
            if err := s.planRepo.AddTestCase(txCtx, p.ID, tcID); err != nil {
                return err
            }
        }

        plan = p
        return nil
    })

    if err != nil {
        return nil, err
    }

    return &dto.CreateTestPlanResponse{ID: plan.ID}, nil
}
```

### 3.6.5 Repository 使用 DBOrTx

```go
// Repository 实现使用 DBOrTx 确保在事务内
func (r *testCaseRepository) Save(ctx context.Context, tc *testcase.TestCase) error {
    db := DBOrTx(ctx, r.db) // 自动使用事务或普通连接
    _, err := db.ExecContext(ctx, "INSERT INTO ...", ...)
    return err
}
```

**规则**：
- ✅ Application 层通过 `txManager.InTx()` 管理事务边界
- ✅ Repository 使用 `DBOrTx(ctx, r.db)` 自动判断是否在事务中
- ❌ 禁止 Repository 自行管理事务（Begin/Commit）
- ❌ 禁止跨服务调用的长事务

## 3.7 跨模块事件机制

### 3.7.1 Domain Event 接口

```go
// internal/domain/shared/event.go
package shared

import "context"

// Event 领域事件接口
type Event interface {
    EventName() string
    OccurredAt() time.Time
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event Event) error

// EventBus 事件总线接口
type EventBus interface {
    Publish(ctx context.Context, events ...Event) error
    Subscribe(eventName string, handler EventHandler)
}
```

### 3.7.2 异步 EventBus 实现

```go
// internal/infrastructure/eventbus/async.go
package eventbus

type asyncEventBus struct {
    handlers map[string][]shared.EventHandler
    workers  int
    queue    chan eventJob
}

type eventJob struct {
    ctx   context.Context
    event shared.Event
}

func NewAsyncEventBus(workers int) *asyncEventBus {
    b := &asyncEventBus{
        handlers: make(map[string][]shared.EventHandler),
        workers:  workers,
        queue:    make(chan eventJob, 100),
    }
    for i := 0; i < workers; i++ {
        go b.worker()
    }
    return b
}

func (b *asyncEventBus) worker() {
    for job := range b.queue {
        handlers := b.handlers[job.event.EventName()]
        for _, h := range handlers {
            h(job.ctx, job.event) // 忽略错误，记录日志
        }
    }
}

func (b *asyncEventBus) Publish(ctx context.Context, events ...shared.Event) error {
    for _, e := range events {
        b.queue <- eventJob{ctx: ctx, event: e}
    }
    return nil
}

func (b *asyncEventBus) Subscribe(eventName string, handler shared.EventHandler) {
    b.handlers[eventName] = append(b.handlers[eventName], handler)
}
```

### 3.7.3 事件定义和使用

```go
// internal/domain/testcase/events.go
package testcase

type TestCaseCreatedEvent struct {
    ID        shared.ID
    ProjectID shared.ID
    Title     string
    Timestamp time.Time
}

func (e TestCaseCreatedEvent) EventName() string    { return "testcase.created" }
func (e TestCaseCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

// Application 层发布事件
func (s *Service) CreateTestCase(ctx context.Context, req dto.CreateTestCaseRequest) (*dto.CreateTestCaseResponse, error) {
    // ... 创建用例 ...

    s.eventBus.Publish(ctx, testcase.TestCaseCreatedEvent{
        ID:        tc.ID,
        ProjectID: tc.ProjectID,
        Title:     tc.Title,
        Timestamp: time.Now(),
    })

    return response, nil
}

// 订阅：用例创建后清除缓存
eventBus.Subscribe("testcase.created", func(ctx context.Context, e shared.Event) error {
    evt := e.(testcase.TestCaseCreatedEvent)
    cache.Delete(ctx, fmt.Sprintf("project:%s:testcases", evt.ProjectID))
    return nil
})
```

**使用场景**：
- 用例创建/更新 → 清除缓存
- 文件上传完成 → 触发 RAG 索引
- AI 生成完成 → 通知前端（SSE）

## 4. 模块划分

根据第一阶段功能，划分以下领域模块：

| 模块 | 职责 | 领域对象 |
|------|------|----------|
| `testcase` | 测试用例管理 | TestCase, Step, Priority, Status |
| `plan` | 测试计划管理 | TestPlan, PlanIteration, PlanStatus |
| `execution` | 执行记录管理 | ExecutionRecord, ExecutionResult, BugReference |
| `file` | 文件管理 | File, FileVersion, FileType |
| `rag` | RAG 向量检索 | Document, Chunk, Embedding |
| `user` | 用户管理 | User, Session |
| `project` | 项目管理 | Project, ProjectMember |
| `shared` | 共享概念 | ID, Timestamp, Errors |

## 5. 依赖注入示例

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "heka-backend/internal/application/ai"
    "heka-backend/internal/application/file"
    "heka-backend/internal/application/project"
    "heka-backend/internal/application/rag"
    "heka-backend/internal/application/testcase"
    "heka-backend/internal/application/user"
    "heka-backend/internal/infrastructure/ai"
    infrafile "heka-backend/internal/infrastructure/file"
    "heka-backend/internal/infrastructure/persistence/milvus"
    "heka-backend/internal/infrastructure/persistence/postgres"
    "heka-backend/internal/infrastructure/storage"
    "heka-backend/internal/interface/http"
)

func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化基础设施
    db := postgres.NewDB(cfg.Database.DSN)
    redisClient := redis.NewClient(cfg.Redis.URL)
    milvusClient := milvus.NewClient(cfg.Milvus.Address)
    localStorage := storage.NewLocalStorage(cfg.Storage.Path)

    // 3. 初始化 AI 客户端
    var aiClient ai.Client
    switch cfg.AI.Provider {
    case "claude":
        aiClient = ai.NewClaudeClient(cfg.AI.ApiKey)
    case "openai":
        aiClient = ai.NewOpenAIClient(cfg.AI.ApiKey)
    case "local":
        aiClient = ai.NewLocalClient(cfg.AI.Endpoint)
    }

    // 4. 初始化仓储（Domain 层接口的实现）
    tcRepo := postgres.NewTestCaseRepository(db)
    planRepo := postgres.NewTestPlanRepository(db)
    executionRepo := postgres.NewExecutionRepository(db)
    fileRepo := postgres.NewFileRepository(db)
    ragRepo := milvus.NewVectorRepository(milvusClient)
    userRepo := postgres.NewUserRepository(db)
    projectRepo := postgres.NewProjectRepository(db)

    // 5. 初始化应用服务
    userSvc := user.NewService(userRepo)
    projectSvc := project.NewService(projectRepo)

    fileSvc := file.NewService(
        fileRepo,
        localStorage,
        infrafile.NewPDFParser(),
        infrafile.NewWordParser(),
        infrafile.NewExcelParser(),
        infrafile.NewImageParser(),
        infrafile.NewFigmaClient(cfg.Figma.AccessToken),
    )

    ragSvc := rag.NewService(ragRepo, aiClient)

    tcSvc := testcase.NewService(tcRepo, ragSvc)

    aiSvc := ai.NewService(
        fileSvc,
        ragSvc,
        tcSvc,
        aiClient,
    )

    // 6. 初始化 HTTP 接口
    router := http.NewRouter(
        tcSvc,
        aiSvc,
        userSvc,
        projectSvc,
    )

    // 7. 启动服务
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router.Setup(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    log.Printf("Server starting on :8080")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

## 6. 命名约定

### 6.1 文件命名

- `entity.go` - 实体定义
- `valueobject.go` - 值对象
- `repository.go` - 仓储接口（Domain 层）
- `service.go` - 领域服务（Domain 层）或应用服务（Application 层）
- `dto.go` - 数据传输对象
- `handler.go` - HTTP 处理器

### 6.2 包命名

- Domain 层：`internal/domain/<module>`
- Application 层：`internal/application/<module>`
- Infrastructure 层：`internal/infrastructure/<category>/<tech>`
- Interface 层：`internal/interface/http/<category>`

### 6.3 接口命名

- Repository 接口放在 Domain 层
- 实现放在 Infrastructure 层
- 命名：`<Name>Repository` 接口，`<tech><Name>Repository` 实现

```go
// Domain
type TestCaseRepository interface { ... }

// Infrastructure
type testCaseRepository struct { ... }
func NewTestCaseRepository(db *sql.DB) testcase.TestCaseRepository { ... }
```

### 6.4 变量命名

- 应用服务：`svc`
- 仓储：`repo`
- Handler：`h`
- 上下文：`ctx`
- 请求：`req`
- 响应：`res` / `resp`

## 7. 错误处理

```go
// internal/domain/shared/errors.go
package shared

import "errors"

var (
    ErrNotFound      = errors.New("not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("conflict")
)

// 包装错误
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
```

```go
// 使用示例
func (s *Service) GetTestCase(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
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

## 7.5 测试文件组织

### 7.5.1 测试目录结构

```
internal/
├── domain/
│   └── testcase/
│       ├── entity.go
│       ├── entity_test.go           # 单元测试：实体行为
│       ├── valueobject.go
│       └── valueobject_test.go      # 单元测试：值对象
├── application/
│   └── testcase/
│       ├── service.go
│       └── service_test.go          # 单元测试：应用服务（Mock Repository）
├── infrastructure/
│   └── persistence/
│       └── postgres/
│           └── testcase_test.go     # 集成测试（需真实数据库）
└── interface/
    └── http/
        └── handler/
            └── testcase_test.go     # 集成测试（HTTP 端点）
```

### 7.5.2 测试分层要求

| 层 | 测试类型 | 覆盖率 | 框架 | 说明 |
|----|---------|--------|------|------|
| Domain | 单元测试 | >= 80% | testify | 纯逻辑，无外部依赖 |
| Application | 单元测试 | >= 80% | testify + mock | Mock Repository |
| Infrastructure | 集成测试 | >= 60% | testify + testcontainers | 需要真实数据库 |
| Interface | API 测试 | >= 60% | httptest | HTTP 端点测试 |

### 7.5.3 单元测试示例

```go
// internal/application/testcase/service_test.go
package testcase

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock Repository
type MockTestCaseRepository struct {
    mock.Mock
}

func (m *MockTestCaseRepository) Save(ctx context.Context, tc *testcase.TestCase) error {
    args := m.Called(ctx, tc)
    return args.Error(0)
}

func (m *MockTestCaseRepository) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*testcase.TestCase), args.Error(1)
}

func TestService_CreateTestCase(t *testing.T) {
    // Arrange
    mockRepo := new(MockTestCaseRepository)
    svc := NewService(mockRepo, nil, nil, nil)
    ctx := context.Background()

    mockRepo.On("Save", ctx, mock.AnythingOfType("*testcase.TestCase")).Return(nil)

    // Act
    req := dto.CreateTestCaseRequest{
        ProjectID: "project-uuid",
        Title:     "测试标题",
        Steps:     []dto.StepDTO{{Action: "步骤1", Expected: "预期1"}},
        Priority:  1,
    }
    res, err := svc.CreateTestCase(ctx, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, res)
    mockRepo.AssertCalled(t, "Save", ctx, mock.Anything)
}
```

## 8. Standard Go Project Layout 可选目录详解

### 8.1 `/api` - API 协议文件

**用途**：存放 API 定义文件，便于前后端协作和自动生成客户端代码。

**推荐结构**：

```
api/
├── openapi/
│   ├── openapi.yaml              # 主 API 规范文件
│   └── schemas/                  # JSON Schema 定义
│       ├── testcase.schema.json
│       ├── plan.schema.json
│       └── execution.schema.json
└── protobuf/                     # 如果使用 gRPC
    └── heka.proto
```

**OpenAPI 规范示例**：

```yaml
# api/openapi/openapi.yaml
openapi: 3.0.3
info:
  title: Heka AI 测试管理平台 API
  version: 1.0.0
  description: 简化版 AI 测试管理平台 API

servers:
  - url: http://localhost:8080/api/v1
    description: 本地开发环境

paths:
  /testcases:
    get:
      summary: 获取测试用例列表
      parameters:
        - name: project_id
          in: query
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: 成功
    post:
      summary: 创建测试用例
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTestCaseRequest'
```

### 8.2 `/configs` - 配置文件模板

**用途**：提供配置文件示例和不同环境的配置模板。

**推荐结构**：

```
configs/
├── config.example.yaml           # 配置示例（所有字段注释）
├── config.dev.yaml               # 开发环境配置
├── config.test.yaml              # 测试环境配置
├── config.prod.yaml              # 生产环境配置
└── docker-compose.example.yaml   # Docker Compose 示例
```

**配置示例**：

```yaml
# configs/config.example.yaml
server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

database:
  host: localhost
  port: 5432
  database: heka
  # 使用环境变量：${DB_PASSWORD}
  password: ${DB_PASSWORD}

redis:
  addr: localhost:6379
  password: ${REDIS_PASSWORD}
  db: 0

milvus:
  address: localhost:19530

ai:
  # Provider: claude, openai, gemini, ollama
  provider: claude
  api_key: ${AI_API_KEY}
  timeout: 60s
  max_retries: 3
```

### 8.3 `/test` - 额外测试数据

**用途**：存放测试 fixtures、测试数据文件和测试辅助工具（与 `_test.go` 文件分离）。

**推荐结构**：

```
test/
├── fixtures/                     # 测试 fixtures
│   ├── testcases.json            # 测试用例 fixtures
│   ├── plans.json                # 测试计划 fixtures
│   └── users.json                # 用户 fixtures
├── testdata/                     # 测试数据文件
│   ├── files/                    # 测试用文件
│   │   ├── sample.pdf
│   │   ├── sample.docx
│   │   └── sample.xlsx
│   └── requirements/             # 测试用需求文档
│       └── login-feature.md
├── integration/                  # 集成测试辅助
│   ├── setup.go                  # 测试环境设置
│   └── teardown.go               # 测试环境清理
└── mocks/                        # Mock 生成器配置
    └── mockgen.config.yaml
```

**Fixtures 示例**：

```json
// test/fixtures/testcases.json
{
  "testcases": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "project_id": "550e8400-e29b-41d4-a716-446655440001",
      "title": "用户登录测试",
      "description": "验证用户登录功能",
      "priority": 1,
      "status": "ready",
      "steps": [
        {
          "number": 1,
          "action": "打开登录页面",
          "expected": "显示登录表单"
        }
      ]
    }
  ]
}
```

### 8.4 `/docs` - 项目文档

**用途**：存放项目文档、设计文档、架构图、API 文档等。
*注：Heka 使用 `/specs` 目录存放设计文档，两者功能类似。*

**推荐结构**：

```
docs/
├── architecture/                 # 架构文档
│   ├── overview.md              # 系统概述
│   ├── clean-architecture.md    # 分层架构说明
│   └── deployment.md            # 部署架构
├── api/                         # API 文档
│   ├── authentication.md        # 认证说明
│   ├── endpoints.md             # 端点列表
│   └── errors.md                # 错误码说明
├── development/                 # 开发指南
│   ├── getting-started.md       # 快速开始
│   ├── contributing.md          # 贡献指南
│   └── coding-standards.md      # 编码规范
└── diagrams/                    # 架构图
    ├── architecture.png
    └── data-flow.png
```

### 8.5 使用建议

| 目录 | 优先级 | 建议补充时机 |
|------|--------|-------------|
| `/api` | 🔵 推荐 | 项目启动时即添加，便于前后端协作 |
| `/configs` | 🔵 推荐 | 项目启动时添加，提供配置示例 |
| `/test` | 🟢 可选 | 测试数据变多时添加，保持项目整洁 |
| `/docs` | 🟢 可选 | Heka 已有 `/specs` 目录，无需额外创建 |

### 8.6 与现有目录的关系

Heka 项目使用 `/specs` 目录存放设计文档，这与 Standard Go Layout 的 `/docs` 功能类似：

```
heka-backend/
├── specs/                        # Heka 的设计文档目录
│   ├── heka-design-doc.md
│   ├── backend-code-organization.md
│   ├── database-performance-spec.md
│   └── ...
└── docs/                         # 可选：用户文档、开发指南
    ├── getting-started.md
    └── api-reference.md
```

**建议**：
- `/specs`：继续用于存放技术设计文档、架构规范、性能分析等
- `/docs`（可选）：用于存放面向用户的文档、快速开始指南等
