# Heka 后端原子化任务列表

> 基于 `spec.md` v1.2 + `plan.md` v1.1，按分层架构拆解
> 版本：v1.1（review 修复）
> 日期：2026-05-15

---

## 约定

- **[P]** = 可与同组其他 [P] 任务并行执行
- **TDD RED** = 先写测试，预期测试失败（运行时失败或功能未实现）
- **TDD GREEN** = 写实现使测试通过
- **Depends** = 前置任务 ID
- 每个任务只涉及**一个主要文件**的创建或修改

---

## Phase 1: Foundation（数据结构定义）

> 纯类型、接口、DDL 定义。无业务逻辑，无实现代码。

---

### 1.1 项目骨架 + 基础设施

#### T001: 初始化 Go Module 和目录结构
- **File:** `go.mod`, `cmd/server/main.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `go mod init github.com/heka/heka`
  - 创建完整目录结构（参照 CLAUDE.md 目录规范）
  - `cmd/server/main.go` 仅 `func main() { log.Println("heka: starting...") }`
  - 确保 `go build ./...` 通过

#### T002: [P] 配置结构体定义（仅类型）
- **File:** `internal/shared/config/config.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - 定义 `Config` 结构体，嵌套 `DatabaseConfig`、`RedisConfig`、`MilvusConfig`、`JWTConfig`、`ServerConfig`、`AIConfig`、`UploadConfig`
  - 字段映射 spec.md 第 8 节全部环境变量（HEKA_DB_HOST 等）
  - 包含 `mapstructure` tag
  - **不包含** `Load()` 函数（留给 Phase 2 T040a）

#### T003: [P] 日志初始化
- **File:** `internal/shared/logger/logger.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `Init(level string) *zap.Logger` 初始化 Zap logger
  - 支持 debug/info/warn/error 级别
  - 全局 `L()` 快捷函数获取 logger

#### T004: [P] Context 工具
- **File:** `internal/shared/context/context.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type contextKey struct{ name string }` 私有 typed key
  - `WithUserID(ctx, id shared.ID) context.Context`
  - `UserIDFromContext(ctx) (shared.ID, bool)`
  - `WithProjectID(ctx, id shared.ID) context.Context`
  - `ProjectIDFromContext(ctx) (shared.ID, bool)`

#### T005: [P] 请求校验器
- **File:** `internal/shared/validator/validator.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - 封装 `github.com/go-playground/validator/v10`
  - `New()` 返回已注册自定义规则的 validator
  - 注册自定义规则：`uuid`（调用 uuid.Parse）
  - `ValidateStruct(s interface{}) error` 返回格式化错误

#### T005a: Docker Compose 开发环境（PostgreSQL + Redis）
- **File:** `docker-compose.yml`
- **Depends:** —
- **Action:** Create
- **Details:**
  - PostgreSQL 15 + Redis 7，健康检查，volume 持久化，端口映射
  - Phase 2 Repository TDD 测试需要 PostgreSQL 运行
  - Milvus 后续通过 `docker-compose.override.yml` 追加

---

### 1.2 Shared Domain 类型

#### T006: [P] ID 类型
- **File:** `internal/domain/shared/types.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type ID string`
  - `func NewID() ID` — 生成 UUID v4
  - `func (id ID) String() string`
  - `func (id ID) IsEmpty() bool`
  - `func ParseID(s string) (ID, error)` — 校验 UUID 格式

#### T007: [P] 统一错误码
- **File:** `internal/domain/shared/errors.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type AppError struct { Code string; Message string; HTTPStatus int }`
  - `func (e *AppError) Error() string`
  - `func NewAppError(code, message string, httpStatus int) *AppError`
  - 定义 spec.md 第 9 节全部错误码常量（AUTH-AU-001 到 SYS-RT-001，共 ~30 个）
  - 按 Domain 分组：`authErrors`, `tcErrors`, `tpErrors`, `exErrors`, `fileErrors`, `aiErrors`, `projErrors`, `userErrors`, `ragErrors`, `sysErrors`

#### T008: [P] TransactionManager 接口
- **File:** `internal/domain/shared/transaction.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type TransactionManager interface { InTx(ctx context.Context, fn func(ctx context.Context) error) error }`

#### T009: [P] EventBus 接口
- **File:** `internal/domain/shared/event.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type Event interface { EventName() string; OccurredAt() time.Time }`
  - `type EventHandler func(ctx context.Context, event Event) error`
  - `type EventBus interface { Publish(ctx context.Context, events ...Event) error; Subscribe(eventName string, handler EventHandler) }`

---

### 1.3 Domain 实体定义

#### T010: [P] User Entity（仅结构体，不含业务逻辑）
- **File:** `internal/domain/user/entity.go`
- **Depends:** T006, T007
- **Action:** Create
- **Details:**
  - `type User struct { ID shared.ID; Name string; Email string; PasswordHash string; CreatedAt time.Time; UpdatedAt time.Time; DeletedAt *time.Time }`
  - 仅定义结构体。密码哈希/校验逻辑留给 Phase 3 UserService

#### T011: [P] Project + ProjectMember Entity
- **File:** `internal/domain/project/entity.go`
- **Depends:** T006, T007
- **Action:** Create
- **Details:**
  - `type Project struct { ID shared.ID; Name string; Description string; CreatedBy shared.ID; CreatedAt time.Time; UpdatedAt time.Time; DeletedAt *time.Time }`
  - `type ProjectMember struct { ProjectID shared.ID; UserID shared.ID; JoinedAt time.Time }`

#### T012: [P] TestCase 值对象
- **File:** `internal/domain/testcase/valueobject.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type Priority int` — 常量 P0=0, P1=1, P2=2, P3=3
  - `type CaseStatus string` — 常量 CaseDraft, CaseReady, CaseArchived
  - `func ValidateCaseTransition(from, to CaseStatus) error` — 实现 spec 3.2 状态转换规则：draft→ready/archived, ready→archived/draft, archived→ready
  - `func (p Priority) Valid() bool` — 0 <= p <= 3

#### T013: [P] TestCase Domain 实体（含 Module, Tag, Collection, Step）
- **File:** `internal/domain/testcase/entity.go`
- **Depends:** T006, T007, T012
- **Action:** Create
- **Details:**
  - `type Module struct { ID shared.ID; ProjectID shared.ID; Name string; Description string; ParentID *shared.ID; OrderIndex int; CreatedBy shared.ID }`
  - `type Tag struct { ID shared.ID; ProjectID shared.ID; Name string; Color string; CreatedBy shared.ID }`
  - `type Step struct { ID shared.ID; TestCaseID shared.ID; Number int; Action string; Expected string }`
  - `type TestCase struct { ID shared.ID; ProjectID shared.ID; ModuleID *shared.ID; Title string; Description string; Status CaseStatus; Priority Priority; Tags []string; Steps []Step; CreatedBy shared.ID; UpdatedBy *shared.ID; Version int; CreatedAt time.Time; UpdatedAt time.Time; DeletedAt *time.Time }`
  - `type Collection struct { ID shared.ID; ProjectID shared.ID; Name string; Description string; CreatedBy shared.ID; CreatedAt time.Time }`
  - `type CollectionCase struct { CollectionID shared.ID; TestCaseID shared.ID; AddedAt time.Time }`

#### T014: [P] Plan 值对象
- **File:** `internal/domain/plan/valueobject.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type PlanStatus string` — 常量 PlanDraft, PlanActive, PlanPaused, PlanCompleted, PlanCancelled
  - `func ValidatePlanTransition(from, to PlanStatus) error` — 实现 spec 3.3 转换规则：draft→active/cancelled, active→paused/completed/cancelled(需至少一个用例), paused→active/cancelled, completed/cancelled=终态

#### T015: [P] Plan Domain 实体
- **File:** `internal/domain/plan/entity.go`
- **Depends:** T006, T007, T014
- **Action:** Create
- **Details:**
  - `type PlanTestCase struct { PlanID shared.ID; TestCaseID shared.ID; AssignedTo *shared.ID; OrderIndex int }`
  - `type TestPlan struct { ID shared.ID; ProjectID shared.ID; Name string; Description string; Status PlanStatus; CurrentExecutionID *shared.ID; StartedAt *time.Time; PausedAt *time.Time; EndedAt *time.Time; CreatedBy shared.ID; CreatedAt time.Time; UpdatedAt time.Time; DeletedAt *time.Time }`

#### T016: [P] Execution Domain 实体
- **File:** `internal/domain/execution/entity.go`
- **Depends:** T006, T007
- **Action:** Create
- **Details:**
  - `type ExecStatus string` — 常量 ExecInProgress, ExecPaused, ExecCompleted, ExecCancelled
  - `type ResultStatus string` — 常量 ResultPassed, ResultFailed, ResultBlocked, ResultSkipped
  - `type TestExecution struct { ID shared.ID; PlanID shared.ID; Name string; Status ExecStatus; ExecutorID shared.ID; StartedAt time.Time; PausedAt *time.Time; CompletedAt *time.Time; Notes string }`
  - `type ExecutionResult struct { ID shared.ID; ExecutionID shared.ID; TestCaseID shared.ID; ExecutorID shared.ID; Status ResultStatus; BugID string; BugURL string; Notes string; ExecutedAt time.Time }`

#### T017: [P] File Domain 实体
- **File:** `internal/domain/file/entity.go`
- **Depends:** T006, T007
- **Action:** Create
- **Details:**
  - `type FileType string` — 常量 FilePDF, FileDOCX, FileXLSX, FileImage, FileFigma
  - `type SourceType string` — 常量 SourceUpload, SourceFigma
  - `type IndexStatus string` — 常量 IndexPending, IndexProcessing, IndexCompleted, IndexFailed
  - `type File struct { ID shared.ID; ProjectID shared.ID; Name string; Type FileType; Size int64; Path string; SourceType SourceType; SourceURL string; ContentPreview string; IsIndexed bool; IndexStatus IndexStatus; IndexError string; IndexedAt *time.Time; UploadedBy shared.ID; Version int; UploadedAt time.Time; DeletedAt *time.Time }`
  - `type FileVersion struct { ID shared.ID; FileID shared.ID; Version int; Path string; Size int64; UploadedBy shared.ID; UploadedAt time.Time }`

#### T018: [P] RAG Domain 实体
- **File:** `internal/domain/rag/entity.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type DocumentChunk struct { ID shared.ID; FileID shared.ID; Content string; Index int; Tokens int; CreatedAt time.Time }`
  - `type VectorEmbedding struct { ID shared.ID; ChunkID shared.ID; Model string; Dimension int; MilvusID string; CreatedAt time.Time }`
  - `type ChunkConfig struct { MaxTokens int; Overlap int; MinChunkSize int }` — 默认值 500/50/100

#### T018a: [P] AsyncTask Domain 实体
- **File:** `internal/domain/shared/task.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type AsyncTask struct { ID shared.ID; ProjectID shared.ID; Type string; Status string; ProgressCurrent int; ProgressTotal int; Input json.RawMessage; Result json.RawMessage; Error string; CreatedBy shared.ID; CreatedAt time.Time; StartedAt *time.Time; CompletedAt *time.Time }`
  - `type IndexTask struct { ID shared.ID; FileID shared.ID; Status string; RetryCount int; MaxRetries int; Error string; CreatedAt time.Time; UpdatedAt time.Time; CompletedAt *time.Time }`
  - 对应 spec.md 3.6 ai_tasks + index_tasks 表

---

### 1.4 Repository 接口定义

#### T019: [P] UserRepository 接口
- **File:** `internal/domain/user/repository.go`
- **Depends:** T010
- **Action:** Create
- **Details:** `type UserRepository interface { FindByID(ctx, id) (*User, error); FindByEmail(ctx, email string) (*User, error); Create(ctx, user *User) error; Update(ctx, user *User) error }`

#### T020: [P] ProjectRepository 接口
- **File:** `internal/domain/project/repository.go`
- **Depends:** T011
- **Action:** Create
- **Details:** `type ProjectRepository interface { Create(ctx, project *Project) error; FindByID(ctx, id) (*Project, error); FindByUserID(ctx, userID) ([]*Project, error); IsMember(ctx, projectID, userID) (bool, error); AddMember(ctx, member *ProjectMember) error; CountMembers(ctx, projectID) (int64, error) }`

#### T021: [P] TestCase Repository 接口组
- **File:** `internal/domain/testcase/repository.go`
- **Depends:** T013
- **Action:** Create
- **Details:**
  - `type ModuleRepository interface { Create(ctx, m *Module) error; Update(ctx, m *Module) error; Delete(ctx, id shared.ID) error; FindByProject(ctx, projectID) ([]*Module, error) }`
  - `type TagRepository interface { FindByProject(ctx, projectID) ([]*Tag, error); Create(ctx, tag *Tag) error }`
  - `type TestCaseRepository interface { Create(ctx, tc *TestCase) error; FindByID(ctx, id) (*TestCase, error); List(ctx, filter TestCaseFilter) ([]*TestCase, int64, error); Update(ctx, tc *TestCase) error; SoftDelete(ctx, id shared.ID) error; BatchUpdateStatus(ctx, ids []shared.ID, status CaseStatus) error; BatchDelete(ctx, ids []shared.ID) error; BatchMove(ctx, ids []shared.ID, moduleID *shared.ID) error }`
  - `type CollectionRepository interface { Create(ctx, c *Collection) error; AddCases(ctx, collectionID shared.ID, caseIDs []shared.ID) error; RemoveCases(ctx, collectionID shared.ID, caseIDs []shared.ID) error; ListCases(ctx, collectionID shared.ID, page, pageSize int) ([]*TestCase, int64, error) }`
  - `type TestCaseFilter struct { ProjectID shared.ID; ModuleID *shared.ID; Status *CaseStatus; Priority *Priority; Keyword string; Tags []string; Page int; PageSize int; SortBy string; SortDesc bool }`

#### T022: [P] TestPlanRepository 接口
- **File:** `internal/domain/plan/repository.go`
- **Depends:** T015
- **Action:** Create
- **Details:** `type TestPlanRepository interface { Create(ctx, plan *TestPlan) error; FindByID(ctx, id) (*TestPlan, error); List(ctx, projectID shared.ID, status *PlanStatus, page, pageSize int) ([]*TestPlan, int64, error); Update(ctx, plan *TestPlan) error; AddCases(ctx, planID shared.ID, cases []PlanTestCase) error; RemoveCases(ctx, planID shared.ID, caseIDs []shared.ID) error }`

#### T023: [P] ExecutionRepository 接口
- **File:** `internal/domain/execution/repository.go`
- **Depends:** T016
- **Action:** Create
- **Details:** `type ExecutionRepository interface { Create(ctx, exec *TestExecution) error; FindByID(ctx, id) (*TestExecution, error); SubmitResult(ctx, result *ExecutionResult) error; BatchSubmitResults(ctx, results []*ExecutionResult) error; GetSummary(ctx, executionID shared.ID) (*ExecutionSummary, error) }`
  - `type ExecutionSummary struct { Total int; Passed int; Failed int; Blocked int; Skipped int }`

#### T024: [P] FileRepository 接口
- **File:** `internal/domain/file/repository.go`
- **Depends:** T017
- **Action:** Create
- **Details:** `type FileRepository interface { Create(ctx, file *File) error; FindByID(ctx, id) (*File, error); FindByProject(ctx, projectID shared.ID, page, pageSize int) ([]*File, int64, error); UpdateIndexStatus(ctx, id shared.ID, status IndexStatus, errMsg string) error; SoftDelete(ctx, id shared.ID) error; CreateVersion(ctx, version *FileVersion) error }`

#### T025: [P] RAG Repository 接口组
- **File:** `internal/domain/rag/repository.go`
- **Depends:** T018
- **Action:** Create
- **Details:**
  - `type ChunkRepository interface { CreateBatch(ctx, chunks []*DocumentChunk) error; FindByFile(ctx, fileID shared.ID) ([]*DocumentChunk, error); DeleteByFile(ctx, fileID shared.ID) error }`
  - `type VectorRepository interface { Upsert(ctx, chunks []*DocumentChunk, embeddings [][]float32) error; DeleteByFile(ctx, fileID shared.ID) error; Search(ctx, projectID shared.ID, query []float32, topK int) ([]*SearchResult, error) }`
  - `type SearchResult struct { ChunkID shared.ID; Content string; Score float32 }`

#### T025a: [P] AsyncTask Repository 接口组
- **File:** `internal/domain/shared/task_repository.go`
- **Depends:** T018a
- **Action:** Create
- **Details:**
  - `type AsyncTaskRepository interface { Create(ctx, task *AsyncTask) error; FindByID(ctx, id shared.ID) (*AsyncTask, error); FindPendingByType(ctx, projectID shared.ID, taskType string, limit int) ([]*AsyncTask, error); Update(ctx, task *AsyncTask) error }`
  - `type IndexTaskRepository interface { Create(ctx, task *IndexTask) error; FindPending(ctx, limit int) ([]*IndexTask, error); FindStale(ctx, olderThan time.Duration, limit int) ([]*IndexTask, error); Update(ctx, task *IndexTask) error }`

---

### 1.5 领域事件定义

#### T026: [P] TestCase 领域事件
- **File:** `internal/domain/testcase/events.go`
- **Depends:** T009, T013
- **Action:** Create
- **Details:**
  - `type TestCaseCreatedEvent struct { ProjectID shared.ID; TestCaseID shared.ID }` — 实现 Event 接口
  - `type TestCaseUpdatedEvent struct { ProjectID shared.ID; TestCaseID shared.ID }`
  - `type TestCaseDeletedEvent struct { ProjectID shared.ID; TestCaseID shared.ID }`

#### T027: [P] File 领域事件
- **File:** `internal/domain/file/events.go`
- **Depends:** T009, T017
- **Action:** Create
- **Details:**
  - `type FileUploadedEvent struct { ProjectID shared.ID; FileID shared.ID }`
  - `type FileDeletedEvent struct { ProjectID shared.ID; FileID shared.ID }`

#### T028: [P] AI 领域事件
- **File:** `internal/domain/shared/ai_events.go`
- **Depends:** T009
- **Action:** Create
- **Details:**
  - `type AITaskCompletedEvent struct { TaskID shared.ID; Result json.RawMessage }`
  - `type AITaskFailedEvent struct { TaskID shared.ID; Error string }`

---

### 1.6 数据库迁移 DDL

#### T029: [P] 迁移 000001 — users + projects + project_members
- **File:** `scripts/migration/000001_init_schema.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** 严格按 spec.md 3.1 DDL，包含 users（含 deleted_at）、projects（含 deleted_at）、project_members 三张表。Down 文件按反向顺序 DROP。

#### T030: [P] 迁移 000002 — modules
- **File:** `scripts/migration/000002_modules.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** 按 spec.md 3.2 modules DDL + idx_modules_project, idx_modules_parent 索引 + UNIQUE(project_id, parent_id, name)

#### T031: [P] 迁移 000003 — tags
- **File:** `scripts/migration/000003_tags.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** 按 spec.md 3.2 tags DDL + idx_tags_project + UNIQUE(project_id, name)

#### T032: [P] 迁移 000004 — test_cases + test_steps
- **File:** `scripts/migration/000004_testcases.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** 按 spec.md 3.2 完整 DDL：test_cases（含 CHECK status、全部 9 个索引含 GIN tags + GIN fulltext + partial active）、test_steps（含 UNIQUE(test_case_id, number)、idx_test_steps_case）

#### T033: [P] 迁移 000005 — collections
- **File:** `scripts/migration/000005_collections.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** test_case_collections + collection_cases 表，参照 spec.md 3.2

#### T034: [P] 迁移 000006 — plans
- **File:** `scripts/migration/000006_plans.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** test_plans（含 CHECK status、current_execution_id 注释为逻辑外键）+ plan_test_cases + idx_test_plans_project, idx_test_plans_status 索引

#### T035: [P] 迁移 000007 — executions
- **File:** `scripts/migration/000007_executions.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** test_executions（含 CHECK status）+ execution_results（含 CHECK status + UNIQUE(execution_id, test_case_id)）+ idx_executions_single_active 部分唯一索引 + ALTER TABLE test_plans ADD CONSTRAINT fk_current_execution FOREIGN KEY(current_execution_id) REFERENCES test_executions(id)

#### T036: [P] 迁移 000008 — files
- **File:** `scripts/migration/000008_files.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** files（含 CHECK index_status、全部 4 个索引）+ file_versions（含 UNIQUE(file_id, version)）

#### T037: [P] 迁移 000009 — RAG + index_tasks
- **File:** `scripts/migration/000009_rag.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** document_chunks + vector_embeddings + index_tasks + 全部索引

#### T038: [P] 迁移 000010 — ai_tasks
- **File:** `scripts/migration/000010_ai_tasks.up.sql`, `.down.sql`
- **Depends:** —
- **Action:** Create
- **Details:** ai_tasks + idx_ai_tasks_project, idx_ai_tasks_status, idx_ai_tasks_created_by

---

## Phase 2: Infrastructure（基础设施实现，TDD）

> 实现 Phase 1 定义的全部接口。数据库层需要 PostgreSQL 运行。

---

### 2.1 PostgreSQL 基础设施

#### T039: PostgreSQL 连接 + TransactionManager 实现
- **File:** `internal/infrastructure/persistence/postgres/database.go`
- **Depends:** T002, T008
- **Action:** Create
- **Details:**
  - `func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error)` — 配置连接池（MaxOpenConns, MaxIdleConns, ConnMaxLifetime）
  - `type transactionManager struct { db *gorm.DB }`
  - `func NewTransactionManager(db *gorm.DB) shared.TransactionManager`
  - 实现 `InTx(ctx, fn)` — GORM Transaction 包装

#### T040: DBOrTx 上下文辅助工具
- **File:** `internal/infrastructure/persistence/postgres/helper.go`
- **Depends:** T039
- **Action:** Create
- **Details:**
  - `func DBOrTx(ctx context.Context, db *gorm.DB) *gorm.DB` — 从 ctx 提取 tx key，若在事务中返回 tx，否则返回 db
  - `func txKey() *struct{}`
  - `func withTx(ctx context.Context, tx *gorm.DB) context.Context`

#### T040a: Config Load 函数实现
- **File:** `internal/shared/config/config.go`（修改）
- **Depends:** T002
- **Action:** Modify
- **Details:**
  - 追加 `Load() (*Config, error)` 函数 — Viper 绑定环境变量
  - 追加 Docker Compose 旧命名映射逻辑（DATABASE_HOST → HEKA_DB_HOST 等）
  - Phase 1 只定义了类型，此处补充加载逻辑

#### T041: EventBus 内存实现
- **File:** `internal/infrastructure/event/bus.go`
- **Depends:** T009
- **Action:** Create
- **Details:**
  - `type eventBus struct { handlers map[string][]shared.EventHandler; queue chan eventWrapper; workers int; stop chan struct{}; done chan struct{} }`
  - `func NewEventBus(workers int) shared.EventBus`
  - 异步 Worker 消费队列
  - `Shutdown()` 停止 Worker

---

### 2.2 Auth 基础设施

#### T042: JWT 工具
- **File:** `internal/infrastructure/auth/jwt.go`
- **Depends:** T002, T006
- **Action:** Create
- **Details:**
  - `type Claims struct { UserID shared.ID; Email string; jwt.RegisteredClaims }`
  - `func GenerateToken(secret string, userID shared.ID, email string, ttl time.Duration) (string, error)`
  - `func ParseToken(secret string, tokenStr string) (*Claims, error)`
  - `func GenerateRefreshToken(secret string, userID shared.ID, ttl time.Duration) (string, error)`
  - Access Token TTL 24h, Refresh Token TTL 7d

#### T042a: 密码工具
- **File:** `internal/infrastructure/auth/password.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `func HashPassword(password string) (string, error)` — bcrypt cost=12
  - `func CheckPassword(password, hash string) bool`
  - 从 Domain Entity 移出的业务逻辑，放入 infrastructure 层

---

### 2.3 Repository TDD — 测试先行

> 每个 Repository 写测试时需要 PostgreSQL 连接（T005a docker-compose 已提供）。使用 testcontainers 或测试数据库。

#### T043: TDD RED — UserRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/user_test.go`
- **Depends:** T010, T019, T040
- **[P]**
- **Action:** Create
- **Details:**
  - TestCreateUser / TestFindByEmail / TestFindByID / TestUpdateUser / TestSoftDelete
  - 使用 testify/assert + testcontainers PostgreSQL
  - 测试引用 Phase 1 定义的接口，使用 mock 或直连数据库。因实现文件不存在，直接 DB 测试无法运行（RED）

#### T044: TDD GREEN — UserRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/user.go`
- **Depends:** T043
- **[P]**
- **Action:** Create
- **Details:**
  - `type userRepository struct { db *gorm.DB }`
  - `func NewUserRepository(db *gorm.DB) user.UserRepository`
  - GORM 实现 FindByID/FindByEmail/Create/Update + 软删除（`WHERE deleted_at IS NULL`）
  - 运行 T043 测试全部通过

#### T045: TDD RED — ProjectRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/project_test.go`
- **Depends:** T011, T020, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateProject / TestFindByUserID / TestIsMember / TestAddMember / TestCountMembers

#### T046: TDD GREEN — ProjectRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/project.go`
- **Depends:** T045
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 ProjectRepository 全部方法 + Preload members

#### T047: TDD RED — ModuleRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/module_test.go`
- **Depends:** T013, T021, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateModule / TestUpdateModule / TestDeleteModule / TestFindByProject（树形）

#### T048: TDD GREEN — ModuleRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/module.go`
- **Depends:** T047
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 + 树形查询（内存组装或递归 CTE）

#### T049: TDD RED — TagRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/tag_test.go`
- **Depends:** T013, T021, T040
- **[P]**
- **Action:** Create
- **Details:** TestFindByProject / TestCreate / TestDuplicateName

#### T050: TDD GREEN — TagRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/tag.go`
- **Depends:** T049
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 TagRepository

#### T051: TDD RED — TestCaseRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/testcase_test.go`
- **Depends:** T013, T021, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateWithSteps / TestFindByID（Preload Steps） / TestListWithFilter / TestUpdate / TestSoftDelete / TestBatchUpdateStatus / TestBatchDelete / TestBatchMove / TestVersionConflict（乐观锁）

#### T052: TDD GREEN — TestCaseRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/testcase.go`
- **Depends:** T051
- **[P]**
- **Action:** Create
- **Details:**
  - GORM 实现 TestCaseRepository 全部方法
  - Create 在事务中同时写入 test_steps
  - FindByID Preload Steps
  - List 支持 status/priority/tags(GIN)/keyword(全文搜索) + 游标分页
  - Update 使用 `WHERE version = ?` 乐观锁

#### T053: TDD RED — CollectionRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/collection_test.go`
- **Depends:** T013, T021, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreate / TestAddCases / TestRemoveCases / TestListCases（分页）

#### T054: TDD GREEN — CollectionRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/collection.go`
- **Depends:** T053
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 CollectionRepository

#### T055: TDD RED — PlanRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/plan_test.go`
- **Depends:** T015, T022, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreatePlan / TestFindByID / TestList / TestUpdate / TestAddCases / TestRemoveCases

#### T056: TDD GREEN — PlanRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/plan.go`
- **Depends:** T055
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 TestPlanRepository

#### T057: TDD RED — ExecutionRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/execution_test.go`
- **Depends:** T016, T023, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateExecution / TestSubmitResult / TestBatchSubmitResults / TestGetSummary / TestConcurrentCreate（部分唯一索引防并发）

#### T058: TDD GREEN — ExecutionRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/execution.go`
- **Depends:** T057
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 + 利用 idx_executions_single_active 部分唯一索引做并发控制

#### T059: TDD RED — FileRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/file_test.go`
- **Depends:** T017, T024, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateFile / TestFindByID / TestFindByProject / TestUpdateIndexStatus / TestSoftDelete / TestCreateVersion

#### T060: TDD GREEN — FileRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/file.go`
- **Depends:** T059
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 FileRepository + FileVersion 关联

#### T061: TDD RED — ChunkRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/chunk_test.go`
- **Depends:** T018, T025, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateBatch / TestFindByFile / TestDeleteByFile

#### T062: TDD GREEN — ChunkRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/chunk.go`
- **Depends:** T061
- **[P]**
- **Action:** Create
- **Details:** GORM 批量插入实现

#### T062a: TDD RED — AsyncTaskRepository 测试
- **File:** `internal/infrastructure/persistence/postgres/task_test.go`
- **Depends:** T018a, T025a, T040
- **[P]**
- **Action:** Create
- **Details:** TestCreateAsyncTask / TestFindByID / TestFindPendingByType / TestUpdate / TestCreateIndexTask / TestFindPending / TestFindStale / TestUpdateRetryCount

#### T062b: TDD GREEN — AsyncTaskRepository + IndexTaskRepository 实现
- **File:** `internal/infrastructure/persistence/postgres/task.go`
- **Depends:** T062a
- **[P]**
- **Action:** Create
- **Details:** GORM 实现 AsyncTaskRepository + IndexTaskRepository 全部方法

---

### 2.4 Milvus 基础设施

#### T063: Milvus Client 初始化
- **File:** `internal/infrastructure/persistence/milvus/client.go`
- **Depends:** T002
- **Action:** Create
- **Details:** `func NewClient(cfg *config.MilvusConfig) (*milvus.Client, error)` — 连接 Milvus + Ping 验证

#### T064: TDD RED — VectorRepository 测试
- **File:** `internal/infrastructure/persistence/milvus/vector_test.go`
- **Depends:** T018, T025, T063
- **Action:** Create
- **Details:** TestUpsert / TestDeleteByFile / TestSearch（需要 Milvus 运行）

#### T065: TDD GREEN — VectorRepository 实现
- **File:** `internal/infrastructure/persistence/milvus/vector.go`
- **Depends:** T064
- **Action:** Create
- **Details:**
  - Milvus SDK 实现 VectorRepository
  - 启动时自动创建 `heka_chunks` Collection（schema 参照 spec.md 12.4）
  - IVF_FLAT 索引（nlist=64，初始数据量小）

---

### 2.5 文件存储 + Figma

#### T066: [P] 本地文件存储
- **File:** `internal/infrastructure/storage/local.go`
- **Depends:** T002
- **Action:** Create
- **Details:** `type localStorage struct { basePath string }` — `Save(ctx, name, reader) (path, error)` / `GetPath(name) string` / `Delete(name) error`。按 project_id/YYYY/MM/ 组织目录

#### T067: [P] Figma Client
- **File:** `internal/infrastructure/figma/client.go`
- **Depends:** T002
- **Action:** Create
- **Details:**
  - `type Client struct { accessToken string; httpClient *http.Client }`
  - `func (c *Client) GetFile(ctx, fileURL string) (*FigmaDocument, error)` — Figma REST API
  - `func (c *Client) ExtractText(doc *FigmaDocument) (string, error)`

---

### 2.6 文件解析器

#### T068: Parser 接口 + 注册表
- **File:** `internal/infrastructure/parser/parser.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type Parser interface { Parse(ctx, reader io.Reader) (string, error); SupportedTypes() []string }`
  - `type Registry struct { parsers map[string]Parser }`
  - `func NewRegistry() *Registry` + `Register(fileType string, p Parser)` + `GetParser(fileType string) (Parser, error)`

#### T069: TDD RED — PDF 解析器测试
- **File:** `internal/infrastructure/parser/pdf_test.go`
- **Depends:** T068
- **[P]**
- **Action:** Create
- **Details:** TestParsePDF — 使用测试 PDF 文件验证文本提取

#### T070: TDD GREEN — PDF 解析器
- **File:** `internal/infrastructure/parser/pdf.go`
- **Depends:** T069
- **[P]**
- **Action:** Create
- **Details:** 使用 `ledongthuc/pdfgof` 或 `unidoc/unipdf` 提取文本

#### T071: TDD RED — DOCX 解析器测试
- **File:** `internal/infrastructure/parser/docx_test.go`
- **Depends:** T068
- **[P]**
- **Action:** Create
- **Details:** TestParseDOCX — 使用测试 .docx 文件验证

#### T072: TDD GREEN — DOCX 解析器
- **File:** `internal/infrastructure/parser/docx.go`
- **Depends:** T071
- **[P]**
- **Action:** Create
- **Details:** 使用 `nguyenthenguyen/docx` 提取文本

#### T073: TDD RED — XLSX 解析器测试
- **File:** `internal/infrastructure/parser/xlsx_test.go`
- **Depends:** T068
- **[P]**
- **Action:** Create
- **Details:** TestParseXLSX — 使用测试 .xlsx 文件验证

#### T074: TDD GREEN — XLSX 解析器
- **File:** `internal/infrastructure/parser/xlsx.go`
- **Depends:** T073
- **[P]**
- **Action:** Create
- **Details:** 使用 `excelize/v2` 提取文本

#### T075: TDD RED — 图片 OCR 解析器测试
- **File:** `internal/infrastructure/parser/image_test.go`
- **Depends:** T068
- **[P]**
- **Action:** Create
- **Details:** TestParseImage — 使用测试图片验证 OCR

#### T076: TDD GREEN — 图片 OCR 解析器
- **File:** `internal/infrastructure/parser/image.go`
- **Depends:** T075
- **[P]**
- **Action:** Create
- **Details:** 使用 Tesseract CLI 或 Go OCR 库

#### T077: TDD RED — 文本分块器测试
- **File:** `internal/infrastructure/parser/chunker_test.go`
- **Depends:** T018
- **Action:** Create
- **Details:** TestSemanticOverlap — 验证段落分块、超长段落按句分割、overlap 重叠、小块丢弃

#### T078: TDD GREEN — 文本分块器
- **File:** `internal/infrastructure/parser/chunker.go`
- **Depends:** T077
- **Action:** Create
- **Details:** 实现 semantic_overlap 分块策略，参照 spec.md 3.5 ChunkConfig

---

### 2.7 AI 基础设施

#### T079: [P] AI 类型定义
- **File:** `internal/infrastructure/ai/types.go`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `type ChatRequest struct { Model string; Messages []Message; MaxTokens int; Temperature float64 }`
  - `type Message struct { Role string; Content string }`
  - `type ChatResponse struct { Content string; TokensUsed int; Model string }`
  - `type LLMClient interface { Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) }`
  - `type ProviderConfig struct { Name string; APIKey string; BaseURL string; Model string; Priority int }`

#### T080: TDD RED — Circuit Breaker 测试
- **File:** `internal/infrastructure/ai/breaker_test.go`
- **Depends:** —
- **[P]**
- **Action:** Create
- **Details:** TestClosedState / TestOpenAfterConsecutiveFailures / TestHalfOpen / TestResetOnSuccess

#### T081: TDD GREEN — Circuit Breaker 实现
- **File:** `internal/infrastructure/ai/breaker.go`
- **Depends:** T080
- **[P]**
- **Action:** Create
- **Details:** 三状态机 Closed/Open/HalfOpen，连续 5 次失败熔断，30s 后半开

#### T082: TDD RED — Retry Wrapper 测试
- **File:** `internal/infrastructure/ai/retry_test.go`
- **Depends:** —
- **[P]**
- **Action:** Create
- **Details:** TestRetryOnFailure / TestExponentialBackoff / TestMaxAttempts

#### T083: TDD GREEN — Retry Wrapper 实现
- **File:** `internal/infrastructure/ai/retry.go`
- **Depends:** T082
- **[P]**
- **Action:** Create
- **Details:** 指数退避 1s → 2s → 4s → 8s，max 30s，max 3 attempts

#### T084: [P] Timeout Config
- **File:** `internal/infrastructure/ai/timeout.go`
- **Depends:** —
- **Action:** Create
- **Details:** 分级超时常量 — Dial 10s, TLS 5s, ResponseHeader 30s, Request 60s, Generation 55s。`func ApplyTimeout(ctx, cfg) context.Context`

#### T085: TDD RED — Worker Pool 测试
- **File:** `internal/infrastructure/ai/pool_test.go`
- **Depends:** —
- **[P]**
- **Action:** Create
- **Details:** TestConcurrentExecution / TestQueueFull / TestStop

#### T086: TDD GREEN — Worker Pool 实现
- **File:** `internal/infrastructure/ai/pool.go`
- **Depends:** T085
- **[P]**
- **Action:** Create
- **Details:** 协程池 + 任务队列（默认 10 worker, queue 50）+ Stop 可停止

#### T087: [P] Claude Provider
- **File:** `internal/infrastructure/ai/claude.go`
- **Depends:** T079
- **Action:** Create
- **Details:** Anthropic Messages API 封装，实现 LLMClient 接口

#### T088: [P] OpenAI Provider
- **File:** `internal/infrastructure/ai/openai.go`
- **Depends:** T079
- **Action:** Create
- **Details:** OpenAI Chat Completions API 封装，实现 LLMClient 接口

#### T089: [P] Gemini Provider
- **File:** `internal/infrastructure/ai/gemini.go`
- **Depends:** T079
- **Action:** Create
- **Details:** Gemini GenerateContent API 封装，实现 LLMClient 接口

#### T090: [P] Ollama Provider
- **File:** `internal/infrastructure/ai/ollama.go`
- **Depends:** T079
- **Action:** Create
- **Details:** Ollama HTTP API 封装，实现 LLMClient 接口

#### T091: Manager（故障转移）
- **File:** `internal/infrastructure/ai/manager.go`
- **Depends:** T079, T081, T083, T084, T086, T087, T088, T089, T090
- **Action:** Create
- **Details:**
  - `type Manager struct` — 按 Priority 排序 Providers + 每个Provider 配 Breaker
  - `func (m *Manager) Chat(ctx, req) (*ChatResponse, error)` — 优先级遍历 + 熔断跳过 + 重试包装
  - `func (m *Manager) StreamChat(ctx, req, callback) error`
  - 全 Provider 熔断时返回缓存或模板

#### T092: [P] Prompt 模板
- **File:** `internal/infrastructure/ai/prompt.go`
- **Depends:** —
- **Action:** Create
- **Details:** 结构化 Prompt 模板：用例生成（含 `<document_content>` 分隔符）、智能分析。参照 spec.md 4.2 + 7.4

#### T093: [P] SanitizeInput + ValidateAIOutput
- **File:** `internal/infrastructure/ai/sanitize.go`
- **Depends:** —
- **Action:** Create
- **Details:** `func SanitizeInput(input string) string`（移除 prompt injection 模式）+ `func ValidateAIOutput(output string) error`。参照 spec.md 4.1

#### T094: [P] Embedding Client（独立轻量客户端）
- **File:** `internal/infrastructure/ai/embedding.go`
- **Depends:** T079
- **Action:** Create
- **Details:**
  - `type EmbeddingClient struct { httpClient *http.Client; apiKey string; model string }`
  - `func (c *EmbeddingClient) Embed(ctx, texts []string) ([][]float32, error)`
  - 调用 OpenAI embedding API，不依赖 Pool/Breaker/Manager

---

### 2.8 Worker

#### T095: Index Worker（RAG 索引补偿）
- **File:** `internal/infrastructure/worker/index_worker.go`
- **Depends:** T024, T025, T025a, T062, T062b, T065, T068, T078, T094
- **Action:** Create
- **Details:**
  - `type IndexWorker struct` — 注入 FileRepository, ChunkRepository, VectorRepository, IndexTaskRepository, ParserRegistry, Chunker, EmbeddingClient, EventBus
  - `func (w *IndexWorker) Start(ctx)` — 每 5s 扫描 index_tasks(status=pending)
  - 补偿逻辑：created_at > 10min 且 status=pending 的记录也扫描
  - retry_count++ 超过 max_retries 标记 failed

#### T096: AI Task Worker
- **File:** `internal/infrastructure/worker/ai_worker.go`
- **Depends:** T025a, T041, T028, T091, T092, T093
- **Action:** Create
- **Details:**
  - `type AIWorker struct` — 注入 Manager, AsyncTaskRepository, EventBus
  - 消费 ai_tasks(status=pending) → 调用 LLM → 写入结果 → 发布事件 ai.task.completed/failed

---

### 2.9 Cache

#### T097: Redis Cache Client
- **File:** `internal/infrastructure/cache/redis.go`
- **Depends:** T002
- **Action:** Create
- **Details:**
  - `type CacheClient struct { client *redis.Client }`
  - `func NewCacheClient(cfg *config.RedisConfig) (*CacheClient, error)`
  - `Get(ctx, key) (string, error)` / `Set(ctx, key, value string, ttl time.Duration) error` / `Delete(ctx, keys ...string) error` / `DeleteByPattern(ctx, pattern string) error`

---

## Phase 3: Application + Interface（业务逻辑 + API，TDD）

> 严格 TDD：测试 → 实现。Service 测试使用 mock Repository。

---

### 3.0 HTTP 基础设施

#### T098: [P] 统一响应封装
- **File:** `internal/interface/http/response/response.go`
- **Depends:** T007
- **Action:** Create
- **Details:**
  - `func Success(w, data interface{})` — `{"code":0,"data":...,"message":"success"}`
  - `func Error(w, appErr *shared.AppError)` — `{"code":"TC-NF-001","message":"测试用例不存在"}`
  - `func PageResult(w, data interface{}, total int64, page, pageSize int)`
  - `func Created(w, data interface{})` — HTTP 201

---

### 3.1 User/Application — TDD

#### T099: User DTO
- **File:** `internal/application/user/dto.go`
- **Depends:** T006, T007
- **Action:** Create
- **Details:**
  - `type LoginRequest struct { Email string \`validate:"required,email"\`; Password string \`validate:"required"\` }`
  - `type UserResponse struct { ID shared.ID; Name string; Email string }`
  - `type TokenResponse struct { AccessToken string; RefreshToken string; ExpiresAt int64 }`

#### T100: TDD RED — UserService 测试
- **File:** `internal/application/user/service_test.go`
- **Depends:** T010, T019, T099
- **Action:** Create
- **Details:**
  - Mock UserRepository（使用 testify/mock）
  - TestLogin_Success / TestLogin_WrongPassword / TestLogin_UserNotFound（不区分用户不存在和密码错误）
  - TestGetMe / TestRefreshToken_Success / TestRefreshToken_Expired

#### T101: TDD GREEN — UserService 实现
- **File:** `internal/application/user/service.go`
- **Depends:** T100, T042, T042a
- **Action:** Create
- **Details:**
  - `type Service struct { repo user.UserRepository; jwtSecret string; accessTokenTTL time.Duration; refreshTokenTTL time.Duration }`
  - `func (s *Service) Login(ctx, req LoginRequest) (*TokenResponse, error)` — 使用 password.CheckPassword 校验
  - `func (s *Service) GetMe(ctx, userID shared.ID) (*UserResponse, error)`
  - `func (s *Service) RefreshToken(ctx, refreshToken string) (*TokenResponse, error)`

---

### 3.2 Project/Application — TDD

#### T102: Project DTO
- **File:** `internal/application/project/dto.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type CreateProjectReq struct { Name string \`validate:"required"\`; Description string }`
  - `type ProjectResponse struct { ID shared.ID; Name string; Description string; Members []MemberResponse; Stats ProjectStats; CreatedAt time.Time }`
  - `type MemberResponse struct { UserID shared.ID; Name string; Email string; JoinedAt time.Time }`
  - `type ProjectStats struct { TestCaseCount int64; PlanCount int64; MemberCount int }`
  - `type AddMemberReq struct { UserID shared.ID \`validate:"required,uuid"\` }`

#### T103: TDD RED — ProjectService 测试
- **File:** `internal/application/project/service_test.go`
- **Depends:** T011, T020, T102
- **Action:** Create
- **Details:** Mock ProjectRepository. TestCreate / TestGetByID（含统计）/ TestListByUser / TestAddMember / TestAddMember_AlreadyExists

#### T104: TDD GREEN — ProjectService 实现
- **File:** `internal/application/project/service.go`
- **Depends:** T103
- **Action:** Create
- **Details:**
  - `type Service struct { repo project.ProjectRepository }`
  - `Create` / `GetByID`（查询成员 + 统计用例/计划数）/ `ListByUser` / `AddMember`（幂等检查）

---

### 3.3 TestCase/Application — TDD

#### T105: TestCase DTO
- **File:** `internal/application/testcase/dto.go`
- **Depends:** T006, T012, T013
- **Action:** Create
- **Details:**
  - `type CreateTestCaseReq struct` — title, description, module_id, priority, tags, steps（action+expected）
  - `type UpdateTestCaseReq struct` — 同上 + version（乐观锁）
  - `type TestCaseResponse struct` — 全字段 + steps
  - `type TestCaseListResponse struct` — 列表项
  - `type BatchStatusReq struct { IDs []shared.ID; Status testcase.CaseStatus }`
  - `type BatchDeleteReq struct { IDs []shared.ID }`
  - `type BatchMoveReq struct { IDs []shared.ID; ModuleID *shared.ID }`
  - `type ModuleDTO`, `ModuleTreeResponse`, `TagDTO`, `CollectionDTO` 等辅助 DTO

#### T106: TDD RED — TestCaseService 测试
- **File:** `internal/application/testcase/service_test.go`
- **Depends:** T013, T021, T026, T105
- **Action:** Create
- **Details:**
  - Mock 全部 testcase Repository + EventBus
  - TestCreateTestCase / TestCreateWithSteps / TestGetByID / TestListWithFilter
  - TestUpdateStatusTransition / TestInvalidTransition
  - TestUpdateVersionConflict（乐观锁 409）
  - TestBatchUpdateStatus / TestBatchDelete / TestBatchMove
  - TestCreateModule / TestGetModuleTree / TestCreateTag / TestCreateCollection

#### T107: TDD GREEN — TestCaseService 实现
- **File:** `internal/application/testcase/service.go`
- **Depends:** T106, T009
- **Action:** Create
- **Details:**
  - `type Service struct { tcRepo testcase.TestCaseRepository; modRepo testcase.ModuleRepository; tagRepo testcase.TagRepository; collRepo testcase.CollectionRepository; eventBus shared.EventBus }`
  - 实现 Module/Tag/TestCase/Collection 全部业务逻辑
  - 状态转换调用 `ValidateCaseTransition`
  - 更新检查 version 乐观锁
  - 创建/更新/删除后发布领域事件

---

### 3.4 Plan/Application — TDD

#### T108: Plan DTO
- **File:** `internal/application/plan/dto.go`
- **Depends:** T006, T014, T015
- **Action:** Create
- **Details:**
  - `type CreatePlanReq struct { Name string; Description string; CaseIDs []PlanCaseItem }`
  - `type PlanResponse struct` — 含状态、进度、当前执行
  - `type PlanDetailResponse struct` — 含关联用例列表 + 执行摘要

#### T109: TDD RED — PlanService 测试
- **File:** `internal/application/plan/service_test.go`
- **Depends:** T015, T022, T108
- **Action:** Create
- **Details:** Mock TestPlanRepository. TestCreate / TestStart（draft→active）/ TestStartWithoutCases / TestPause / TestResume / TestComplete / TestCancel / TestInvalidTransition

#### T110: TDD GREEN — PlanService 实现
- **File:** `internal/application/plan/service.go`
- **Depends:** T109
- **Action:** Create
- **Details:**
  - `type Service struct { repo plan.TestPlanRepository }`
  - CRUD + 状态流转（调用 ValidatePlanTransition）
  - Start 时检查至少关联 1 个用例

---

### 3.5 Execution/Application — TDD

#### T111: Execution DTO
- **File:** `internal/application/execution/dto.go`
- **Depends:** T006, T016
- **Action:** Create
- **Details:**
  - `type SubmitResultReq struct { TestCaseID shared.ID; Status ResultStatus; BugID string; BugURL string; Notes string }`
  - `type BatchSubmitReq struct { Results []SubmitResultReq }`
  - `type ExecutionResponse struct` — 含状态 + 汇总统计
  - `type ExecutionSummaryResponse struct { Total int; Passed int; Failed int; Blocked int; Skipped int }`

#### T112: TDD RED — ExecutionService 测试
- **File:** `internal/application/execution/service_test.go`
- **Depends:** T016, T023, T111
- **Action:** Create
- **Details:** Mock ExecutionRepository. TestCreateExecution / TestSubmitResult / TestBatchSubmit / TestSubmitDuplicate（409）/ TestConcurrentCreate（部分唯一索引冲突）

#### T113: TDD GREEN — ExecutionService 实现
- **File:** `internal/application/execution/service.go`
- **Depends:** T112
- **Action:** Create
- **Details:**
  - `type Service struct { repo execution.ExecutionRepository }`
  - Create/SubmitResult/BatchSubmit/GetByID/GetSummary

---

### 3.6 File/Application — TDD

#### T114: File DTO
- **File:** `internal/application/file/dto.go`
- **Depends:** T006, T017
- **Action:** Create
- **Details:**
  - `type FileResponse struct` — 全字段
  - `type FileListResponse struct`
  - `type IndexStatusResponse struct { FileID shared.ID; Status IndexStatus; Error string }`
  - `type FigmaUploadReq struct { URL string \`validate:"required,url"\` }`

#### T115: TDD RED — FileService 测试
- **File:** `internal/application/file/service_test.go`
- **Depends:** T017, T024, T114
- **Action:** Create
- **Details:** Mock FileRepository, Storage, EventBus. TestUpload / TestUploadInvalidMIME / TestUploadExceedsSize / TestDelete / TestReindex / TestFigmaUpload

#### T116: TDD GREEN — FileService 实现
- **File:** `internal/application/file/service.go`
- **Depends:** T115, T009
- **Action:** Create
- **Details:**
  - `type Service struct { repo file.FileRepository; storage Storage; figma FigmaClient; eventBus shared.EventBus; maxUploadSize int64 }`
  - Upload：校验 MIME + 大小 + 文件头检测 → 存储物理文件 → 写 DB → 发布 FileUploadedEvent
  - Delete：软删除 DB 记录
  - Reindex：重置 index_status → 发布事件

---

### 3.7 RAG/Application — TDD

#### T117: RAG DTO
- **File:** `internal/application/rag/dto.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type SearchRequest struct { ProjectID shared.ID; Query string; TopK int }`
  - `type SearchResponse struct { Results []SearchResultItem }`
  - `type SearchResultItem struct { ChunkID shared.ID; Content string; Score float32; FileName string }`

#### T118: TDD RED — RAG Service 测试
- **File:** `internal/application/rag/service_test.go`
- **Depends:** T018, T025, T117
- **Action:** Create
- **Details:** Mock VectorRepository. TestSearch / TestSearchEmptyQuery / TestSearchTopK

#### T119: TDD GREEN — RAG Service 实现
- **File:** `internal/application/rag/service.go`
- **Depends:** T118
- **Action:** Create
- **Details:**
  - `type Service struct { vectorRepo rag.VectorRepository; embeddingClient *ai.EmbeddingClient }`
  - `Search(ctx, req) (*SearchResponse, error)` — query → embedding → vector search

---

### 3.8 AI/Application — TDD

#### T120: AI DTO
- **File:** `internal/application/ai/dto.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type GenerateRequest struct { ProjectID shared.ID; FileID shared.ID; Count int; IncludeNegative bool }`
  - `type AnalyzeRequest struct { ProjectID shared.ID; Description string }`
  - `type TaskResponse struct { TaskID shared.ID; Status string; Progress int; Total int; Result json.RawMessage; Error string }`

#### T121: TDD RED — AIService 测试
- **File:** `internal/application/ai/service_test.go`
- **Depends:** T018a, T025a, T120
- **Action:** Create
- **Details:**
  - Mock Manager, AsyncTaskRepository, RAGService, EventBus
  - TestGenerateTestCases / TestAnalyze / TestGetTask / TestGenerate_QueueFull（429）

#### T122: TDD GREEN — AIService 实现
- **File:** `internal/application/ai/service.go`
- **Depends:** T121, T009
- **Action:** Create
- **Details:**
  - `type Service struct { manager *ai.Manager; ragSvc *rag.Service; taskRepo shared.AsyncTaskRepository; eventBus shared.EventBus }`
  - `GenerateTestCases` — RAG 检索上下文 + 创建 ai_tasks 记录 + 异步
  - `Analyze` — 同上
  - `GetTask` — 查询任务状态

---

### 3.9 Report/Application — TDD

#### T123: Report DTO
- **File:** `internal/application/report/dto.go`
- **Depends:** T006
- **Action:** Create
- **Details:**
  - `type PlanReportResponse struct` — 计划报告（通过率、执行进度）
  - `type CoverageResponse struct` — 覆盖度（按模块统计）
  - `type TrendResponse struct` — 趋势（按天统计执行结果）
  - `type BugDistributionResponse struct` — 缺陷分布
  - `type WorkloadResponse struct` — 个人工作量

#### T124: TDD RED — ReportService 测试
- **File:** `internal/application/report/service_test.go`
- **Depends:** T123
- **Action:** Create
- **Details:** Mock PlanRepo, ExecutionRepo, TestCaseRepo, UserRepo. TestPlanReport / TestCoverage / TestTrend / TestBugDistribution / TestWorkload

#### T125: TDD GREEN — ReportService 实现
- **File:** `internal/application/report/service.go`
- **Depends:** T124
- **Action:** Create
- **Details:**
  - `type Service struct { planRepo plan.TestPlanRepository; execRepo execution.ExecutionRepository; tcRepo testcase.TestCaseRepository; userRepo user.UserRepository }`
  - 5 种报告查询 + 统计计算（SQL 聚合或内存计算）

---

### 3.10 Cache-Aside 集成

#### T125a: Cache-Aside 装饰器
- **File:** `internal/application/testcase/cache.go`
- **Depends:** T097, T107
- **Action:** Create
- **Details:**
  - `type CachedTestCaseService struct` — 包装 TestCaseService，注入 CacheClient
  - 读操作：先查缓存（key `testcase:{id}`），未命中查 DB → 写缓存
  - 列表：缓存 `project:{id}:testcases:{page}:{hash}`，TTL 5min
  - 模块树：缓存 `project:{id}:modules`，TTL 10min
  - 写操作：先写 DB → 删除相关缓存

#### T125b: EventBus 订阅缓存失效
- **File:** `internal/application/cache/subscriber.go`
- **Depends:** T009, T041, T097
- **Action:** Create
- **Details:**
  - `func RegisterCacheHandlers(bus shared.EventBus, cache *cache.CacheClient)`
  - 订阅 `testcase.created/updated/deleted` → `DeleteByPattern("project:{id}:testcases:*")` + `Delete("testcase:{id}")`
  - 订阅 `file.uploaded` → `Delete("file:{id}:index")`

---

### 3.11 HTTP Handler — TDD

#### T126: TDD RED — Auth Handler 测试
- **File:** `internal/interface/http/handler/auth_test.go`
- **Depends:** T098, T099, T101
- **Action:** Create
- **Details:** httptest + mock UserService. TestLoginEndpoint / TestRefreshEndpoint / TestMeEndpoint / TestUnauthorized

#### T127: TDD GREEN — Auth Handler 实现
- **File:** `internal/interface/http/handler/auth.go`
- **Depends:** T126
- **Action:** Create
- **Details:** POST /login, POST /refresh, GET /me — 调用 UserService，返回 JSON

#### T128: TDD RED — Project Handler 测试
- **File:** `internal/interface/http/handler/project_test.go`
- **Depends:** T098, T102, T104
- **Action:** Create
- **Details:** TestCreateProject / TestGetProject / TestListProjects / TestAddMember

#### T129: TDD GREEN — Project Handler 实现
- **File:** `internal/interface/http/handler/project.go`
- **Depends:** T128
- **Action:** Create
- **Details:** 4 个端点，参数校验 + 调用 ProjectService

#### T130: TDD RED — Testcase Handler 测试
- **File:** `internal/interface/http/handler/testcase_test.go`
- **Depends:** T098, T105, T107
- **Action:** Create
- **Details:** Module CRUD + Tag CRUD + TestCase 8 端点 + Collection 4 端点

#### T131: TDD GREEN — Testcase Handler 实现
- **File:** `internal/interface/http/handler/testcase.go`
- **Depends:** T130
- **Action:** Create
- **Details:** 全部 18 个端点（4 module + 2 tag + 8 testcase + 4 collection），参数解析 + 校验 + 调用 Service

#### T132: TDD RED — Plan Handler 测试
- **File:** `internal/interface/http/handler/plan_test.go`
- **Depends:** T098, T108, T110
- **Action:** Create
- **Details:** 8 个端点（CRUD + start/pause/resume/complete/cancel）

#### T133: TDD GREEN — Plan Handler 实现
- **File:** `internal/interface/http/handler/plan.go`
- **Depends:** T132
- **Action:** Create
- **Details:** 8 个端点

#### T134: TDD RED — Execution Handler 测试
- **File:** `internal/interface/http/handler/execution_test.go`
- **Depends:** T098, T111, T113
- **Action:** Create
- **Details:** 3 个端点（submit result / batch submit / get detail）

#### T135: TDD GREEN — Execution Handler 实现
- **File:** `internal/interface/http/handler/execution.go`
- **Depends:** T134
- **Action:** Create
- **Details:** 3 个端点

#### T136: TDD RED — File Handler 测试
- **File:** `internal/interface/http/handler/file_test.go`
- **Depends:** T098, T114, T116
- **Action:** Create
- **Details:** 7 个端点（upload / figma / list / detail / reindex / index-status / delete）

#### T137: TDD GREEN — File Handler 实现
- **File:** `internal/interface/http/handler/file.go`
- **Depends:** T136
- **Action:** Create
- **Details:** 7 个端点。Upload 使用 multipart/form-data 解析

#### T138: TDD RED — AI Handler 测试
- **File:** `internal/interface/http/handler/ai_test.go`
- **Depends:** T098, T120, T122
- **Action:** Create
- **Details:** 4 个端点（generate / get task / sse events / analyze）

#### T139: TDD GREEN — AI Handler 实现
- **File:** `internal/interface/http/handler/ai.go`
- **Depends:** T138
- **Action:** Create
- **Details:**
  - 4 个端点 + SSE 推送
  - SSE 实现：使用 `http.Flusher`，设置 `Content-Type: text/event-stream`，`Cache-Control: no-cache`，`Connection: keep-alive`
  - 推送格式：`fmt.Fprintf(w, "data: %s\n\n", jsonPayload)` + `w.(http.Flusher).Flush()`
  - 客户端断开时通过 `r.Context().Done()` 检测并退出

#### T140: TDD RED — Report Handler 测试
- **File:** `internal/interface/http/handler/report_test.go`
- **Depends:** T098, T123, T125
- **Action:** Create
- **Details:** 5 个端点

#### T141: TDD GREEN — Report Handler 实现
- **File:** `internal/interface/http/handler/report.go`
- **Depends:** T140
- **Action:** Create
- **Details:** 5 个端点

#### T142: Monitoring Handler
- **File:** `internal/interface/http/handler/monitoring.go`
- **Depends:** T098
- **Action:** Create
- **Details:**
  - `GET /api/v1/health` — DB + Redis + Milvus 连通性
  - `GET /api/v1/monitoring/ai` — 各 Provider 熔断状态

---

### 3.12 Middleware

#### T143: [P] Auth Middleware
- **File:** `internal/interface/http/middleware/auth.go`
- **Depends:** T042, T004
- **Action:** Create
- **Details:** Bearer Token 解析 → ParseToken → WithUserID 注入 Context。无效 Token 返回 401

#### T144: [P] Project Access Middleware
- **File:** `internal/interface/http/middleware/project.go`
- **Depends:** T020, T004
- **Action:** Create
- **Details:** 从 query/body 提取 project_id → IsMember 校验 → 非成员返回 403

#### T145: [P] CORS Middleware
- **File:** `internal/interface/http/middleware/cors.go`
- **Depends:** —
- **Action:** Create
- **Details:** 白名单 origins 配置。参照 spec.md 5.1

#### T146: [P] Logger Middleware
- **File:** `internal/interface/http/middleware/logger.go`
- **Depends:** T003
- **Action:** Create
- **Details:** 请求日志（method, path, status, latency）+ 慢请求告警（> 1s）

#### T147: [P] Rate Limit Middleware
- **File:** `internal/interface/http/middleware/ratelimit.go`
- **Depends:** —
- **Action:** Create
- **Details:** 令牌桶实现。全局 30/s, AI 2/min, 上传 5/min, 登录 5/min。超限返回 429 SYS-RT-001

---

### 3.13 Router

#### T148: 路由注册
- **File:** `internal/interface/http/router.go`
- **Depends:** T126-T147
- **Action:** Create
- **Details:**
  - Chi/Echo 路由注册全部 54 个端点（Auth 3 + Project 4 + Module 4 + Tag 2 + TC 8 + Collection 4 + Plan 8 + Exec 3 + File 7 + AI 4 + Report 5 + Monitor 2）
  - 中间件链：Logger → CORS → Auth（排除 /login） → RateLimit → ProjectAccess
  - 分组：auth / projects / modules / tags / testcases / collections / plans / executions / files / ai / reports / monitoring

---

## Phase 4: Integration（集成与部署）

> 组装全部组件 + Docker + E2E 测试

---

### 4.1 种子数据

#### T149: 种子脚本
- **File:** `scripts/seed/admin.go`
- **Depends:** T029, T044, T042a
- **Action:** Create
- **Details:**
  - CLI 命令：`go run scripts/seed/admin.go --email admin@heka.io --password xxx --name Admin`
  - 创建初始管理员用户（调用 password.HashPassword）

---

### 4.2 依赖注入 + 启动

#### T150: main.go 完整 DI + Graceful Shutdown
- **File:** `cmd/server/main.go`
- **Depends:** Phase 2 全部完成, Phase 3 全部完成
- **Action:** Modify（替换空骨架）
- **Details:**
  - 依赖注入顺序：Config → Logger → DB → Redis → Milvus → Repos → Services → Handlers → Router → Server
  - Graceful Shutdown：SIGTERM → http.Server.Shutdown → EventBus.Shutdown → DB.Close → Redis.Close
  - 启动 Milvus Collection 初始化
  - 启动 Index Worker + AI Worker goroutine
  - 注册缓存失效订阅（T125b RegisterCacheHandlers）

---

### 4.3 Swagger

#### T151: Swagger 注解集成
- **File:** `docs/swagger.go`
- **Depends:** T148
- **Action:** Create
- **Details:**
  - `swag init` 配置
  - main.go 添加 Swagger 注解（title, version, host, basePath）
  - Router 注册 swagger-ui 静态文件路由

---

### 4.4 Docker

#### T152: [P] Dockerfile（多阶段构建）
- **File:** `Dockerfile`
- **Depends:** —
- **Action:** Create
- **Details:** golang:1.21-alpine builder → 最终镜像 < 50MB (scratch 或 distroless)

#### T153: Docker Compose — Milvus 扩展
- **File:** `docker-compose.override.yml`
- **Depends:** T005a
- **Action:** Create
- **Details:** Milvus 2.3 standalone + etcd + MinIO（Milvus 依赖）。`docker compose -f docker-compose.yml -f docker-compose.override.yml up` 加载完整环境

#### T154: [P] Nginx 配置
- **File:** `nginx.conf`
- **Depends:** —
- **Action:** Create
- **Details:** 反向代理 Backend:8080 + Frontend:3000。HTTPS 配置 + WebSocket 支持（SSE）

---

### 4.5 配置 + 文档

#### T155: [P] 配置示例文件
- **File:** `configs/config.example.yaml`
- **Depends:** T002
- **Action:** Create
- **Details:** 全部环境变量 + 注释说明（参照 spec.md 第 8 节）

---

### 4.6 E2E 测试

#### T156: TDD RED — E2E 测试
- **File:** `tests/e2e/full_flow_test.go`
- **Depends:** T148
- **Action:** Create
- **Details:**
  - 使用 httptest + 真实数据库（testcontainers）
  - 全流程：种子用户 → 登录 → 创建项目 → 添加成员 → 创建模块/标签 → CRUD 用例 → 创建计划 → 执行 → 提交结果 → 上传文件 → AI 生成 → 查看报告
  - 预期：多数失败（依赖完整系统）

#### T157: TDD GREEN — E2E 通过
- **File:** `cmd/server/main.go`（微调）
- **Depends:** T150, T156
- **Action:** Modify
- **Details:**
  - 修复 E2E 暴露的集成问题
  - 确保全部 E2E 测试通过

#### T158: 数据库迁移验证
- **File:** `tests/migration/migration_test.go`
- **Depends:** T029-T038
- **Action:** Create
- **Details:**
  - 从空库执行全部 up 迁移 → 验证表存在 + 索引 + 约束
  - 执行全部 down 迁移 → 验证清理干净

---

### 4.7 文档

#### T159: [P] README
- **File:** `README.md`
- **Depends:** —
- **Action:** Create
- **Details:**
  - 项目简介 + 技术栈
  - 快速启动（Docker Compose + seed + run）
  - 环境变量说明
  - API 文档入口（Swagger URL）

---

## 任务统计

| Phase | 任务数 | 说明 |
|-------|--------|------|
| Phase 1: Foundation | 42 | 类型、接口、DDL 定义 + Docker Compose |
| Phase 2: Infrastructure | 66 | TDD: 测试先行 → 实现（含 ai_tasks/index_tasks repo） |
| Phase 3: Application + Interface | 56 | TDD: DTO → 测试 → 实现 + Cache-Aside |
| Phase 4: Integration | 11 | DI、Docker、E2E |
| **合计** | **175** | |

---

## 依赖关系总览

```
Phase 1 (T001-T038)
  ↓
Phase 2 (T039-T099) — 依赖 Phase 1 的接口和类型
  ↓
Phase 3 (T100-T155) — 依赖 Phase 2 的实现
  ↓
Phase 4 (T156-T166) — 依赖 Phase 3 的全部组件
```

**Phase 内并行性**：
- Phase 1: T006-T009 之间 [P]、T010-T018a 之间 [P]、T019-T025a 之间 [P]、T029-T038 之间 [P]
- Phase 2: T043-T062b 各组测试/实现之间 [P]、T069-T076 各解析器之间 [P]、T080-T090 AI 组件之间 [P]
- Phase 3: 各模块 TDD 链之间 [P]（但同一模块内严格 TDD 顺序）

---

**文档版本**：v1.1（review 修复）
**最后更新**：2026-05-15
