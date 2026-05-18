# Heka 后端实现计划

> 基于 `specs/001-backend-functionality/spec.md` v1.2 拆解的详细任务清单
> 版本：v1.1（review 修复）
> 日期：2026-05-15

---

## 说明

- 每个任务标注 **D**(Domain) / **I**(Infrastructure) / **A**(Application) / **IF**(Interface) 层
- 同一小节内任务按表格从上到下的顺序即隐含依赖关系（先定义后实现，先底层后上层）
- 预估工时为单人开发估计，含测试
- 验收标准（AC）= 该阶段完成时可演示的功能

---

## 阶段 1：基础骨架（1 周）

> 交付目标：用户登录、创建项目、添加成员，JWT 全链路通

### 1.1 项目初始化 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.1.1 | 初始化 Go Module，目录结构 | — | `go.mod`, `cmd/server/main.go` | 可编译的空项目 |
| 1.1.2 | 配置加载（Viper） | shared | `internal/shared/config/config.go` | 结构体 `Config` + 环境变量绑定 |
| 1.1.3 | 日志初始化（Zap） | shared | `internal/shared/logger/logger.go` | 全局 logger |
| 1.1.4 | HTTP 服务器骨架（Chi/Echo） | IF | `internal/interface/http/router.go` | 健康检查端点 `GET /api/v1/health` |
| 1.1.5 | Docker Compose 开发环境（Phase 1） | — | `docker-compose.yml` | PostgreSQL + Redis 可启动（Milvus 延至 Phase 4） |
| 1.1.6 | 数据库迁移工具集成 | I | `scripts/migration/000001_init_schema.up.sql` | `golang-migrate` 集成 |
| 1.1.7 | Swagger/OpenAPI 集成 | IF | `docs/swagger.go` | `swag` 注解 + `swagger-ui` 端点 |

### 1.2 shared 共享模块 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.2.1 | ID 类型（UUID v4 别名） | D | `internal/domain/shared/types.go` | `type ID string` + `NewID()` |
| 1.2.2 | 统一错误码体系 | D | `internal/domain/shared/errors.go` | `AppError` 结构体 + 全部错误码常量 |
| 1.2.3 | 统一响应封装 | IF | `internal/interface/http/response/response.go` | `Success()` / `Error()` / `PageResult()` |
| 1.2.4 | Context 工具（typed key） | shared | `internal/shared/context/context.go` | `WithUserID` / `UserIDFromContext` |
| 1.2.5 | 请求校验器 | shared | `internal/shared/validator/validator.go` | `validator` 封装 + 自定义规则（UUID 等） |
| 1.2.6 | TransactionManager 接口 | D | `internal/domain/shared/transaction.go` | `InTx(ctx, fn)` 接口 |
| 1.2.7 | GORM TransactionManager 实现 | I | `internal/infrastructure/persistence/postgres/transaction.go` | `InTx` 实现 |
| 1.2.8 | `DBOrTx` 上下文辅助工具 | I | `internal/infrastructure/persistence/postgres/helper.go` | `DBOrTx(ctx, r.db)` 自动判断是否在事务中（spec 5.3 要求） |
| 1.2.9 | EventBus 接口 + 内存实现 | D/I | `internal/domain/shared/event.go`, `internal/infrastructure/event/bus.go` | `Publish` / `Subscribe` + 异步 Worker |

### 1.3 user 用户模块 [1.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.3.1 | User Entity + 值对象 | D | `internal/domain/user/entity.go` | `User` 结构体 + `HashPassword()` / `CheckPassword()` |
| 1.3.2 | UserRepository 接口 | D | `internal/domain/user/repository.go` | `FindByID` / `FindByEmail` / `Create` / `Update` |
| 1.3.3 | UserRepository PostgreSQL 实现 | I | `internal/infrastructure/persistence/postgres/user.go` | GORM 实现 + 软删除 |
| 1.3.4 | DTO 定义 | A | `internal/application/user/dto.go` | `LoginRequest` / `UserResponse` / `TokenResponse` |
| 1.3.5 | UserService | A | `internal/application/user/service.go` | `Login` / `GetMe` / `RefreshToken` |
| 1.3.6 | JWT 工具 | I | `internal/infrastructure/auth/jwt.go` | `GenerateToken` / `ParseToken` + Access/Refresh 双 Token |
| 1.3.7 | Auth Handler | IF | `internal/interface/http/handler/auth.go` | `POST /login`, `POST /refresh`, `GET /me` |
| 1.3.8 | Auth Middleware | IF | `internal/interface/http/middleware/auth.go` | Bearer Token 解析 + `userID` 注入 Context |
| 1.3.9 | 数据库迁移：users 表 | I | `scripts/migration/000001_init_schema.up.sql` | users DDL |
| 1.3.10 | 种子数据脚本（创建初始管理员） | I | `scripts/seed/admin.go` | MVP 阶段通过 CLI 创建首个用户（spec 无 register 端点） |
| 1.3.11 | 单元测试 | A | `internal/application/user/service_test.go` | Login/GetMe/RefreshToken 覆盖率 ≥ 80% |

> **注意**：spec.md 4.2 未定义 register 端点。MVP 阶段通过种子脚本创建首个管理员，后续用户由管理员通过项目成员添加流程邀请。若后续需要自助注册，需 spec 补充 `POST /api/v1/auth/register` 端点。

### 1.4 project 项目模块 [1.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.4.1 | Project + ProjectMember Entity | D | `internal/domain/project/entity.go` | 结构体定义 |
| 1.4.2 | ProjectRepository 接口 | D | `internal/domain/project/repository.go` | `Create` / `FindByID` / `FindByUserID` / `IsMember` / `AddMember` / `CountMembers` |
| 1.4.3 | ProjectRepository PostgreSQL 实现 | I | `internal/infrastructure/persistence/postgres/project.go` | GORM 实现 + 成员关联 |
| 1.4.4 | DTO 定义 | A | `internal/application/project/dto.go` | `CreateProjectReq` / `ProjectResponse`（含成员列表+用例/计划统计） / `AddMemberReq` |
| 1.4.5 | ProjectService | A | `internal/application/project/service.go` | `Create` / `GetByID`（含成员+统计数据） / `ListByUser` / `AddMember` |
| 1.4.6 | Project Handler | IF | `internal/interface/http/handler/project.go` | 4 个端点 |
| 1.4.7 | 项目隔离 Middleware | IF | `internal/interface/http/middleware/project.go` | 校验 `project_id` 参数 + 成员身份 |
| 1.4.8 | 数据库迁移：projects + project_members | I | `scripts/migration/000001_init_schema.up.sql` | DDL |
| 1.4.9 | 单元测试 | A | `internal/application/project/service_test.go` | 覆盖率 ≥ 80% |

### 1.5 依赖注入整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.5.1 | main.go 完整依赖注入 | — | `cmd/server/main.go` | Wire 或手动 DI |
| 1.5.2 | Graceful Shutdown | — | `cmd/server/main.go` | SIGTERM 信号处理 → 停止 HTTP Server → 停止 EventBus Worker → 关闭 DB/Redis |
| 1.5.3 | CORS Middleware | IF | `internal/interface/http/middleware/cors.go` | 白名单配置 |
| 1.5.4 | Logger Middleware | IF | `internal/interface/http/middleware/logger.go` | 请求日志 + 慢请求告警 |
| 1.5.5 | 路由注册（阶段 1 全部端点） | IF | `internal/interface/http/router.go` | auth + project 路由组 |

### 阶段 1 验收

- [ ] `POST /api/v1/auth/login` 返回 JWT
- [ ] `GET /api/v1/auth/me` 正确返回用户信息
- [ ] `POST /api/v1/auth/refresh` 刷新 Token
- [ ] `POST /api/v1/projects` 创建项目
- [ ] `GET /api/v1/projects/{id}` 返回项目详情含成员和统计
- [ ] `POST /api/v1/projects/{id}/members` 添加成员
- [ ] 非成员访问项目返回 403
- [ ] 种子脚本可创建管理员用户
- [ ] Swagger UI 可访问
- [ ] 单元测试全部通过

---

## 阶段 2：测试用例核心（1.5 周）

> 交付目标：完整的用例管理流程——模块树、标签、用例 CRUD、步骤、批量操作、集合

### 2.1 module 模块管理 [1.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.1.1 | Module Entity | D | `internal/domain/testcase/entity.go`（追加） | `Module` 结构体 |
| 2.1.2 | ModuleRepository 接口 | D | `internal/domain/testcase/repository.go`（追加） | `Create` / `Update` / `Delete` / `FindByProject` / `GetTree` |
| 2.1.3 | ModuleRepository 实现 | I | `internal/infrastructure/persistence/postgres/module.go` | 树形查询（递归 CTE 或内存组装） |
| 2.1.4 | DTO | A | `internal/application/testcase/dto.go`（追加） | `CreateModuleReq` / `ModuleTreeResponse` |
| 2.1.5 | ModuleService（含树形组装） | A | `internal/application/testcase/service.go`（追加） | `Create` / `Update` / `Delete` / `GetTree` |
| 2.1.6 | Module Handler | IF | `internal/interface/http/handler/testcase.go` | 4 个端点 |
| 2.1.7 | 数据库迁移：modules 表 | I | `scripts/migration/000002_modules.up.sql` | DDL + 索引 |

### 2.2 tag 标签管理 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.2.1 | Tag Entity | D | `internal/domain/testcase/entity.go`（追加） | `Tag` 结构体 |
| 2.2.2 | TagRepository 接口 + 实现 | D/I | `internal/domain/testcase/repository.go`, `postgres/tag.go` | `FindByProject` / `Create` |
| 2.2.3 | DTO + Service | A | `internal/application/testcase/dto.go`, `service.go` | `Create` / `ListByProject` |
| 2.2.4 | Tag Handler | IF | `internal/interface/http/handler/testcase.go`（追加） | 2 个端点 |
| 2.2.5 | 数据库迁移：tags 表 | I | `scripts/migration/000003_tags.up.sql` | DDL |

### 2.3 testcase 用例管理 [3d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.3.1 | TestCase + Step Entity | D | `internal/domain/testcase/entity.go` | 结构体 + 状态转换规则 `ValidateTransition` |
| 2.3.2 | Priority / Status 值对象 | D | `internal/domain/testcase/valueobject.go` | 枚举 + 校验 |
| 2.3.3 | TestCaseRepository 接口 | D | `internal/domain/testcase/repository.go` | `Create` / `FindByID` / `List` / `Update` / `SoftDelete` / `BatchUpdateStatus` / `BatchDelete` / `BatchMove` |
| 2.3.4 | TestCaseRepository 实现 | I | `internal/infrastructure/persistence/postgres/testcase.go` | GORM 实现 + Preload Steps + 分页 + 标签过滤 + 全文搜索 |
| 2.3.5 | DTO | A | `internal/application/testcase/dto.go` | CRUD DTO + 批量操作 DTO + 过滤参数 |
| 2.3.6 | TestCaseService | A | `internal/application/testcase/service.go` | 全部 CRUD + 状态转换 + 乐观锁 + 批量操作 + 领域事件发布 |
| 2.3.7 | testcase 领域事件 | D | `internal/domain/testcase/events.go` | `TestCaseCreated` / `Updated` / `Deleted` |
| 2.3.8 | Testcase Handler | IF | `internal/interface/http/handler/testcase.go` | 8 个端点（含 batch） |
| 2.3.9 | 数据库迁移：test_cases + test_steps | I | `scripts/migration/000004_testcases.up.sql` | DDL + 全部索引 |
| 2.3.10 | 单元测试 | A | `internal/application/testcase/service_test.go` | CRUD + 状态转换 + 乐观锁冲突 + 批量操作 |
| 2.3.11 | 集成测试 | A | `tests/integration/testcase_test.go` | 关键 API 端点测试：CRUD + 批量操作 + 分页过滤 |

### 2.4 collection 用例集合 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.4.1 | Collection Entity | D | `internal/domain/testcase/entity.go`（追加） | `Collection` 结构体 |
| 2.4.2 | CollectionRepository 接口 + 实现 | D/I | `repository.go`, `postgres/collection.go` | `Create` / `AddCases` / `RemoveCases` / `ListCases` |
| 2.4.3 | DTO + Service | A | `dto.go`, `service.go` | 集合 CRUD + 用例关联 |
| 2.4.4 | Collection Handler | IF | `internal/interface/http/handler/testcase.go`（追加） | 4 个端点 |
| 2.4.5 | 数据库迁移：test_case_collections + collection_cases | I | `scripts/migration/000005_collections.up.sql` | DDL |

### 2.5 路由整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.5.1 | 阶段 2 全部路由注册 | IF | `internal/interface/http/router.go` | modules / tags / testcases / collections 路由组 |
| 2.5.2 | main.go 依赖注入更新 | — | `cmd/server/main.go` | 新模块注入 |

### 阶段 2 验收

- [ ] 模块树 CRUD，支持多级嵌套
- [ ] 标签创建 + 项目标签列表
- [ ] 用例 CRUD，含步骤管理
- [ ] 用例状态转换（draft → ready → archived）
- [ ] 用例列表支持分页、过滤（status/priority/tags/keyword）
- [ ] 批量状态更新、批量删除、批量移动
- [ ] 集合创建 + 添加/移除用例
- [ ] 乐观锁冲突返回 409
- [ ] 单元测试 + 集成测试通过

---

## 阶段 3：计划与执行（1 周）

> 交付目标：创建计划 → 关联用例 → 执行 → 提交结果 → 查看汇总

### 3.1 plan 测试计划 [2.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.1.1 | TestPlan + PlanTestCase Entity | D | `internal/domain/plan/entity.go` | 结构体 + 状态转换规则 |
| 3.1.2 | PlanStatus 值对象 | D | `internal/domain/plan/valueobject.go` | 枚举 + 转换校验 |
| 3.1.3 | TestPlanRepository 接口 | D | `internal/domain/plan/repository.go` | `Create` / `FindByID` / `List` / `Update` / `AddCases` / `RemoveCases` |
| 3.1.4 | TestPlanRepository 实现 | I | `internal/infrastructure/persistence/postgres/plan.go` | GORM + 计划用例关联 |
| 3.1.5 | DTO | A | `internal/application/plan/dto.go` | CRUD + 关联用例 + 状态流转 |
| 3.1.6 | PlanService | A | `internal/application/plan/service.go` | CRUD + 状态流转 + 关联用例 + 创建执行记录 |
| 3.1.7 | Plan Handler | IF | `internal/interface/http/handler/plan.go` | 8 个端点 |
| 3.1.8 | 数据库迁移：test_plans + plan_test_cases | I | `scripts/migration/000006_plans.up.sql` | DDL + 索引 + 部分唯一索引 |
| 3.1.9 | 单元测试 | A | `internal/application/plan/service_test.go` | 状态转换 + 用例关联校验 |

### 3.2 execution 执行记录 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.2.1 | TestExecution + ExecutionResult Entity | D | `internal/domain/execution/entity.go` | 结构体 + 状态定义 |
| 3.2.2 | ExecutionRepository 接口 | D | `internal/domain/execution/repository.go` | `Create` / `FindByID` / `SubmitResult` / `BatchSubmitResults` / `GetSummary` |
| 3.2.3 | ExecutionRepository 实现 | I | `internal/infrastructure/persistence/postgres/execution.go` | GORM + 并发控制（利用部分唯一索引） |
| 3.2.4 | DTO | A | `internal/application/execution/dto.go` | 提交结果 + 执行详情 + 汇总统计 |
| 3.2.5 | ExecutionService | A | `internal/application/execution/service.go` | 创建执行 / 提交结果 / 批量提交 / 暂停恢复 / 完成取消 |
| 3.2.6 | Execution Handler | IF | `internal/interface/http/handler/execution.go` | 3 个端点 |
| 3.2.7 | 数据库迁移：test_executions + execution_results | I | `scripts/migration/000007_executions.up.sql` | DDL + 索引 + 部分唯一索引 |
| 3.2.8 | ALTER TABLE 后置 FK | I | `scripts/migration/000007_executions.up.sql` | `test_plans.current_execution_id` FK |
| 3.2.9 | 单元测试 | A | `internal/application/execution/service_test.go` | 并发执行控制 + 结果去重 |
| 3.2.10 | 集成测试 | A | `tests/integration/execution_test.go` | 创建计划 → 关联用例 → 执行 → 提交结果 → 并发冲突 409 |

### 3.3 路由整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.3.1 | 阶段 3 路由注册 | IF | `internal/interface/http/router.go` | plans / executions 路由组 |
| 3.3.2 | main.go 依赖注入更新 | — | `cmd/server/main.go` | plan + execution 注入 |

### 阶段 3 验收

- [ ] 创建测试计划并关联用例
- [ ] 计划状态流转（draft → active → paused → completed/cancelled）
- [ ] 同一计划并发执行被拒绝（409）
- [ ] 执行中提交单个/批量结果
- [ ] 执行详情含结果汇总（passed/failed/blocked/skipped 计数）
- [ ] 单元测试 + 集成测试通过

---

## 阶段 4：文件与 RAG（1.5 周）

> 交付目标：上传文件 → 自动解析分块 → 向量化 → 向量检索；测试报告

### 4.0 Milvus 环境准备 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.0.1 | Docker Compose 追加 Milvus + etcd + MinIO | — | `docker-compose.yml` | Milvus 2.3+ 可启动 + 健康检查 |
| 4.0.2 | Milvus Go SDK 连通验证 | I | `internal/infrastructure/persistence/milvus/client.go` | Milvus Client 初始化 + Ping |

> Phase 1 的 Docker Compose 仅含 PostgreSQL + Redis，Milvus 延至此阶段引入，节省早期开发资源。

### 4.1 file 文件管理 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.1.1 | File + FileVersion Entity | D | `internal/domain/file/entity.go` | 结构体 + FileType/SourceType 枚举 |
| 4.1.2 | FileRepository 接口 | D | `internal/domain/file/repository.go` | `Create` / `FindByID` / `FindByProject` / `UpdateIndexStatus` / `SoftDelete` |
| 4.1.3 | FileRepository 实现 | I | `internal/infrastructure/persistence/postgres/file.go` | GORM + 版本关联 |
| 4.1.4 | 本地文件存储 | I | `internal/infrastructure/storage/local.go` | `Save` / `GetPath` / `Delete` |
| 4.1.5 | 文件上传校验 | A | `internal/application/file/service.go` | MIME 白名单 + 大小限制 + 文件头检测 |
| 4.1.6 | DTO + FileService | A | `internal/application/file/dto.go`, `service.go` | 上传 + 版本管理 + Figma + 索引状态 + 领域事件 `file.uploaded` |
| 4.1.7 | Figma Client | I | `internal/infrastructure/figma/client.go` | Figma REST API 调用 |
| 4.1.8 | File Handler | IF | `internal/interface/http/handler/file.go` | 7 个端点（upload / figma / list / detail / reindex / index-status / delete） |
| 4.1.9 | 数据库迁移：files + file_versions | I | `scripts/migration/000008_files.up.sql` | DDL + 索引 |
| 4.1.10 | 单元测试 | A | `internal/application/file/service_test.go` | 上传校验 + 版本管理 |

### 4.2 RAG 系统 [3d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.2.1 | DocumentChunk + VectorEmbedding Entity | D | `internal/domain/rag/entity.go` | 结构体 |
| 4.2.2 | ChunkRepository 接口 | D | `internal/domain/rag/repository.go` | `CreateBatch` / `FindByFile` / `DeleteByFile` |
| 4.2.3 | VectorRepository 接口 | D | `internal/domain/rag/repository.go` | `Upsert` / `DeleteByFile` / `Search` |
| 4.2.4 | ChunkRepository PostgreSQL 实现 | I | `internal/infrastructure/persistence/postgres/chunk.go` | 批量插入 |
| 4.2.5 | VectorRepository Milvus 实现 | I | `internal/infrastructure/persistence/milvus/vector.go` | Collection 初始化 + Upsert + Search + Delete |
| 4.2.6 | 文件解析器 | I | `internal/infrastructure/parser/` | `pdf.go` / `docx.go` / `xlsx.go` / `image.go` — 统一 `Parser` 接口 |
| 4.2.7 | 文本分块器 | I | `internal/infrastructure/parser/chunker.go` | `semantic_overlap` 策略实现 |
| 4.2.8 | Embedding Client（独立轻量客户端） | I | `internal/infrastructure/ai/embedding.go` | 调用 OpenAI embedding API 或本地模型。**不依赖 Phase 5 的 Pool/Breaker/Manager**，是独立的 HTTP 客户端 |
| 4.2.9 | DTO + RAG Service | A | `internal/application/rag/dto.go`, `service.go` | `IndexFile` / `Search` / `Reindex` |
| 4.2.10 | 索引补偿 Worker | I | `internal/infrastructure/worker/index_worker.go` | 扫描 index_tasks + 重试 + 补偿定时任务 |
| 4.2.11 | 数据库迁移：document_chunks + vector_embeddings + index_tasks | I | `scripts/migration/000009_rag.up.sql` | DDL + 索引 |
| 4.2.12 | Milvus Collection 初始化 | I | `scripts/migration/milvus_init.go` 或 main.go 启动时 | `heka_chunks` Collection 创建 |
| 4.2.13 | 集成测试 | A | `internal/application/rag/service_test.go` | 上传 → 解析 → 分块 → 向量化 → 检索 |

### 4.3 report 测试报告 [1.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.3.1 | Report DTO | A | `internal/application/report/dto.go` | 计划报告 / 覆盖度 / 趋势 / 缺陷 / 工作量 |
| 4.3.2 | ReportService | A | `internal/application/report/service.go` | 5 种报告查询 + 统计计算。依赖：`PlanRepository` + `ExecutionRepository` + `TestCaseRepository` + `UserRepository`（复用现有 Repository，只做读操作） |
| 4.3.3 | Report Handler | IF | `internal/interface/http/handler/report.go` | 5 个端点 |

### 4.4 路由整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.4.1 | 阶段 4 路由注册 | IF | `internal/interface/http/router.go` | files / reports 路由组 |
| 4.4.2 | main.go 依赖注入更新 | — | `cmd/server/main.go` | file + rag + report + milvus 注入 |

### 阶段 4 验收

- [ ] 文件上传（PDF/Word/Excel/图片）+ 版本管理
- [ ] 文件删除（软删除）
- [ ] 上传后自动触发 RAG 索引（异步）
- [ ] 文件索引状态可查询
- [ ] 向量检索返回相关文档片段
- [ ] Figma 链接导入
- [ ] 测试计划报告（通过率、执行进度）
- [ ] 覆盖度/趋势/缺陷/工作量报告
- [ ] index_tasks 补偿机制运行正常

---

## 阶段 5：AI 功能（1 周）

> 交付目标：AI 生成测试用例 + 智能分析 + SSE 推送

### 5.1 AI 基础设施 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.1.1 | LLMClient 接口 + ChatRequest/ChatResponse | I | `internal/infrastructure/ai/types.go` | 接口定义（放 infrastructure 层，Application 层通过依赖注入使用） |
| 5.1.2 | Claude Provider | I | `internal/infrastructure/ai/claude.go` | Anthropic SDK 封装 |
| 5.1.3 | OpenAI Provider | I | `internal/infrastructure/ai/openai.go` | OpenAI SDK 封装 |
| 5.1.4 | Gemini Provider | I | `internal/infrastructure/ai/gemini.go` | Gemini SDK 封装 |
| 5.1.5 | Ollama Provider | I | `internal/infrastructure/ai/ollama.go` | HTTP 客户端封装 |
| 5.1.6 | Worker Pool | I | `internal/infrastructure/ai/pool.go` | 协程池 + 任务队列 |
| 5.1.7 | Circuit Breaker | I | `internal/infrastructure/ai/breaker.go` | 熔断 + 半开 + 关闭状态机 |
| 5.1.8 | Retry Wrapper | I | `internal/infrastructure/ai/retry.go` | 指数退避 1s → 2s → 4s → 8s（max 30s） |
| 5.1.9 | Timeout Config | I | `internal/infrastructure/ai/timeout.go` | 分级超时（Dial/TLS/Response/Request/Generation） |
| 5.1.10 | MultiProviderClient + Manager | I | `internal/infrastructure/ai/manager.go` | 优先级排序 + 故障转移 + 全 Provider 熔断降级 |
| 5.1.11 | AI 健康检查端点 | IF | `internal/interface/http/handler/monitoring.go` | `GET /api/v1/monitoring/ai` 各 Provider 状态 |
| 5.1.12 | 单元测试 | I | `internal/infrastructure/ai/*_test.go` | 熔断状态机 + 重试策略 + Pool 并发控制 |

### 5.2 AI 应用服务 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.2.1 | Prompt 模板 + SanitizeInput | I | `internal/infrastructure/ai/prompt.go`, `sanitize.go` | 结构化 Prompt + 输入清洗 + 输出校验 |
| 5.2.2 | AI DTO | A | `internal/application/ai/dto.go` | 生成请求/响应 + 分析请求/响应 + 任务状态 |
| 5.2.3 | AIService | A | `internal/application/ai/service.go` | `GenerateTestCases` / `Analyze` — 异步任务创建 + RAG 上下文检索 |
| 5.2.4 | AI Task Worker | I | `internal/infrastructure/worker/ai_worker.go` | 消费 ai_tasks + 调用 LLM + 写入结果 + 领域事件 `ai.task.completed/failed` |
| 5.2.5 | SSE Handler | IF | `internal/interface/http/handler/ai.go` | SSE 推送 + 任务状态查询 |
| 5.2.6 | AI Handler | IF | `internal/interface/http/handler/ai.go` | 4 个端点 |
| 5.2.7 | 数据库迁移：ai_tasks | I | `scripts/migration/000010_ai_tasks.up.sql` | DDL + 索引 |
| 5.2.8 | 集成测试 | A | `internal/application/ai/service_test.go` | Mock LLM + 异步任务流程 |

### 5.3 路由整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.3.1 | 阶段 5 路由注册 | IF | `internal/interface/http/router.go` | ai 路由组 |
| 5.3.2 | main.go 依赖注入更新 | — | `cmd/server/main.go` | ai 全链路注入 |

### 阶段 5 验收

- [ ] AI 用例生成：提交需求 → 返回 task_id → SSE 推送进度 → 查看结果
- [ ] AI 智能分析：变更影响分析返回结果
- [ ] Provider 故障自动转移（模拟一个 Provider 失败）
- [ ] 熔断后自动恢复
- [ ] 队列满时返回 429 AI-RT-001
- [ ] Prompt Injection 防护生效
- [ ] 单元测试 + 集成测试通过

---

## 阶段 6：集成与上线（0.5 周）

> 交付目标：全功能端到端验证 + Docker Compose 一键部署

### 6.1 缓存层 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 6.1.1 | Redis Cache Client | I | `internal/infrastructure/cache/redis.go` | `Get` / `Set` / `Delete` / `DeleteByPattern` |
| 6.1.2 | Cache-Aside 装饰器 | A | 各 Service 增加 cache 层 | 读缓存未命中 → 查 DB → 写缓存；写时删缓存 |
| 6.1.3 | EventBus 订阅缓存失效 | A | `cmd/server/main.go` 注册 | `testcase.created/updated/deleted` → 清缓存 |

### 6.2 速率限制 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 6.2.1 | RateLimiter Middleware | IF | `internal/interface/http/middleware/ratelimit.go` | 令牌桶 / 滑动窗口 |
| 6.2.2 | 端点级别限流配置 | IF | `internal/interface/http/router.go` | 全局 30/s, AI 2/min, 上传 5/min, 登录 5/min |

### 6.3 监控 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 6.3.1 | Health Handler 完善 | IF | `internal/interface/http/handler/monitoring.go` | DB + Redis + Milvus 连通性 |
| 6.3.2 | Prometheus metrics（可选） | IF | `internal/interface/http/middleware/metrics.go` | 请求耗时 / 状态码分布 |

### 6.4 部署整合 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 6.4.1 | Docker Compose 完善 | — | `docker-compose.yml` | 全部服务 + 健康检查 + 启动顺序 |
| 6.4.2 | Dockerfile（多阶段构建） | — | `Dockerfile` | builder → 最终镜像 < 50MB |
| 6.4.3 | Nginx 配置 | — | `nginx.conf` | 反向代理 + HTTPS |
| 6.4.4 | 环境变量映射 | I | `internal/shared/config/config.go` | Docker Compose 旧命名兼容 |
| 6.4.5 | 配置示例文件 | — | `configs/config.example.yaml` | 全部变量 + 注释 |
| 6.4.6 | README | — | `README.md` | 启动指南 + 环境要求 |

### 6.5 端到端测试 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 6.5.1 | E2E 测试脚本 | — | `tests/e2e/` | 全流程：种子用户 → 登录 → 创建项目 → 用例管理 → 计划执行 → 文件上传 → AI 生成 |
| 6.5.2 | 数据库迁移验证 | I | `scripts/migration/` | 从空库到全部表 + 索引 + 约束 |

### 阶段 6 验收

- [ ] Docker Compose 一键启动全部服务
- [ ] 端到端全流程通过
- [ ] 缓存命中率 > 50%（重复查询）
- [ ] 速率限制生效（超限返回 429）
- [ ] 健康检查反映各组件真实状态
- [ ] AI 服务状态端点可用
- [ ] 文件上传 + RAG 索引 + AI 生成完整链路

---

## 数据库迁移文件汇总

| 序号 | 文件名 | 包含表 |
|------|--------|--------|
| 000001 | `init_schema.up.sql` | users, projects, project_members |
| 000002 | `modules.up.sql` | modules |
| 000003 | `tags.up.sql` | tags |
| 000004 | `testcases.up.sql` | test_cases, test_steps |
| 000005 | `collections.up.sql` | test_case_collections, collection_cases |
| 000006 | `plans.up.sql` | test_plans, plan_test_cases |
| 000007 | `executions.up.sql` | test_executions, execution_results + ALTER test_plans FK |
| 000008 | `files.up.sql` | files, file_versions |
| 000009 | `rag.up.sql` | document_chunks, vector_embeddings, index_tasks |
| 000010 | `ai_tasks.up.sql` | ai_tasks |

---

## 工时汇总

| 阶段 | 子任务合计 | 含缓冲日历天 |
|------|-----------|-------------|
| 阶段 1：基础骨架 | 5.5d | 1 周 |
| 阶段 2：测试用例核心 | 7.5d | 1.5 周 |
| 阶段 3：计划与执行 | 5.5d | 1 周 |
| 阶段 4：文件与 RAG | 7.5d | 1.5 周 |
| 阶段 5：AI 功能 | 4.5d | 1 周 |
| 阶段 6：集成与上线 | 3.5d | 0.5 周 |
| **合计** | **34d** | **~7 周** |

---

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Milvus Go SDK 文档较少 | RAG 开发延期 | 提前做 Spike 验证 SDK API（阶段 4.0） |
| AI Provider API 变更 | AI 功能不可用 | 接口抽象 + 多 Provider 故障转移 |
| 文件解析质量 | RAG 检索效果差 | 分阶段支持：先 PDF/Word，后 Excel/图片 |
| execution_results 数据量 | 查询性能下降 | 提前准备按月分区 DDL（阶段 3 注释预留） |
| 前后端联调 | 接口不一致 | 每个 Phase 交付后同步 Swagger 文档 |
| spec.md 缺少 register 端点 | 用户创建流程不完整 | MVP 用种子脚本，后续视需求补充 spec |

---

**文档版本**：v1.1（review 修复）
**最后更新**：2026-05-15
