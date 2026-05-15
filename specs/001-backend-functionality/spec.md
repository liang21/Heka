# Heka 后端功能实现规格

> 编译自 specs/ 目录下 8 份设计文档，面向实现的统一规格
> 版本：v1.2（二次审读精修）
> 日期：2026-05-15

---

## 1. 概述

### 1.1 项目定位

Heka 是面向 20-50 人内部团队的简化版 AI 测试管理平台，基于 MeterSphere 理念大幅简化，聚焦核心测试管理并深度融合 AI 能力。

**核心价值**：简化测试管理 | AI 赋能 | 本地部署 | 一人可维护

### 1.2 技术栈

| 层级 | 技术选型 |
|------|----------|
| 后端 | Go 1.21+ + GORM |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| 向量库 | Milvus 2.3+ |
| AI Provider | Claude / OpenAI / Gemini / Ollama（多 Provider 故障转移） |

### 1.3 第一阶段功能范围

| 模块 | 功能 |
|------|------|
| 用户管理 | 登录认证（JWT）、项目成员管理 |
| 项目管理 | 多项目隔离、项目切换 |
| 测试用例 | CRUD、模块目录、步骤管理、状态管理、优先级、标签、用例集合 |
| 测试计划 | 计划管理、关联用例、执行分配、迭代管理 |
| 执行记录 | 执行结果记录、缺陷关联、执行历史 |
| 测试报告 | 计划报告、覆盖度、执行趋势、缺陷分布、工作量 |
| 文件管理 | 上传（PDF/Word/Excel/图片/Figma）、版本管理 |
| RAG 系统 | 文件解析、文本分块、向量化、向量检索 |
| AI 用例生成 | 需求解析、用例生成、批量创建 |
| AI 智能分析 | 变更分析、回归推荐、影响分析 |

**不包含**：接口测试、UI 测试、性能测试、完整缺陷管理、三层租户、资源池、DevOps 集成、插件体系

---

## 2. 领域模块清单

| 模块 | 核心实体 | 值对象 | 仓储接口 | 领域服务 |
|------|---------|--------|---------|---------|
| `user` | User | — | UserRepository | — |
| `project` | Project, ProjectMember | — | ProjectRepository | — |
| `testcase` | TestCase, Step, Module, Tag, Collection | Priority, Status | TestCaseRepository, ModuleRepository, TagRepository, CollectionRepository | ValidateTransition, CalculatePriority |
| `plan` | TestPlan, PlanTestCase | PlanStatus | TestPlanRepository | — |
| `execution` | TestExecution, ExecutionResult | ExecutionStatus | ExecutionRepository | — |
| `file` | File, FileVersion | FileType, SourceType | FileRepository | — |
| `rag` | DocumentChunk, VectorEmbedding | — | VectorRepository, ChunkRepository | — |
| `shared` | — | ID, Timestamp | TransactionManager, EventBus | — |

---

## 3. 数据模型

### 3.1 用户相关

```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 项目表
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 项目成员表
CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, user_id)
);
```

### 3.2 测试用例

```sql
-- 业务模块表
CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES modules(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, parent_id, name)
);

CREATE INDEX idx_modules_project ON modules(project_id);
CREATE INDEX idx_modules_parent ON modules(parent_id);

-- 测试用例表
CREATE TABLE test_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    module_id UUID REFERENCES modules(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'ready', 'archived')),
    priority INTEGER NOT NULL DEFAULT 1,
    tags TEXT[],
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_test_cases_project ON test_cases(project_id);
CREATE INDEX idx_test_cases_module ON test_cases(module_id);
CREATE INDEX idx_test_cases_status ON test_cases(status);
CREATE INDEX idx_test_cases_priority ON test_cases(priority);
CREATE INDEX idx_test_cases_tags ON test_cases USING GIN(tags);
CREATE INDEX idx_test_cases_project_status ON test_cases(project_id, status);
CREATE INDEX idx_test_cases_module_status ON test_cases(module_id, status);
CREATE INDEX idx_test_cases_fulltext ON test_cases USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));
CREATE INDEX idx_test_cases_active ON test_cases(project_id, created_at) WHERE deleted_at IS NULL;

-- 测试步骤表（独立表存储，非 JSON 序列化）
-- 设计决策：steps 使用独立表而非 JSON 列，原因：
-- 1. 步骤有独立 ID 和 UNIQUE(test_case_id, number) 约束
-- 2. 支持 GORM Preload 预加载
-- 3. 步骤顺序由数据库层保证
CREATE TABLE test_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    action TEXT NOT NULL,
    expected TEXT NOT NULL,
    UNIQUE (test_case_id, number)
);

CREATE INDEX idx_test_steps_case ON test_steps(test_case_id);

-- 用例集合表
CREATE TABLE test_case_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE collection_cases (
    collection_id UUID NOT NULL REFERENCES test_case_collections(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (collection_id, test_case_id)
);

-- 标签表
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, name)
);

CREATE INDEX idx_tags_project ON tags(project_id);
```

#### 用例状态转换规则

```
draft    → ready, archived
ready    → archived, draft
archived → ready
```

### 3.3 测试计划与执行

```sql
-- 测试计划表
CREATE TABLE test_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'cancelled')),
    started_at TIMESTAMP,
    paused_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_execution_id UUID, -- 逻辑外键指向 test_executions(id)，DDL 后置添加
    deleted_at TIMESTAMP
);

CREATE INDEX idx_test_plans_project ON test_plans(project_id);
CREATE INDEX idx_test_plans_status ON test_plans(status);

-- 测试计划用例关联表
CREATE TABLE plan_test_cases (
    plan_id UUID NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES users(id),
    order_index INTEGER NOT NULL,
    PRIMARY KEY (plan_id, test_case_id)
);

-- 执行记录表
CREATE TABLE test_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
    name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'paused', 'completed', 'cancelled')),
    executor_id UUID NOT NULL REFERENCES users(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paused_at TIMESTAMP,
    completed_at TIMESTAMP,
    notes TEXT
);

CREATE INDEX idx_executions_plan ON test_executions(plan_id);
CREATE INDEX idx_executions_executor ON test_executions(executor_id);
CREATE INDEX idx_executions_status ON test_executions(status);
CREATE INDEX idx_executions_plan_status ON test_executions(plan_id, status);

-- 执行结果表（建议按月分区）
CREATE TABLE execution_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES test_executions(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id),
    executor_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed', 'blocked', 'skipped')),
    bug_id VARCHAR(255),
    bug_url VARCHAR(500),
    notes TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (execution_id, test_case_id)
);

CREATE INDEX idx_execution_results_execution ON execution_results(execution_id);
CREATE INDEX idx_execution_results_status ON execution_results(status);
```

#### 计划状态转换

```
draft     → active, cancelled
active    → paused, completed, cancelled（需至少关联一个用例）
paused    → active, cancelled
completed → （终态）
cancelled → （终态）
```

#### 并发执行控制

同一测试计划同时只能有一个 `in_progress` 状态的执行。数据库层保障：

```sql
-- 部分唯一索引：每个 plan 最多一条 in_progress 记录
CREATE UNIQUE INDEX idx_executions_single_active 
ON test_executions(plan_id) 
WHERE status = 'in_progress';
```

### 3.4 文件管理

```sql
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL,
    size BIGINT NOT NULL,
    path VARCHAR(1000) NOT NULL,
    source_type VARCHAR(20) NOT NULL,
    source_url TEXT,
    content_preview TEXT,
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE,
    index_status VARCHAR(20) DEFAULT 'pending' CHECK (index_status IN ('pending', 'processing', 'completed', 'failed')),
    index_error TEXT,
    indexed_at TIMESTAMP,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_files_project ON files(project_id);
CREATE INDEX idx_files_type ON files(type);
CREATE INDEX idx_files_indexed ON files(is_indexed);
CREATE INDEX idx_files_index_status ON files(index_status);

CREATE TABLE file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    path VARCHAR(1000) NOT NULL,
    size BIGINT NOT NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_id, version)
);
```

### 3.5 RAG 相关

```sql
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    index INTEGER NOT NULL,
    tokens INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (file_id, index)
);

CREATE INDEX idx_chunks_file ON document_chunks(file_id);

CREATE TABLE vector_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id UUID NOT NULL REFERENCES document_chunks(id) ON DELETE CASCADE,
    model VARCHAR(100) NOT NULL,
    dimension INTEGER NOT NULL,
    milvus_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_embeddings_chunk ON vector_embeddings(chunk_id);
```

#### RAG 分块配置

```go
type ChunkConfig struct {
    MaxTokens    int // 最大 token 数，默认 500
    Overlap      int // 重叠 token 数，默认 50
    MinChunkSize int // 最小块大小，默认 100 tokens
}
```

分块策略：`semantic_overlap`（语义重叠分块）
- 按段落分块（优先保留段落完整性）
- 段落超过 max_tokens 按句子分割
- 块之间保留 overlap_tokens 重叠
- 丢弃小于 min_chunk_size 的块（最后一块除外）

#### 文件解析技术栈（Go 生态）

| 文件类型 | Go 库 | 说明 |
|----------|-------|------|
| PDF | `ledongthuc/pdfgof` 或 `unidoc/unipdf` | 文本提取 |
| Word (.docx) | `nguyenthenguyen/docx` | 文本提取 |
| Excel (.xlsx) | `excelize/v2` | 文本提取 |
| 图片 | `leadgo/ocr` 或调用 Tesseract CLI | OCR 文字识别 |
| Figma | Figma REST API | API 提取 |

### 3.6 异步任务表

```sql
-- AI 异步任务表（支持 AI 生成、分析等耗时操作）
CREATE TABLE ai_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,             -- generate_testcases, analyze
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    progress_current INTEGER NOT NULL DEFAULT 0,
    progress_total INTEGER NOT NULL DEFAULT 0,
    input JSONB NOT NULL,                  -- 任务输入参数
    result JSONB,                          -- 任务结果（完成后填充）
    error TEXT,                            -- 失败原因
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_ai_tasks_project ON ai_tasks(project_id);
CREATE INDEX idx_ai_tasks_status ON ai_tasks(status);
CREATE INDEX idx_ai_tasks_created_by ON ai_tasks(created_by);

-- Milvus 索引补偿任务表（PostgreSQL ↔ Milvus 最终一致性）
CREATE TABLE index_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_index_tasks_status ON index_tasks(status);
CREATE INDEX idx_index_tasks_file ON index_tasks(file_id);
```

### 3.7 软删除策略

| 表 | 策略 | 归档周期 |
|----|------|---------|
| users | 软删除 | 永久保留 |
| projects | 软删除 | 1 年后归档 |
| modules | 硬删除 | — |
| test_cases | 软删除 | 6 个月后归档 |
| test_plans | 软删除 | 1 年后归档 |
| files | 软删除 + 异步清理 | 30 天后清理物理文件 |
| execution_results | 分区归档 | 6 个月 |

---

## 4. API 规格

### 4.1 通用规范

- **版本前缀**：`/api/v1/...`
- **认证**：`Authorization: Bearer {token}`（除登录接口外全部需要）
- **成功响应**：`{"code": 0, "data": {...}, "message": "success"}`
- **错误响应**：`{"code": "TC-NF-001", "message": "测试用例不存在"}`
- **ID 格式**：UUID v4
- **分页**：`page`（min=1, default=1）、`page_size`（min=1, max=100, default=20）
- **排序**：`sort_by`（created_at/updated_at/title/priority/status）、`sort_desc`（bool）

### 4.2 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录，返回 JWT token |
| POST | `/api/v1/auth/refresh` | 刷新 Token（Refresh Token 有效期 7 天） |
| GET | `/api/v1/auth/me` | 获取当前用户信息 |

### 4.3 项目

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/projects` | 创建项目 |
| GET | `/api/v1/projects` | 获取项目列表 |
| GET | `/api/v1/projects/{id}` | 获取项目详情（含成员、统计） |
| POST | `/api/v1/projects/{id}/members` | 添加项目成员 |

### 4.4 模块

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/modules` | 创建模块 |
| GET | `/api/v1/modules?project_id={id}` | 获取模块树 |
| PUT | `/api/v1/modules/{id}` | 更新模块 |
| DELETE | `/api/v1/modules/{id}` | 删除模块（用例解除关联） |

### 4.5 测试用例

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/testcases` | 创建测试用例 |
| GET | `/api/v1/testcases?project_id={id}&page=1&page_size=20` | 获取用例列表（支持 status/priority/keyword/tags 过滤） |
| GET | `/api/v1/testcases/{id}` | 获取用例详情（含步骤） |
| PUT | `/api/v1/testcases/{id}` | 更新测试用例 |
| DELETE | `/api/v1/testcases/{id}` | 删除测试用例 |
| PUT | `/api/v1/testcases/batch/status` | 批量更新状态 |
| DELETE | `/api/v1/testcases/batch` | 批量删除 |
| PUT | `/api/v1/testcases/batch/move` | 批量移动到模块 |

### 4.6 标签

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/tags?project_id={id}` | 获取项目标签列表 |
| POST | `/api/v1/tags` | 创建标签 |

### 4.7 用例集合

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/collections` | 创建用例集合 |
| POST | `/api/v1/collections/{id}/cases` | 添加用例到集合 |
| GET | `/api/v1/collections/{id}/cases?page=1&page_size=20` | 获取集合用例列表 |
| DELETE | `/api/v1/collections/{id}/cases` | 从集合移除用例 |

### 4.8 测试计划

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/testplans` | 创建测试计划（含关联用例） |
| GET | `/api/v1/testplans?project_id={id}&status=active` | 获取计划列表 |
| GET | `/api/v1/testplans/{id}` | 获取计划详情（含进度、当前执行） |
| POST | `/api/v1/testplans/{id}/start` | 开始执行（创建执行记录） |
| POST | `/api/v1/testplans/{id}/pause` | 暂停执行 |
| POST | `/api/v1/testplans/{id}/resume` | 恢复执行 |
| POST | `/api/v1/testplans/{id}/complete` | 完成计划 |
| POST | `/api/v1/testplans/{id}/cancel` | 取消计划 |

### 4.9 执行记录

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/executions/{execution_id}/results` | 提交单个执行结果 |
| POST | `/api/v1/executions/{execution_id}/results/batch` | 批量提交执行结果 |
| GET | `/api/v1/executions/{id}` | 获取执行详情（含结果汇总） |

### 4.10 文件管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/files/upload` | 上传文件（multipart/form-data） |
| POST | `/api/v1/files/figma` | 上传 Figma 链接 |
| GET | `/api/v1/files?project_id={id}` | 获取文件列表 |
| GET | `/api/v1/files/{id}` | 获取文件详情 |
| POST | `/api/v1/files/{id}/reindex` | 重新索引文件 |
| GET | `/api/v1/files/{id}/index-status` | 获取索引状态 |
| DELETE | `/api/v1/files/{id}` | 删除文件 |

### 4.11 AI 功能

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/ai/generate-testcases` | AI 生成测试用例（异步，返回 task_id） |
| GET | `/api/v1/ai/tasks/{task_id}` | 获取 AI 任务状态 |
| GET | `/api/v1/ai/tasks/{task_id}/events` | SSE 推送任务进度 |
| POST | `/api/v1/ai/analyze` | AI 智能分析（变更影响分析） |

### 4.12 报告

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/reports/plan/{plan_id}` | 测试计划报告 |
| GET | `/api/v1/reports/coverage?project_id={id}` | 用例覆盖度报告 |
| GET | `/api/v1/reports/trend?project_id={id}&days=30` | 执行趋势报告 |
| GET | `/api/v1/reports/bugs?project_id={id}` | 缺陷分布报告 |
| GET | `/api/v1/reports/workload?user_id={id}` | 个人工作量报告 |

### 4.13 监控

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 系统健康检查 |
| GET | `/api/v1/monitoring/ai` | AI 服务状态 |

---

## 5. 分层架构实现指南

### 5.1 依赖方向

```
Interface → Application → Domain ← Infrastructure
```

- Domain 不依赖任何其他层
- Application 依赖 Domain（通过接口）
- Infrastructure 实现 Domain 定义的接口
- Interface 只调用 Application 层

### 5.2 目录结构

```
heka-backend/
├── cmd/server/main.go                          # 入口，依赖注入
├── internal/
│   ├── domain/
│   │   ├── testcase/entity.go                  # TestCase, Step
│   │   ├── testcase/valueobject.go             # Priority, Status
│   │   ├── testcase/repository.go              # Repository 接口
│   │   ├── testcase/service.go                 # 领域服务
│   │   ├── testcase/events.go                  # 领域事件
│   │   ├── plan/entity.go
│   │   ├── plan/repository.go
│   │   ├── execution/entity.go
│   │   ├── execution/repository.go
│   │   ├── file/entity.go
│   │   ├── file/repository.go
│   │   ├── rag/entity.go
│   │   ├── rag/repository.go
│   │   ├── user/entity.go
│   │   ├── user/repository.go
│   │   ├── project/entity.go
│   │   ├── project/repository.go
│   │   └── shared/errors.go, types.go, transaction.go, event.go
│   ├── application/
│   │   ├── testcase/service.go, dto.go
│   │   ├── plan/service.go, dto.go
│   │   ├── execution/service.go, dto.go
│   │   ├── file/service.go, dto.go
│   │   ├── rag/service.go, dto.go
│   │   ├── ai/service.go, dto.go
│   │   ├── user/service.go, dto.go
│   │   └── project/service.go, dto.go
│   ├── infrastructure/
│   │   ├── persistence/postgres/               # PostgreSQL Repository 实现
│   │   ├── persistence/milvus/                 # Milvus VectorRepository 实现
│   │   ├── cache/redis.go
│   │   ├── ai/                                 # AI 客户端（pool/circuit/retry/timeout/manager）
│   │   ├── storage/local.go
│   │   └── figma/client.go
│   ├── interface/http/
│   │   ├── router.go
│   │   ├── middleware/auth.go, cors.go, logger.go
│   │   ├── handler/testcase.go, plan.go, execution.go, file.go, ai.go, user.go, project.go
│   │   └── response/response.go
│   └── shared/config/config.go, logger/logger.go, validator/validator.go
├── scripts/migration/                          # 数据库迁移
└── configs/config.example.yaml
```

### 5.3 事务管理

- Application 层通过 `TransactionManager.InTx()` 管理事务边界
- Repository 使用 `DBOrTx(ctx, r.db)` 自动判断是否在事务中
- 禁止 Repository 自行 Begin/Commit

### 5.4 Context 传递

- 使用 `sharedctx.WithUserID(ctx, id)` / `sharedctx.UserIDFromContext(ctx)` 传递用户身份
- 禁止字符串 key：`ctx.Value("user_id")`
- Domain 层不读取 Context 业务值

### 5.5 领域事件

#### 接口定义

```go
// internal/domain/shared/event.go
type Event interface {
    EventName() string
    OccurredAt() time.Time
}

type EventHandler func(ctx context.Context, event Event) error

type EventBus interface {
    Publish(ctx context.Context, events ...Event) error
    Subscribe(eventName string, handler EventHandler)
}
```

#### 实现方式

异步 EventBus（内存队列），Worker 数量可配置。事件发布不阻塞主流程。

#### 事件清单

| 事件名 | 触发时机 | 处理动作 |
|--------|---------|---------|
| `testcase.created` | 用例创建 | 清除项目用例列表缓存 |
| `testcase.updated` | 用例更新 | 清除用例详情缓存 |
| `testcase.deleted` | 用例删除 | 清除项目用例列表缓存 |
| `file.uploaded` | 文件上传完成 | 触发 RAG 索引（写入 index_tasks） |
| `ai.task.completed` | AI 任务完成 | SSE 通知前端 |
| `ai.task.failed` | AI 任务失败 | SSE 通知前端 |

#### 跨数据库事务补偿

Milvus 不支持 ACID 事务，采用最终一致性：

1. 先写入 PostgreSQL（事务保证）
2. 创建 `index_tasks` 记录（status=pending）
3. 后台 Worker 异步处理 Milvus 索引
4. 失败时 `retry_count++`，超过 `max_retries` 标记为 failed
5. 定时补偿（每 5 分钟扫描 status=pending 且 created_at > 10min 的记录）

---

## 6. AI 调用层

### 6.1 架构

```
请求 → Manager（重试包装）→ MultiProviderClient（按优先级排序）→ Provider Client（超时控制）→ Pool（并发控制）
                                          ↓ 失败
                                    Breaker（熔断判断）→ 跳到下一个 Provider
```

### 6.2 关键接口

```go
type LLMClient interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    GenerateText(ctx context.Context, prompt string) (string, error)
}

type Manager struct { ... }
func (m *Manager) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
func (m *Manager) GenerateText(ctx context.Context, prompt string) (string, error)
func (m *Manager) StreamChat(ctx context.Context, req ChatRequest, callback func(chunk string)) error
```

### 6.3 配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| pool_max_workers | 5 | 并发 Worker 数 |
| pool_queue_size | 50 | 任务队列长度 |
| retry_max_attempts | 3 | 最大重试次数 |
| retry_base_delay | 1s | 重试基础延迟 |
| retry_max_delay | 30s | 重试最大延迟 |
| timeout_dial | 10s | 连接超时 |
| timeout_tls_handshake | 5s | TLS 握手超时 |
| timeout_response_header | 30s | 响应头超时 |
| timeout_request | 60s | 请求总超时 |
| timeout_generation | 55s | 生成超时 |

### 6.4 Provider 优先级

1. Claude (Priority 1)
2. OpenAI (Priority 2)
3. Gemini (Priority 3)
4. Ollama (Priority 4, 本地兜底)

### 6.5 降级策略

| 条件 | 动作 |
|------|------|
| 单 Provider 连续 5 次失败 | 熔断该 Provider，30s 后半开 |
| 所有 Provider 熔断 | 返回缓存或模板结果 |
| 队列超过 50 | 拒绝新请求（AI-RT-001） |

---

## 7. 安全规范

### 7.1 JWT 认证

- Access Token 有效期 24h
- 密码使用 bcrypt（cost=12）
- Token 使用 HS256 签名
- 登录失败不区分"用户不存在"和"密码错误"

### 7.2 项目隔离

Middleware 校验用户是否为项目成员，非成员返回 403。

### 7.3 输入校验

- 服务端使用 `validator` 库校验结构体
- 文件上传：白名单 MIME 类型 + 大小限制 100MB + 读取文件头判断真实类型
- UUID 参数必须验证格式

### 7.4 AI 安全

- 用户输入经过 `SanitizeInput` 清洗（移除 Prompt Injection 模式）
- AI 输出经过 `ValidateAIOutput` 校验
- 使用结构化 Prompt 模板，用户输入用 `<document_content>` 标签包裹

### 7.5 速率限制

| 端点 | 限制 |
|------|------|
| 全局 API | 30 req/s |
| AI 生成 | 2 req/min |
| 文件上传 | 5 req/min |
| 登录 | 5 req/min |

### 7.6 CORS

生产环境限制允许的域名，开发环境允许 `localhost:3000`。

---

## 8. 环境变量

### 必需

| 变量 | 说明 |
|------|------|
| `HEKA_DB_HOST` | PostgreSQL 主机 |
| `HEKA_DB_PORT` | PostgreSQL 端口 |
| `HEKA_DB_NAME` | 数据库名 |
| `HEKA_DB_USER` | 数据库用户 |
| `HEKA_DB_PASSWORD` | 数据库密码（敏感） |
| `HEKA_JWT_SECRET` | JWT 签名密钥（敏感，>= 32 字符） |

### 可选（有默认值）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HEKA_DB_MAX_OPEN_CONNS` | 50 | 最大开放连接数 |
| `HEKA_DB_MAX_IDLE_CONNS` | 10 | 最大空闲连接数 |
| `HEKA_DB_CONN_MAX_LIFETIME` | 1h | 连接最大存活时间 |
| `HEKA_JWT_TTL` | 24h | Access Token 有效期 |
| `HEKA_REDIS_HOST` | localhost | Redis 主机 |
| `HEKA_REDIS_PORT` | 6379 | Redis 端口 |
| `HEKA_REDIS_PASSWORD` | "" | Redis 密码（敏感） |
| `HEKA_REDIS_DB` | 0 | Redis 数据库编号 |
| `HEKA_MILVUS_HOST` | localhost | Milvus 主机 |
| `HEKA_MILVUS_PORT` | 19530 | Milvus 端口 |
| `HEKA_MILVUS_COLLECTION` | heka_chunks | Milvus Collection 名称 |
| `HEKA_EMBEDDING_DIMENSION` | 1536 | 向量维度 |
| `HEKA_SERVER_PORT` | 8080 | HTTP 端口 |
| `HEKA_SERVER_MODE` | release | 运行模式 |
| `HEKA_LOG_LEVEL` | info | 日志级别 |
| `HEKA_AI_CLAUDE_API_KEY` | — | Claude API Key（敏感） |
| `HEKA_AI_OPENAI_API_KEY` | — | OpenAI API Key（敏感） |
| `HEKA_AI_GEMINI_API_KEY` | — | Gemini API Key（敏感） |
| `HEKA_AI_OLLAMA_HOST` | localhost:11434 | Ollama 地址 |
| `HEKA_AI_POOL_WORKERS` | 5 | AI 并发数 |
| `HEKA_AI_POOL_QUEUE_SIZE` | 50 | AI 任务队列长度 |
| `HEKA_AI_TIMEOUT_REQUEST` | 60s | AI 请求超时 |
| `HEKA_AI_RETRY_MAX_ATTEMPTS` | 3 | AI 调用重试次数 |
| `HEKA_UPLOAD_PATH` | /var/heka/uploads | 文件存储路径 |
| `HEKA_UPLOAD_MAX_SIZE` | 104857600 | 最大文件 100MB |
| `HEKA_FIGMA_ACCESS_TOKEN` | — | Figma Token（敏感） |

---

## 9. 错误码

### 格式

`{DOMAIN}-{TYPE}-{NUMBER}`

DOMAIN: AUTH / TC / TP / EX / FILE / RAG / AI / PROJ / USER / SYS
TYPE: NF(Not Found) / VL(Validation) / AU(Auth) / CF(Conflict) / IE(Internal Error) / TE(Timeout) / RT(Rate Limit)

### 错误码表

| 错误码 | HTTP | 含义 |
|--------|------|------|
| **AUTH** | | |
| AUTH-AU-001 | 401 | 未认证，缺少或无效 Token |
| AUTH-AU-002 | 401 | Token 已过期 |
| AUTH-AU-003 | 403 | 无权限访问该项目 |
| AUTH-VL-001 | 400 | 邮箱或密码错误 |
| **TC** | | |
| TC-NF-001 | 404 | 测试用例不存在 |
| TC-VL-001 | 400 | 标题不能为空 |
| TC-VL-002 | 400 | 至少需要一个步骤 |
| TC-VL-003 | 400 | 无效的优先级值 |
| TC-VL-004 | 400 | 无效的状态转换 |
| TC-CF-001 | 409 | 版本冲突，数据已被修改 |
| TC-CF-002 | 409 | 用例标题重复 |
| **TP** | | |
| TP-NF-001 | 404 | 测试计划不存在 |
| TP-VL-001 | 400 | 计划必须关联至少一个用例 |
| TP-VL-002 | 400 | 无效的计划状态转换 |
| TP-CF-001 | 409 | 已有执行中的测试 |
| **EX** | | |
| EX-NF-001 | 404 | 执行记录不存在 |
| EX-VL-001 | 400 | 无效的执行状态 |
| EX-CF-001 | 409 | 该用例已提交执行结果 |
| **FILE** | | |
| FILE-NF-001 | 404 | 文件不存在 |
| FILE-VL-001 | 400 | 不支持的文件类型 |
| FILE-VL-002 | 400 | 文件大小超限 |
| FILE-VL-003 | 400 | 文件正在索引中 |
| **AI** | | |
| AI-IE-001 | 503 | AI 服务不可用（所有 Provider 故障） |
| AI-TE-001 | 504 | AI 调用超时 |
| AI-RT-001 | 429 | AI 请求排队已满 |
| AI-IE-002 | 500 | AI 响应解析失败 |
| **PROJ** | | |
| PROJ-NF-001 | 404 | 项目不存在 |
| PROJ-CF-001 | 409 | 项目名已存在 |
| **USER** | | |
| USER-NF-001 | 404 | 用户不存在 |
| USER-VL-001 | 400 | 邮箱格式无效 |
| USER-CF-001 | 409 | 邮箱已注册 |
| **RAG** | | |
| RAG-NF-001 | 404 | 文档块不存在 |
| RAG-IE-001 | 500 | 向量化失败 |
| RAG-IE-002 | 500 | 向量检索失败 |
| **SYS** | | |
| SYS-IE-001 | 500 | 数据库错误 |
| SYS-IE-002 | 500 | 缓存服务错误 |
| SYS-IE-003 | 503 | 向量数据库不可用 |
| SYS-VL-001 | 400 | 参数验证失败 |
| SYS-RT-001 | 429 | 请求频率超限 |

---

## 10. 缓存策略

### Cache-Aside 模式

读：先查缓存 → 未命中查数据库 → 写入缓存
写：先写数据库 → 删除缓存

### 缓存 Key 定义

| Key 模式 | TTL | 说明 |
|----------|-----|------|
| `user:{id}` | 1h | 用户信息 |
| `project:{id}` | 30min | 项目信息 |
| `project:{id}:members` | 30min | 项目成员列表 |
| `project:{id}:testcases:{page}:{filters_hash}` | 5min | 用例列表 |
| `testcase:{id}` | 10min | 用例详情 |
| `project:{id}:modules` | 10min | 模块树 |
| `project:{id}:stats` | 10min | 统计数据 |
| `ai:task:{id}` | 1h（完成后） | AI 任务状态 |
| `file:{id}:index` | 5min | 文件索引状态 |
| `ai:cache:{sha256(prompt)}` | 1h | AI 生成结果缓存 |
| `rag:search:{project_id}:{query_hash}` | 10min | RAG 查询结果 |

---

## 11. 性能基线

### API 响应时间目标

| 端点 | P50 | P95 | P99 |
|------|-----|-----|-----|
| GET /api/v1/testcases | < 100ms | < 300ms | < 500ms |
| GET /api/v1/testcases/:id | < 50ms | < 100ms | < 200ms |
| POST /api/v1/testcases | < 200ms | < 500ms | < 1s |
| PUT /api/v1/testcases/:id | < 200ms | < 500ms | < 1s |
| POST /api/v1/files/upload | < 1s | < 5s | < 10s |
| POST /api/v1/ai/generate-testcases | < 15s | < 30s | < 60s |
| POST /api/v1/ai/analyze | < 10s | < 20s | < 45s |
| GET /api/v1/reports/* | < 500ms | < 2s | < 5s |

### 系统资源基线

| 指标 | 正常值 | 告警阈值 |
|------|--------|---------|
| CPU 使用率 | < 40% | > 70% |
| 内存使用率 | < 60% | > 80% |
| PostgreSQL 连接数 | < 20 | > 40 |
| Redis 内存 | < 200MB | > 400MB |
| Milvus 查询延迟 | < 50ms | > 200ms |
| 磁盘使用率 | < 50% | > 75% |

### 资源要求

- CPU：4 核以上
- 内存：8GB 以上（PostgreSQL 2GB + Milvus 2GB + Redis 0.5GB + Backend 1GB + 其他 2.5GB）
- 磁盘：100GB 以上 SSD

---

## 12. 部署配置

### 12.1 端口分配

| 组件 | 端口 | 对外暴露 |
|------|------|---------|
| Nginx | 80, 443 | 是 |
| Backend | 8080 | 否（通过 Nginx） |
| Frontend | 3000 | 否（通过 Nginx） |
| PostgreSQL | 5432 | 否 |
| Redis | 6379 | 否 |
| Milvus | 19530 | 否 |
| etcd | 2379 | 否 |
| MinIO | 9000 | 否 |

### 12.2 启动顺序

1. 基础设施：etcd → MinIO → Milvus → PostgreSQL → Redis
2. 应用：Backend → Frontend
3. 反向代理：Nginx（可选）

### 12.3 数据库迁移

使用 `golang-migrate`，迁移文件放在 `scripts/migration/`：
```
000001_init_schema.up.sql / .down.sql
```

### 12.4 Milvus Collection

#### Collection Schema

```
Collection: heka_chunks
┌──────────────────────────────────────────────────────────────────┐
| 字段        | 类型               | 说明                        |
|-------------|--------------------|-----------------------------|
| id          | VARCHAR(36) PK     | chunk UUID                  |
| chunk_id    | VARCHAR(36)        | document_chunks 表 ID       |
| project_id  | VARCHAR(36)        | 项目 ID（用于过滤）          |
| content     | VARCHAR(65535)     | 文本内容                     |
| embedding   | FLOAT_VECTOR(1536) | 向量嵌入                     |
└──────────────────────────────────────────────────────────────────┘
```

#### 索引策略

| 数据量 | 索引类型 | 参数 |
|--------|---------|------|
| < 10万 | IVF_FLAT | nlist=64 |
| 10万-100万 | IVF_PQ | nlist=128, m=8 |
| > 100万 | HNSW | M=16 |

- 向量维度：1536（text-embedding-ada-002，可通过 `HEKA_EMBEDDING_DIMENSION` 配置）
- 跨数据库事务：最终一致性（见 5.5 领域事件 → 跨数据库事务补偿）

### 12.5 Docker Compose 环境变量映射

为兼容现有 `docker-compose.yml`，Backend 容器同时支持旧命名：

| Docker Compose 变量 | 映射到 |
|---------------------|--------|
| `DATABASE_HOST` | `HEKA_DB_HOST` |
| `DATABASE_PORT` | `HEKA_DB_PORT` |
| `DATABASE_NAME` | `HEKA_DB_NAME` |
| `DATABASE_USER` | `HEKA_DB_USER` |
| `DATABASE_PASSWORD` | `HEKA_DB_PASSWORD` |
| `REDIS_HOST` | `HEKA_REDIS_HOST` |
| `REDIS_PORT` | `HEKA_REDIS_PORT` |
| `MILVUS_HOST` | `HEKA_MILVUS_HOST` |
| `MILVUS_PORT` | `HEKA_MILVUS_PORT` |
| `JWT_SECRET` | `HEKA_JWT_SECRET` |
| `CLAUDE_API_KEY` | `HEKA_AI_CLAUDE_API_KEY` |
| `OPENAI_API_KEY` | `HEKA_AI_OPENAI_API_KEY` |

---

## 13. 实现路线图

按依赖关系排序，每个阶段可独立测试和交付：

```
阶段 1: 基础骨架（1 周）
├── shared（ID、错误、类型、事务管理器、EventBus）
├── user（Entity + Repository + Service + Handler）
├── project（Entity + Repository + Service + Handler）
└── 验收：用户登录、创建项目、添加成员

阶段 2: 测试用例核心（1.5 周）
├── module（模块树 CRUD）
├── tag（标签 CRUD）
├── testcase（用例 CRUD + 步骤 + 批量操作 + 状态转换）
├── collection（集合管理）
└── 验收：完整的用例管理流程

阶段 3: 计划与执行（1 周）
├── plan（计划 CRUD + 关联用例 + 状态流转）
├── execution（执行记录 + 结果提交 + 并发控制）
└── 验收：创建计划 → 执行 → 提交结果 → 查看报告

阶段 4: 文件与 RAG（1.5 周）
├── file（上传 + 版本管理 + Figma）
├── rag（分块 + 向量化 + 异步索引 + index_tasks 补偿）
├── report（覆盖度 + 趋势 + 缺陷 + 工作量）
└── 验收：上传文件 → 自动索引 → 向量检索

阶段 5: AI 功能（1 周）
├── ai（Worker Pool + 熔断 + 重试 + 多 Provider）
├── ai_tasks（异步任务管理 + SSE 推送）
└── 验收：AI 生成用例 + 智能分析

阶段 6: 集成与上线（0.5 周）
├── 缓存层集成
├── 速率限制
├── 监控端点
└── 验收：端到端全流程 + Docker Compose 部署
```

---

## 14. 编码规范（关键条目）

1. 包名全小写、单数
2. 错误使用 `%w` 包装，只处理一次（不既 log 又 return）
3. 生产代码无 panic
4. 类型断言使用 comma ok
5. Mutex 使用零值，不内嵌
6. 每个 goroutine 可停止
7. 资源清理使用 defer
8. 公共结构体不内嵌类型
9. 禁止裸 SQL 拼接（参数化查询）
10. 禁止 SELECT *（只查需要的列）
11. 禁止大 Offset 分页（游标分页）
12. 外键列必须有索引
13. 更新操作使用乐观锁（version 字段）
14. Application 层 Service 测试覆盖率 >= 80%
