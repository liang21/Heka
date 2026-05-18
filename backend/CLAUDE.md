# CLAUDE.md - Heka 后端操作手册

## 1. 指令优先级
1. **项目宪法**: 任何行动前必须阅读并对齐 `@constitutions.md`。
2. **环境感知**: 修改前执行 `ls -R` 及 `go mod graph`。

## 2. 项目定位

Heka 是**单体架构**的 AI 测试管理平台，不是微服务。

- 目标用户：20-50 人内部团队，本地部署
- 核心功能：测试用例管理、测试计划与执行、文件管理 + RAG、AI 用例生成与分析
- 一人可维护：清晰分层，严格边界

### 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.21+ |
| ORM | GORM | 1.25+ |
| 数据库 | PostgreSQL | 15+ |
| 日志库 | zerolog | 1.29+ |
| 缓存 | Redis | 7+ |
| 向量库 | Milvus | 2.3+ |
| AI | Claude / OpenAI / Gemini / Ollama | 多 Provider 故障转移 |
| HTTP | chi / 标准库 | - |

## 3. DDD 四层架构（严格遵守）

```
Interface 层 (internal/interface/http/)
    ↓ 调用
Application 层 (internal/application/)
    ↓ 调用
Domain 层 (internal/domain/)
    ↑ 实现
Infrastructure 层 (internal/infrastructure/)
```

**依赖规则**：
- ✅ Interface → Application → Domain
- ✅ Infrastructure → Domain（实现 Repository 接口）
- ❌ Domain 不依赖任何其他层（零第三方库）
- ❌ Application 不依赖 Infrastructure 实现
- ❌ Interface 不跳过 Application 直接调用 Domain

## 4. 领域划分

| 领域 | 目录 | 职责 | 核心实体 |
|------|------|------|----------|
| testcase | `domain/testcase/` | 用例管理 | TestCase, Step, Module |
| plan | `domain/plan/` | 计划管理 | TestPlan, PlanTestCase |
| execution | `domain/execution/` | 执行记录 | TestExecution, ExecutionResult |
| file | `domain/file/` | 文件管理 | File, FileVersion |
| rag | `domain/rag/` | 向量检索 | Chunk, Embedding |
| user | `domain/user/` | 用户管理 | User |
| project | `domain/project/` | 项目管理 | Project, ProjectMember |
| shared | `domain/shared/` | 共享概念 | ID, Error, Timestamp |

## 5. 目录结构

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口，依赖注入，Graceful Shutdown
├── internal/
│   ├── domain/                     # 领域层：纯业务逻辑，零第三方依赖
│   │   ├── testcase/               #   entity.go, valueobject.go, repository.go
│   │   ├── plan/
│   │   ├── execution/
│   │   ├── file/
│   │   ├── rag/
│   │   ├── user/
│   │   ├── project/
│   │   └── shared/                 #   errors.go（错误码）, types.go
│   ├── application/                # 应用层：业务编排
│   │   ├── testcase/               #   service.go, dto.go
│   │   ├── plan/
│   │   ├── execution/
│   │   ├── file/
│   │   ├── rag/
│   │   ├── ai/                     #   AI 编排服务（多 Provider 故障转移）
│   │   ├── user/
│   │   └── project/
│   ├── infrastructure/             # 基础设施层：技术实现
│   │   ├── persistence/
│   │   │   ├── postgres/           #   GORM Repository 实现
│   │   │   └── milvus/             #   VectorRepository 实现
│   │   ├── cache/
│   │   │   └── redis.go
│   │   ├── ai/                     #   AI 客户端（Claude/OpenAI/Gemini/Ollama）
│   │   ├── storage/                #   文件存储
│   │   └── figma/                  #   Figma 集成
│   ├── interface/                  # 接口层：HTTP API
│   │   └── http/
│   │       ├── router.go
│   │       ├── middleware/
│   │       ├── handler/
│   │       └── dto/
│   └── shared/                     # 内部共享
│       ├── config/
│       ├── logger/
│       └── validator/
├── scripts/
│   └── migration/
├── Makefile
├── CLAUDE.md
├── constitutions.md
├── go.mod
└── go.sum
```

### 文件命名约定

| 文件类型 | 命名 | 示例 |
|---------|------|------|
| 实体定义 | `entity.go` | `domain/testcase/entity.go` |
| 值对象 | `valueobject.go` | `domain/testcase/valueobject.go` |
| 仓储接口 | `repository.go` | `domain/testcase/repository.go` |
| 应用服务 | `service.go` | `application/testcase/service.go` |
| DTO | `dto.go` | `application/testcase/dto.go` |
| HTTP Handler | `{resource}.go` | `interface/http/handler/testcase.go` |
| 仓储实现 | `{resource}_repo.go` | `infrastructure/persistence/postgres/testcase_repo.go` |

## 6. Go 编码规范

* **工具链**: Go 1.21+，执行 `gofmt -s` 格式化。
* **导包顺序**: 标准库、第三方库、内部包（分块排列，空行分隔）。
* **错误处理**: 禁止 `_ = err`。必须用 `fmt.Errorf("context: %w", err)` 包装。
* **并发安全**: goroutine 必须通过 `context.Context` 管理生命周期。
* **依赖注入**: 禁止 `init()` 修改全局状态，所有依赖通过构造函数注入。
* **包命名**: 全小写、单数、简短（`testcase`、`file`、`user`）。禁止 `utils`、`common`、`lib`。

## 7. AI 相关规范

* **多 Provider**: `infrastructure/ai/` 实现统一接口，支持 Claude / OpenAI / Gemini / Ollama 故障转移。
* **降级策略**: 主 Provider 失败时自动切换到备选，全部不可用时返回明确错误码（`AI-IE-001`）。
* **RAG 检索**: 文件索引通过 Milvus 向量检索，索引失败不影响文件上传功能。
* **异步任务**: AI 生成采用异步任务模式，前端轮询状态。

## 8. 错误码体系

错误码定义在 `domain/shared/errors.go`，格式：`{领域}-{类型}-{序号}`。

| 类型后缀 | 含义 |
|---------|------|
| `NF` | Not Found |
| `VL` | Validation |
| `CF` | Conflict |
| `AU` | Auth |
| `IE` | Internal Error |

示例：`TC-NF-001`（用例不存在）、`TC-CF-001`（用例版本冲突）、`AI-IE-001`（AI 服务不可用）。

## 9. 开发工作流

* **构建/测试**: 通过 `Makefile`。
    * `make test`: 运行全量测试。
    * `make run`: 启动服务。
* **配置管理**: 严禁硬编码。配置通过 `config.yaml` 加载，修改结构必须同步更新 `config.example.yaml`。
* **Git 规范**: Conventional Commits（`type(scope): subject`）。
* **静态检查**: 提交前通过 `go vet ./...`。
