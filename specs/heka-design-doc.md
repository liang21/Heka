# Heka 详细设计文档

> 简化版 AI 测试管理平台
> 版本：v1.2
> 日期：2025-05-15

---

## 1. 项目概述

### 1.1 项目背景

Heka 是一个面向内部团队的简化版 AI 测试管理平台，基于 MeterSphere 的理念但大幅简化，聚焦核心测试管理场景并深度融合 AI 能力。

### 1.2 目标用户

- **规模**：20-50 人的研发团队
- **场景**：内部测试管理，本地部署
- **角色**：测试工程师、开发工程师、产品经理

### 1.3 核心价值

1. **简化测试管理**：聚焦核心功能，去除冗余
2. **AI 赋能**：AI 辅助用例生成、智能分析
3. **本地部署**：数据安全，响应快速
4. **一人可维护**：单体架构，代码简洁

### 1.4 技术栈

| 层级 | 技术选型 |
|------|----------|
| 前端 | React 18 + TypeScript + Tailwind CSS |
| 后端 | Go 1.21+ |
| ORM | GORM |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| 向量库 | Milvus 2.3+ |
| AI Provider | Claude / OpenAI / Gemini / Ollama |

---

## 2. 功能规格

### 2.1 第一阶段功能清单

#### 2.1.1 核心功能

| 模块 | 功能 | 描述 |
|------|------|------|
| **测试用例管理** | CRUD | 创建、查看、编辑、删除测试用例 |
| | 目录组织 | 文件夹/目录分类管理 |
| | 步骤管理 | 多步骤测试用例，包含动作和预期结果 |
| | 状态管理 | 草稿/就绪/已归档 |
| | 优先级 | 低/中/高/紧急 |
| | 标签 | 自定义标签分类 |
| **测试计划** | 计划管理 | 创建测试计划，关联用例 |
| | 执行分配 | 分配执行人 |
| | 迭代管理 | 支持多轮迭代执行 |
| **执行记录** | 执行记录 | 记录每次执行结果 |
| | 缺陷关联 | 关联外部 Bug ID（Jira/Tapd/飞书） |
| | 执行历史 | 查看历史执行记录 |
| **测试报告** | 计划报告 | 测试计划执行报告 |
| | 用例覆盖度 | 用例覆盖度统计 |
| | 执行通过率 | 通过率分析 |
| **多项目隔离** | 项目管理 | 支持多项目，数据隔离 |
| | 项目切换 | 快速切换项目 |
| **用户管理** | 登录认证 | 用户登录 |
| | 项目成员 | 项目成员管理（无角色区分） |
| **文件管理** | 文件上传 | 支持上传需求文档 |
| | 格式支持 | PDF、Word、Excel、图片、Figma 链接/截图 |
| | 文件列表 | 查看已上传文件 |
| | 版本管理 | 文件版本记录 |
| **RAG 系统** | 文件解析 | 解析文档内容 |
| | 文本分块 | 智能分块 |
| | 向量化 | 生成 Embedding |
| | 向量检索 | 语义搜索 |
| **AI 用例生成** | 需求解析 | 解析需求文档 |
| | 用例生成 | AI 自动生成测试用例 |
| | 批量创建 | 批量创建生成的用例 |
| **AI 智能分析** | 变更分析 | 分析代码变更 |
| | 回归推荐 | 推荐需要回归的用例 |
| | 影响分析 | 分析变更影响范围 |

#### 2.1.2 简化功能

| 模块 | MeterSphere | Heka 简化版 |
|------|-------------|-------------|
| 用例版本控制 | 完整版本历史 + diff | 只保留最后修改人+时间 |
| 基线管理 | 用例基线快照 + 对比 | 简化为"用例集合/收藏" |
| 多维度报表 | 各种图表 + 交叉分析 | 3-5 个核心报表 |
| 审批流程 | 工作流引擎 | 状态标记代替（草稿/待审核/已发布） |
| 关联需求 | 完整需求数据模型 | 只支持需求链接/ID |

#### 2.1.3 不包含功能

- ❌ 接口测试模块（第二阶段）
- ❌ UI 测试模块（第二阶段）
- ❌ 性能测试模块
- ❌ 完整缺陷管理（只关联 Bug ID）
- ❌ 三层租户模型（系统-组织-项目）
- ❌ 资源池管理
- ❌ DevOps 集成
- ❌ 插件体系

---

## 3. 系统架构

### 3.1 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         前端 (React + TS)                    │
├─────────────────────────────────────────────────────────────┤
│  用例管理 │ 计划 │ 执行 │ 报告 │ 文件 │ AI生成 │ AI分析      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API 层 (Gin/Echo)                       │
│                    /api/* → HTTP Handler                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      业务逻辑层                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 用例服务 │  │ 计划服务 │  │ 文件服务 │  │ AI 服务  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │ 执行服务 │  │ 报表服务 │  │ RAG 服务 │                   │
│  └──────────┘  └──────────┘  └──────────┘                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      数据访问层                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  PostgreSQL  │  │    Redis     │  │     Milvus       │  │
│  │  (主数据)     │  │   (缓存)     │  │   (向量存储)      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  AI Provider                          │  │
│  │  Claude │ OpenAI │ Gemini │ Ollama                   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface 层                             │
│  - HTTP Handler                                             │
│  - 中间件 (Auth, CORS, Logger)                              │
│  - 请求/响应 DTO                                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application 层                            │
│  - 应用服务 (用例编排)                                       │
│  - 跨模块协调                                               │
│  - DTO 转换                                                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain 层                                │
│  - 领域模型 (Entity, Value Object)                          │
│  - 仓储接口 (Repository Interface)                          │
│  - 领域服务 (Domain Service)                                │
│  - 业务规则                                                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure 层                           │
│  - PostgreSQL 实现                                          │
│  - Redis 缓存                                               │
│  - Milvus 向量存储                                          │
│  - AI 客户端                                                │
│  - 文件存储                                                 │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                       单机部署                               │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Nginx/Caddy │  │ Heka Backend │  │ Heka Frontend│     │
│  │   (可选)      │  │   :8080      │  │   (静态文件)   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  PostgreSQL  │  │    Redis     │  │    Milvus    │     │
│  │    :5432     │  │    :6379     │  │   :19530     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  文件存储     │  │   日志存储    │                        │
│  │  /var/heka   │  │  /var/log    │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

**资源要求**：
- CPU：4 核心以上
- 内存：8GB 以上
- 磁盘：100GB 以上 SSD

---

## 4. 数据模型

### 4.1 ER 图

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Project   │───────│  TestCase   │───────│   Step      │
│─────────────│ 1    N │─────────────│ 1    N │─────────────│
│ id          │       │ id          │       │ id          │
│ name        │       │ project_id  │       │ test_case_id│
│ description │       │ title       │       │ number      │
│ created_at  │       │ description │       │ action      │
│ updated_at  │       │ status      │       │ expected    │
│             │       │ priority    │       └─────────────┘
│             │       │ tags        │
│             │       │ created_by  │
│             │       │ created_at  │
│             │       └─────────────┘
│             │              │
│             │              │ N
│             │              ▼
│             │       ┌─────────────┐       ┌─────────────┐
│             │       │TestPlan    │───────│Execution    │
│             │       │─────────────│ 1    N │─────────────│
│             │       │ id          │       │ id          │
│             │       │ project_id  │       │ test_plan_id│
│             │       │ name        │       │ executor_id ││
│             │       │ status      │       │ status      ││
│             │       │ started_at  │       │ executed_at ││
│             │       │ ended_at    │       └─────────────┘
│             │       └─────────────┘              │
│             │                                      │ 1
│             │                                      ▼
│             │                               ┌─────────────┐
│             │                               │ExecutionRes │
│             │                               │─────────────│
│             │                               │ id          │
│             │                               │ execution_id│
│             │                               │ test_case_id│
│             │                               │ status      │
│             │                               │ bug_id      │
│             │                               │ notes       │
│             │                               └─────────────┘
│             │
│             │ N
│             ▼
│      ┌─────────────┐       ┌─────────────┐
│      │ProjectMember│───────│    User     │
│      │─────────────│ N    1 │─────────────│
│      │ project_id  │       │ id          │
│      │ user_id     │       │ name        │
│      └─────────────┘       │ email       │
│                            │ password    │
│                            │ created_at  │
│                            └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    File     │───────│   Chunk     │───────│  Embedding  │
│─────────────│ 1    N │─────────────│ 1    1 │─────────────│
│ id          │       │ id          │       │ id          │
│ project_id  │       │ file_id     │       │ chunk_id    │
│ name        │       │ content     │       │ vector      │
│ type        │       │ index       │       │ model       │
│ size        │       │ tokens      │       │ created_at  │
│ path        │       │ created_at  │       └─────────────┘
│ source_type │       └─────────────┘
│ source_url  │
│ uploaded_by │
│ uploaded_at │
│ version     │
└─────────────┘
```

### 4.2 数据库表设计

#### 4.2.1 用户相关

```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目表
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目成员表
CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, user_id)
);
```

#### 4.2.2 测试用例

```sql
-- 业务模块表
CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES modules(id) ON DELETE CASCADE, -- 支持模块嵌套
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
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, ready, archived
    priority INTEGER NOT NULL DEFAULT 1, -- 0=low, 1=medium, 2=high, 3=critical
    tags TEXT[], -- PostgreSQL array type
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0 -- 乐观锁
);

CREATE INDEX idx_test_cases_project ON test_cases(project_id);
CREATE INDEX idx_test_cases_module ON test_cases(module_id);
CREATE INDEX idx_test_cases_status ON test_cases(status);
CREATE INDEX idx_test_cases_priority ON test_cases(priority);
CREATE INDEX idx_test_cases_tags ON test_cases USING GIN(tags);

-- 测试步骤表
CREATE TABLE test_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    action TEXT NOT NULL,
    expected TEXT NOT NULL,
    UNIQUE (test_case_id, number)
);

CREATE INDEX idx_test_steps_case ON test_steps(test_case_id);

-- 用例集合表（简化版基线）
CREATE TABLE test_case_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用例集合关联表
CREATE TABLE collection_cases (
    collection_id UUID NOT NULL REFERENCES test_case_collections(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (collection_id, test_case_id)
);

-- 标签表（统一管理项目标签）
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7), -- #RRGGBB
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, name)
);

CREATE INDEX idx_tags_project ON tags(project_id);
```

##### 4.2.2.1 用例状态转换规则

```yaml
# 用例状态转换规则
draft:
  can_transition_to:
    - ready
    - archived

ready:
  can_transition_to:
    - archived
    - draft

archived:
  can_transition_to:
    - ready

# 不允许的转换
forbidden_transitions:
  - ready -> archived (必须先转为 draft)
  - archived -> draft (必须先转为 ready)
```

#### 4.2.3 测试计划

```sql
-- 测试计划表
CREATE TABLE test_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, active, paused, completed, cancelled
    started_at TIMESTAMP,
    paused_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_execution_id UUID REFERENCES test_executions(id) -- 当前活跃执行
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
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress', -- in_progress, paused, completed, cancelled
    executor_id UUID NOT NULL REFERENCES users(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paused_at TIMESTAMP,
    completed_at TIMESTAMP,
    notes TEXT
);

CREATE INDEX idx_executions_plan ON test_executions(plan_id);
CREATE INDEX idx_executions_executor ON test_executions(executor_id);
CREATE INDEX idx_executions_status ON test_executions(status);

-- 执行结果表
CREATE TABLE execution_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES test_executions(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id),
    executor_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL, -- passed, failed, blocked, skipped
    bug_id VARCHAR(255), -- 外部 Bug ID
    bug_url VARCHAR(500), -- Bug 链接
    notes TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (execution_id, test_case_id)
);

CREATE INDEX idx_execution_results_execution ON execution_results(execution_id);
CREATE INDEX idx_execution_results_status ON execution_results(status);
```

##### 4.2.3.1 并发执行控制策略

```yaml
# 并发执行策略
concurrent_execution:
  rule: "同一测试计划同时只能有一个活跃执行"
  
  check_on_start:
    - 检查是否有 status = 'in_progress' 的执行
    - 如果有，返回错误："已有执行中的测试，请先完成或取消当前执行"
  
  status_transition:
    in_progress:
      can_be_created: true
      check_condition: "无其他 in_progress 状态的执行"
    paused:
      can_resume: true
      allow_from: "in_progress"
    completed:
      final_state: true
    cancelled:
      final_state: true
```

##### 4.2.3.2 测试计划状态转换

```yaml
# 测试计划状态转换
plan_status:
  draft:
    can_transition_to: [active, cancelled]
  
  active:
    can_transition_to: [paused, completed, cancelled]
    requirements:
      - 必须关联至少一个测试用例
  
  paused:
    can_transition_to: [active, cancelled]
    description: "暂停后可恢复执行"
  
  completed:
    can_transition_to: []
    description: "已完成，不可变更"
  
  cancelled:
    can_transition_to: []
    description: "已取消，不可变更"
```

#### 4.2.4 文件管理

```sql
-- 文件表
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL, -- pdf, docx, xlsx, image, figma
    size BIGINT NOT NULL,
    path VARCHAR(1000) NOT NULL, -- 存储路径
    source_type VARCHAR(20) NOT NULL, -- upload, figma_link
    source_url TEXT, -- Figma 链接
    content_preview TEXT, -- 文本预览（用于搜索）
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE, -- 是否已建立索引
    index_status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    index_error TEXT, -- 索引失败原因
    indexed_at TIMESTAMP, -- 索引完成时间
    uploaded_by UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_files_project ON files(project_id);
CREATE INDEX idx_files_type ON files(type);
CREATE INDEX idx_files_indexed ON files(is_indexed);
CREATE INDEX idx_files_index_status ON files(index_status);

-- 文件版本表
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

##### 4.2.4.1 文件解析方式

| 文件类型 | 解析方式 | 技术栈 |
|----------|----------|--------|
| PDF | 文本提取 | PyPDF2 / pdfplumber |
| Word (.docx) | 文本提取 | python-docx |
| Excel (.xlsx) | 文本提取 | openpyxl |
| 图片 | OCR 文字识别 | PaddleOCR / Tesseract |
| Figma 链接 | API 提取 | Figma REST API |
| Figma 截图 | OCR 文字识别 | PaddleOCR / Tesseract |

##### 4.2.4.2 RAG 分块策略

```go
// 分块配置
type ChunkConfig struct {
    MaxTokens    int    `json:"max_tokens"`     // 最大 token 数，默认 500
    Overlap      int    `json:"overlap"`        // 重叠 token 数，默认 50
    MinChunkSize int    `json:"min_chunk_size"` // 最小块大小，默认 100 tokens
}

// 分块策略
chunking:
  strategy: "semantic_overlap"  // 语义重叠分块
  
  config:
    max_tokens: 500
    overlap_tokens: 50
    min_chunk_size: 100
  
  rules:
    - 按段落分块（优先保留段落完整性）
    - 如果段落超过 max_tokens，按句子分割
    - 块之间保留 overlap_tokens 重叠
    - 丢弃小于 min_chunk_size 的块（除非是最后一块）
```

#### 4.2.5 RAG 相关

```sql
-- 文档块表（PostgreSQL 存储元数据，向量在 Milvus）
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

-- 向量嵌入记录表
CREATE TABLE vector_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id UUID NOT NULL REFERENCES document_chunks(id) ON DELETE CASCADE,
    model VARCHAR(100) NOT NULL, -- 使用的 Embedding 模型
    dimension INTEGER NOT NULL,
    milvus_id VARCHAR(255), -- Milvus 中的 ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_embeddings_chunk ON vector_embeddings(chunk_id);
```

#### 4.2.6 索引优化

```sql
-- 全文搜索索引
CREATE INDEX idx_test_cases_fulltext ON test_cases USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- 复合索引
CREATE INDEX idx_test_cases_project_status ON test_cases(project_id, status);
CREATE INDEX idx_test_cases_module_status ON test_cases(module_id, status);
CREATE INDEX idx_executions_plan_status ON test_executions(plan_id, status);
```

#### 4.2.7 缓存策略

| 数据 | 缓存策略 | 过期时间 | Key 模式 |
|------|----------|----------|----------|
| 用户信息 | Cache-Aside | 1h | user:{id} |
| 项目信息 | Cache-Aside | 30min | project:{id} |
| 项目成员列表 | Cache-Aside | 30min | project:{id}:members |
| 用例列表 | Cache-Aside | 5min | project:{id}:testcases:{page}:{filters} |
| 用例详情 | Cache-Aside | 10min | testcase:{id} |
| 模块树 | Cache-Aside | 10min | project:{id}:modules |
| 统计数据 | Write-Through | 10min | project:{id}:stats |
| AI 任务状态 | Cache-Aside | 任务完成后 1h | ai:task:{id} |
| 索引状态 | Cache-Aside | 5min | file:{id}:index |

```go
// 缓存键定义
const (
    CacheKeyUser         = "user:%s"
    CacheKeyProject      = "project:%s"
    CacheKeyProjectMembers = "project:%s:members"
    CacheKeyTestCases    = "project:%s:testcases:%d:%s"  // page, filters_hash
    CacheKeyTestCase     = "testcase:%s"
    CacheKeyModules      = "project:%s:modules"
    CacheKeyStats        = "project:%s:stats"
    CacheKeyAITask       = "ai:task:%s"
    CacheKeyFileIndex    = "file:%s:index"
)
```

---

## 5. API 设计

### 5.1 通用规范

#### 5.1.0 API 版本管理

**策略**：URL 前缀版本

```
/api/v1/...     # 当前版本
/api/v2/...     # 未来版本（预留）
```

**规则**：
- 所有 API 端点必须包含版本前缀
- 版本升级时保持向后兼容至少 6 个月
- 破坏性变更必须使用新版本号
- 同一版本内只允许新增字段，不允许删除或重命名

#### 5.1.1 请求/响应格式

**请求头**：
```
Content-Type: application/json
Authorization: Bearer {token}
```

**成功响应**：
```json
{
  "code": 0,
  "data": { ... },
  "message": "success"
}
```

**错误响应**：
```json
{
  "code": 400,
  "message": "error message"
}
```

#### 5.1.2 状态码

| 状态码 | 含义 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器错误 |

#### 5.1.3 分页参数验证

```yaml
pagination:
  page:
    type: integer
    min: 1
    default: 1
    description: 页码，从 1 开始
  
  page_size:
    type: integer
    min: 1
    max: 100
    default: 20
    description: 每页数量，最大 100

# 示例
GET /api/v1/testcases?page=1&page_size=20
```

#### 5.1.4 排序参数验证

```yaml
sorting:
  sort_by:
    type: string
    enum: [created_at, updated_at, title, priority, status]
    default: created_at
    description: 排序字段
  
  sort_desc:
    type: boolean
    default: true
    description: 是否降序

# 示例
GET /api/v1/testcases?sort_by=created_at&sort_desc=true
```

#### 5.1.5 ID 参数格式

```yaml
id_format:
  type: UUID v4
  pattern: '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  example: "550e8400-e29b-41d4-a716-446655440000"
  description: 所有资源 ID 使用 UUID v4 格式
```

### 5.2 认证相关

#### 5.2.1 用户登录

```
POST /api/v1/auth/login
```

**请求**：
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "张三",
      "email": "user@example.com"
    }
  }
}
```

#### 5.2.2 获取当前用户

```
GET /api/v1/auth/me
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "张三",
    "email": "user@example.com"
  }
}
```

### 5.3 项目相关

#### 5.3.1 创建项目

```
POST /api/v1/projects
```

**请求**：
```json
{
  "name": "Heka 测试平台",
  "description": "AI 测试管理平台"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Heka 测试平台",
    "description": "AI 测试管理平台",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-05-15T10:00:00Z"
  }
}
```

#### 5.3.2 获取项目列表

```
GET /api/v1/projects
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "projects": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Heka 测试平台",
        "description": "AI 测试管理平台",
        "member_count": 5,
        "created_at": "2025-05-15T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

#### 5.3.3 获取项目详情

```
GET /api/v1/projects/{id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Heka 测试平台",
    "description": "AI 测试管理平台",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "members": [
      {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "张三",
        "email": "zhangsan@example.com",
        "joined_at": "2025-05-15T10:00:00Z"
      }
    ],
    "statistics": {
      "test_cases": 150,
      "test_plans": 10,
      "executions": 50
    }
  }
}
```

#### 5.3.4 添加项目成员

```
POST /api/v1/projects/{id}/members
```

**请求**：
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "成员添加成功"
}
```

### 5.4 模块管理

#### 5.4.1 创建模块

```
POST /api/v1/modules
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "parent_id": null,
  "name": "用户认证模块",
  "description": "包含登录、注册、密码重置等功能"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "用户认证模块",
    "created_at": "2025-05-15T10:00:00Z"
  }
}
```

#### 5.4.2 获取模块树

```
GET /api/v1/modules?project_id={id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "modules": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "用户认证模块",
        "description": "包含登录、注册、密码重置等功能",
        "case_count": 25,
        "children": [
          {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "name": "登录功能",
            "description": "用户登录相关",
            "case_count": 15,
            "children": []
          },
          {
            "id": "550e8400-e29b-41d4-a716-446655440002",
            "name": "注册功能",
            "description": "用户注册相关",
            "case_count": 10,
            "children": []
          }
        ]
      }
    ]
  }
}
```

#### 5.4.3 更新模块

```
PUT /api/v1/modules/{id}
```

**请求**：
```json
{
  "name": "认证模块",
  "description": "用户认证相关功能"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "模块更新成功"
}
```

#### 5.4.4 删除模块

```
DELETE /api/v1/modules/{id}
```

**响应**：
```json
{
  "code": 0,
  "message": "模块删除成功"
}
```

**注意**：删除模块时，该模块下的用例会自动解除关联（module_id 设为 NULL）。

### 5.5 测试用例相关

#### 5.5.1 创建测试用例

```
DELETE /api/v1/testcases/batch
```

**请求**：
```json
{
  "ids": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "deleted": 2
  }
}
```

#### 5.4.7 批量移动用例

```
PUT /api/v1/testcases/batch/move
```

**请求**：
```json
{
  "ids": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"],
  "folder_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "moved": 2
  }
}
```

#### 5.4.8 获取项目标签

```
GET /api/v1/tags?project_id={id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "tags": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "P0",
        "color": "#FF0000",
        "usage_count": 15
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "name": "登录",
        "color": "#00FF00",
        "usage_count": 20
      }
    ]
  }
}
```

#### 5.4.9 创建标签

```
POST /api/v1/tags
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "P0",
  "color": "#FF0000"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "P0",
    "color": "#FF0000"
  }
}
```

### 5.5 用例集合相关

#### 5.5.1 创建用例集合

```
POST /api/v1/collections
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "回归测试集",
  "description": "每次回归必测"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-05-15T10:00:00Z"
  }
}
```

#### 5.5.2 添加用例到集合

```
POST /api/v1/collections/{id}/cases
```

**请求**：
```json
{
  "test_case_ids": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "added": 2
  }
}
```

#### 5.5.3 获取集合的用例列表

```
GET /api/v1/collections/{id}/cases?page=1&page_size=20
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "test_cases": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "status": "ready",
        "priority": 1,
        "tags": ["登录"],
        "added_at": "2025-05-15T10:00:00Z"
      }
    ],
    "total": 50
  }
}
```

#### 5.5.4 从集合移除用例

```
DELETE /api/v1/collections/{id}/cases
```

**请求**：
```json
{
  "test_case_ids": ["550e8400-e29b-41d4-a716-446655440000"]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "removed": 1
  }
}
```

### 5.6 测试用例相关（原 5.4）

#### 5.6.1 创建测试用例

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "用户登录成功",
  "description": "验证用户使用正确的用户名和密码可以成功登录",
  "steps": [
    {
      "action": "打开登录页面",
      "expected": "显示登录表单"
    },
    {
      "action": "输入正确的用户名和密码",
      "expected": "输入框显示输入内容"
    },
    {
      "action": "点击登录按钮",
      "expected": "登录成功，跳转到首页"
    }
  ],
  "priority": 1,
  "tags": ["登录", "功能测试"]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-05-15T10:00:00Z"
  }
}
```

#### 5.4.2 获取测试用例列表

```
GET /api/v1/testcases?project_id={id}&page=1&page_size=20&status=ready&priority=2&keyword=登录
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "test_cases": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "project_id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "description": "验证用户使用正确的用户名和密码可以成功登录",
        "status": "ready",
        "priority": 1,
        "tags": ["登录", "功能测试"],
        "step_count": 3,
        "created_by": "550e8400-e29b-41d4-a716-446655440000",
        "created_by_name": "张三",
        "created_at": "2025-05-15T10:00:00Z",
        "updated_at": "2025-05-15T10:00:00Z"
      }
    ],
    "total": 150,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.4.3 获取测试用例详情

```
GET /api/v1/testcases/{id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "用户登录成功",
    "description": "验证用户使用正确的用户名和密码可以成功登录",
    "status": "ready",
    "priority": 1,
    "tags": ["登录", "功能测试"],
    "steps": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "number": 1,
        "action": "打开登录页面",
        "expected": "显示登录表单"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "number": 2,
        "action": "输入正确的用户名和密码",
        "expected": "输入框显示输入内容"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "number": 3,
        "action": "点击登录按钮",
        "expected": "登录成功，跳转到首页"
      }
    ],
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_by_name": "张三",
    "created_at": "2025-05-15T10:00:00Z",
    "updated_by": "550e8400-e29b-41d4-a716-446655440000",
    "updated_by_name": "张三",
    "updated_at": "2025-05-15T10:00:00Z",
    "version": 0
  }
}
```

#### 5.4.4 更新测试用例

```
PUT /api/v1/testcases/{id}
```

**请求**：同创建请求

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "updated_at": "2025-05-15T11:00:00Z"
  }
}
```

#### 5.4.5 删除测试用例

```
DELETE /api/v1/testcases/{id}
```

**响应**：
```json
{
  "code": 0,
  "message": "删除成功"
}
```

#### 5.4.6 批量更新状态

```
PUT /api/v1/testcases/batch/status
```

**请求**：
```json
{
  "ids": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"],
  "status": "ready"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "updated": 2
  }
}
```

### 5.7 测试计划相关

#### 5.7.1 创建测试计划

```
POST /api/v1/testplans
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "V1.0 回归测试",
  "description": "V1.0 版本发布前的回归测试",
  "test_case_ids": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001"],
  "assignments": [
    {
      "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
      "assigned_to": "550e8400-e29b-41d4-a716-446655440000"
    }
  ]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-05-15T10:00:00Z"
  }
}
```

#### 5.7.2 获取测试计划列表

```
GET /api/v1/testplans?project_id={id}&status=active
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "plans": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "project_id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "V1.0 回归测试",
        "description": "V1.0 版本发布前的回归测试",
        "status": "active",
        "test_case_count": 50,
        "created_by": "550e8400-e29b-41d4-a716-446655440000",
        "created_by_name": "张三",
        "created_at": "2025-05-15T10:00:00Z"
      }
    ],
    "total": 10
  }
}
```

#### 5.7.3 获取测试计划详情

```
GET /api/v1/testplans/{id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "V1.0 回归测试",
    "description": "V1.0 版本发布前的回归测试",
    "status": "active",
    "started_at": "2025-05-15T10:00:00Z",
    "test_cases": [
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "assigned_to": "550e8400-e29b-41d4-a716-446655440000",
        "assigned_to_name": "张三",
        "order_index": 1
      }
    ],
    "progress": {
      "total": 50,
      "executed": 30,
      "passed": 25,
      "failed": 5,
      "pass_rate": 83.3
    },
    "current_execution": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "in_progress",
      "executor_name": "张三",
      "started_at": "2025-05-15T10:00:00Z"
    }
  }
}
```

#### 5.7.4 开始执行测试计划

```
POST /api/v1/testplans/{id}/start
```

**请求**：
```json
{
  "name": "第一次执行",
  "executor_notes": "回归测试"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "execution_id": "550e8400-e29b-41d4-a716-446655440000",
    "started_at": "2025-05-15T10:00:00Z"
  }
}
```

**错误响应**（已有执行中的测试）：
```json
{
  "code": 1009,
  "message": "已有执行中的测试，请先完成或取消当前执行"
}
```

#### 5.7.5 暂停执行

```
POST /api/v1/testplans/{id}/pause
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "paused_at": "2025-05-15T12:00:00Z"
  }
}
```

#### 5.7.6 恢复执行

```
POST /api/v1/testplans/{id}/resume
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "resumed_at": "2025-05-15T14:00:00Z"
  }
}
```

#### 5.7.7 完成测试计划

```
POST /api/v1/testplans/{id}/complete
```

**请求**：
```json
{
  "summary": "测试完成，发现 3 个缺陷"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "测试计划已完成"
}
```

#### 5.7.8 取消测试计划

```
POST /api/v1/testplans/{id}/cancel
```

**请求**：
```json
{
  "reason": "需求变更，取消本次测试"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "测试计划已取消"
}
```

### 5.6 执行相关

#### 5.6.1 提交执行结果

```
POST /api/v1/executions/{execution_id}/results
```

**请求**：
```json
{
  "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "passed",
  "bug_id": "",
  "notes": "执行通过"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### 5.6.2 批量提交执行结果

```
POST /api/v1/executions/{execution_id}/results/batch
```

**请求**：
```json
{
  "results": [
    {
      "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "passed",
      "bug_id": "",
      "notes": "执行通过"
    },
    {
      "test_case_id": "550e8400-e29b-41d4-a716-446655440001",
      "status": "failed",
      "bug_id": "BUG-123",
      "notes": "发现 Bug"
    }
  ]
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "created": 2
  }
}
```

#### 5.6.3 获取执行详情

```
GET /api/v1/executions/{id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "plan_id": "550e8400-e29b-41d4-a716-446655440000",
    "plan_name": "V1.0 回归测试",
    "status": "completed",
    "executed_at": "2025-05-15T10:00:00Z",
    "completed_at": "2025-05-15T15:00:00Z",
    "executor": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "张三"
    },
    "summary": {
      "total": 50,
      "passed": 45,
      "failed": 3,
      "blocked": 1,
      "skipped": 1,
      "pass_rate": 90
    },
    "results": [
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
        "test_case_title": "用户登录成功",
        "status": "passed",
        "executor_name": "张三",
        "executed_at": "2025-05-15T10:30:00Z",
        "bug_id": "",
        "notes": "执行通过"
      }
    ]
  }
}
```

### 5.10 报告相关

#### 5.10.1 获取测试计划报告

```
GET /api/v1/reports/plan/{plan_id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "plan": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "V1.0 回归测试"
    },
    "summary": {
      "total_cases": 50,
      "executed_cases": 50,
      "passed": 45,
      "failed": 3,
      "blocked": 1,
      "skipped": 1,
      "pass_rate": 90
    },
    "failed_cases": [
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "bug_id": "BUG-123",
        "bug_url": "https://jira.example.com/BUG-123",
        "executor_name": "张三"
      }
    ],
    "execution_history": [
      {
        "execution_id": "550e8400-e29b-41d4-a716-446655440000",
        "executed_at": "2025-05-15T10:00:00Z",
        "pass_rate": 85
      },
      {
        "execution_id": "550e8400-e29b-41d4-a716-446655440001",
        "executed_at": "2025-05-14T10:00:00Z",
        "pass_rate": 88
      }
    ]
  }
}
```

#### 5.10.2 获取用例覆盖度报告

```
GET /api/v1/reports/coverage?project_id={id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "total_cases": 150,
    "by_status": {
      "draft": 20,
      "ready": 100,
      "archived": 30
    },
    "by_priority": {
      "low": 30,
      "medium": 80,
      "high": 30,
      "critical": 10
    },
    "by_folder": [
      {
        "folder_id": "550e8400-e29b-41d4-a716-446655440000",
        "folder_name": "登录模块",
        "case_count": 25
      },
      {
        "folder_id": "550e8400-e29b-41d4-a716-446655440001",
        "folder_name": "注册功能",
        "case_count": 15
      }
    ],
    "recent_activity": {
      "created_today": 5,
      "updated_today": 10,
      "executed_today": 20
    }
  }
}
```

#### 5.10.3 获取执行趋势报告

```
GET /api/v1/reports/trend?project_id={id}&days=30
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "period": {
      "start_date": "2025-04-15",
      "end_date": "2025-05-15",
      "days": 30
    },
    "daily_trend": [
      {
        "date": "2025-05-15",
        "executed": 50,
        "passed": 45,
        "failed": 5,
        "pass_rate": 90
      },
      {
        "date": "2025-05-14",
        "executed": 40,
        "passed": 35,
        "failed": 5,
        "pass_rate": 87.5
      }
    ],
    "summary": {
      "total_executions": 1200,
      "avg_pass_rate": 88.5,
      "best_day": {
        "date": "2025-05-10",
        "pass_rate": 95
      },
      "worst_day": {
        "date": "2025-05-01",
        "pass_rate": 75
      }
    }
  }
}
```

#### 5.10.4 获取缺陷分布报告

```
GET /api/v1/reports/bugs?project_id={id}&start_date=2025-05-01&end_date=2025-05-15
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "period": {
      "start_date": "2025-05-01",
      "end_date": "2025-05-15"
    },
    "summary": {
      "total_bugs": 45,
      "by_status": {
        "open": 20,
        "in_progress": 15,
        "resolved": 8,
        "closed": 2
      },
      "by_severity": {
        "critical": 5,
        "high": 10,
        "medium": 20,
        "low": 10
      }
    },
    "top_failed_cases": [
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "fail_count": 8,
        "bugs": ["BUG-123", "BUG-145", "BUG-167"]
      }
    ],
    "bug_sources": [
      {
        "source": "Jira",
        "count": 30
      },
      {
        "source": "Tapd",
        "count": 10
      },
      {
        "source": "飞书",
        "count": 5
      }
    ]
  }
}
```

#### 5.10.5 获取个人工作量报告

```
GET /api/v1/reports/workload?user_id={id}&start_date=2025-05-01&end_date=2025-05-15
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_name": "张三",
    "period": {
      "start_date": "2025-05-01",
      "end_date": "2025-05-15"
    },
    "summary": {
      "created_cases": 15,
      "updated_cases": 30,
      "executed_cases": 100,
      "execution_pass_rate": 95,
      "bugs_found": 8
    },
    "daily_breakdown": [
      {
        "date": "2025-05-15",
        "created": 2,
        "executed": 10,
        "pass_rate": 90
      }
    ]
  }
}
```

### 5.8 文件相关

#### 5.8.1 上传文件

```
POST /api/v1/files/upload
Content-Type: multipart/form-data
```

**请求**：
```
project_id: 550e8400-e29b-41d4-a716-446655440000
file: [二进制文件]
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "需求文档.pdf",
    "type": "pdf",
    "size": 1024000,
    "uploaded_at": "2025-05-15T10:00:00Z",
    "index_status": "pending"
  }
}
```

#### 5.8.2 上传 Figma 链接

```
POST /api/v1/files/figma
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "figma_url": "https://www.figma.com/file/xxx",
  "name": "UI 设计稿"
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "UI 设计稿",
    "type": "figma",
    "source_url": "https://www.figma.com/file/xxx",
    "uploaded_at": "2025-05-15T10:00:00Z",
    "index_status": "pending"
  }
}
```

#### 5.8.3 获取文件列表

```
GET /api/v1/files?project_id={id}&type=pdf&indexed=true
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "files": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "需求文档.pdf",
        "type": "pdf",
        "size": 1024000,
        "uploaded_by": "550e8400-e29b-41d4-a716-446655440000",
        "uploaded_by_name": "张三",
        "uploaded_at": "2025-05-15T10:00:00Z",
        "is_indexed": true,
        "index_status": "completed",
        "indexed_at": "2025-05-15T10:05:00Z"
      }
    ],
    "total": 10
  }
}
```

#### 5.8.4 获取文件详情

```
GET /api/v1/files/{id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "需求文档.pdf",
    "type": "pdf",
    "size": 1024000,
    "source_type": "upload",
    "source_url": null,
    "uploaded_by": "550e8400-e29b-41d4-a716-446655440000",
    "uploaded_by_name": "张三",
    "uploaded_at": "2025-05-15T10:00:00Z",
    "version": 1,
    "is_indexed": true,
    "index_status": "completed",
    "indexed_at": "2025-05-15T10:05:00Z",
    "chunk_count": 25
  }
}
```

#### 5.8.5 重新索引文件

```
POST /api/v1/files/{id}/reindex
```

**请求**：
```json
{
  "force": true
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "processing"
  }
}
```

#### 5.8.6 获取索引状态

```
GET /api/v1/files/{id}/index-status
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "file_id": "550e8400-e29b-41d4-a716-446655440000",
    "is_indexed": false,
    "index_status": "processing",
    "index_progress": {
      "current": 10,
      "total": 25,
      "percent": 40
    },
    "index_error": null,
    "indexed_at": null
  }
}
```

#### 5.8.7 删除文件

```
DELETE /api/v1/files/{id}
```

**响应**：
```json
{
  "code": 0,
  "message": "文件已删除"
}
```

### 5.9 AI 相关

#### 5.9.1 AI 生成测试用例

```
POST /api/v1/ai/generate-testcases
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "query": "用户登录功能",
  "options": {
    "count": 10,
    "priority": "medium",
    "include_negative": true
  }
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "processing",
    "estimated_time": 30
  }
}
```

#### 5.9.2 获取 AI 任务状态

```
GET /api/v1/ai/tasks/{task_id}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "completed",
    "progress": {
      "current": 10,
      "total": 10,
      "percent": 100
    },
    "result": {
      "test_cases": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "title": "用户登录成功",
          "description": "验证用户使用正确的用户名和密码可以成功登录",
          "steps": [...],
          "priority": 1,
          "tags": ["登录", "功能测试"]
        }
      ],
      "count": 10
    }
  }
}
```

#### 5.9.3 AI 智能分析

```
POST /api/v1/ai/analyze
```

**请求**：
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "changes": {
    "type": "code",
    "description": "修改了用户登录逻辑，增加多因子认证",
    "files": [
      {
        "path": "auth/login.go",
        "change_type": "modified",
        "functions": ["login", "verify_mfa"]
      },
      {
        "path": "auth/session.go",
        "change_type": "modified",
        "functions": ["create_session", "validate_session"]
      }
    ],
    "diff": "@@ -15,7 +15,9 @@ func login(c *gin.Context) {\n-     if validatePassword(username, password) {\n+     if validatePassword(username, password) && verifyMFA(code) {\n         createSession(user)",
    "commit_message": "feat: add MFA support for login",
    "pr_url": "https://github.com/xxx/pull/123"
  },
  "options": {
    "include_indirect": true,
    "max_results": 50,
    "min_confidence": 0.6
  }
}
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "affected_cases": [
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "用户登录成功",
        "reason": "直接关联：测试登录功能，涉及 MFA 认证",
        "confidence": 0.95,
        "impact_level": "high"
      },
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "会话保持",
        "reason": "间接关联：依赖登录会话创建逻辑",
        "confidence": 0.75,
        "impact_level": "medium"
      },
      {
        "test_case_id": "550e8400-e29b-41d4-a716-446655440002",
        "title": "MFA 验证码验证",
        "reason": "新增关联：新功能需要测试",
        "confidence": 0.90,
        "impact_level": "high"
      }
    ],
    "summary": {
      "total": 15,
      "high_confidence": 10,
      "medium_confidence": 5,
      "high_impact": 8,
      "medium_impact": 5,
      "low_impact": 2
    },
    "recommendations": [
      "优先测试 MFA 相关用例",
      "回归测试登录相关用例",
      "检查会话管理功能"
    ]
  }
}
```

### 5.9.4 AI 任务进度推送（SSE）

```
GET /api/v1/ai/tasks/{task_id}/events
```

**响应**（Server-Sent Events 流）：
```
event: progress
data: {"current": 5, "total": 10, "percent": 50}

event: progress
data: {"current": 10, "total": 10, "percent": 100}

event: completed
data: {"task_id": "xxx", "count": 10}

event: error
data: {"code": "AI-IE-001", "message": "AI 服务不可用"}
```

**使用场景**：
- AI 用例生成进度
- 文件索引进度
- AI 智能分析进度

### 5.10 监控相关

#### 5.10.1 获取系统健康状态

```
GET /api/v1/health
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "status": "healthy",
    "services": {
      "database": "ok",
      "redis": "ok",
      "milvus": "ok"
    },
    "version": "1.0.0"
  }
}
```

#### 5.10.2 获取 AI 服务状态

```
GET /api/v1/monitoring/ai
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "providers": {
      "claude": {
        "state": "closed",
        "requests": 150,
        "successes": 145,
        "failures": 5,
        "success_rate": 0.967
      },
      "openai": {
        "state": "closed",
        "requests": 50,
        "successes": 48,
        "failures": 2,
        "success_rate": 0.96
      }
    },
    "pool": {
      "queue_length": 2,
      "max_workers": 10,
      "active_tasks": 8
    }
  }
}
```

---

## 6. 前端设计

### 6.1 技术栈

- **框架**：React 18
- **语言**：TypeScript
- **样式**：Tailwind CSS
- **路由**：React Router v6
- **状态管理**：React Query + Zustand
- **UI 组件**：Shadcn/ui
- **表单**：React Hook Form + Zod
- **HTTP 客户端**：Axios / Fetch

### 6.2 页面结构

```
├── /login                    # 登录页
├── /                         # 首页（项目列表）
├── /project/:id              # 项目主页
│   ├── /overview             # 项目概览
│   ├── /modules             # 模块管理
│   │   ├── /list             # 模块列表（树形）
│   │   └── /create           # 创建模块
│   ├── /testcases            # 测试用例
│   │   ├── /list             # 用例列表
│   │   ├── /collections      # 用例集合
│   │   ├── /create           # 创建用例
│   │   └── /:id              # 用例详情
│   ├── /plans                # 测试计划
│   │   ├── /list             # 计划列表
│   │   ├── /create           # 创建计划
│   │   └── /:id              # 计划详情
│   │       ├── /edit          # 编辑计划
│   │       └── /execute       # 执行页面
│   ├── /executions           # 执行记录
│   │   ├── /list             # 执行列表
│   │   └── /:id              # 执行详情
│   ├── /reports              # 测试报告
│   │   ├── /plan             # 计划报告
│   │   ├── /coverage         # 覆盖度
│   │   ├── /trend            # 执行趋势
│   │   ├── /bugs             # 缺陷分布
│   │   └── /workload         # 工作量
│   ├── /files                # 文件管理
│   │   ├── /list             # 文件列表
│   │   └── /:id              # 文件详情
│   └── /settings             # 项目设置
└── /settings                 # 系统设置
```

### 6.3 核心页面

#### 6.3.1 测试用例列表

**布局**：
- 左侧：模块树（可折叠）
- 中间：用例列表（表格视图）
- 右侧：筛选器（状态、优先级、标签）

**功能**：
- 批量操作（批量删除、批量更新状态）
- 拖拽排序
- 快速创建
- 导出/导入

#### 6.3.2 用例详情

**布局**：
- 顶部：用例信息（标题、状态、优先级、标签）
- 中部：步骤列表
- 底部：执行历史

**功能**：
- 实时编辑
- 步骤拖拽排序
- 版本历史

#### 6.3.3 测试计划执行

**布局**：
- 左侧：用例列表
- 右侧：执行面板

**功能**：
- 一键开始执行
- 快速标记通过/失败
- 关联 Bug
- 添加备注

### 6.4 组件库

基于 Shadcn/ui，扩展以下组件：

```
src/components/
├── ui/                       # Shadcn/ui 组件
├── Module/
│   ├── ModuleTree.tsx         # 模块树（可折叠、拖拽）
│   ├── ModuleCard.tsx         # 模块卡片
│   └── ModuleForm.tsx         # 模块表单
├── TestCase/
│   ├── TestCaseCard.tsx      # 用例卡片
│   ├── TestCaseList.tsx      # 用例列表
│   ├── StepEditor.tsx        # 步骤编辑器
│   └── TestCaseForm.tsx      # 用例表单
├── TestPlan/
│   ├── PlanCard.tsx          # 计划卡片
│   ├── ExecutionPanel.tsx    # 执行面板
│   └── ResultSummary.tsx     # 结果汇总
├── Collection/
│   ├── CollectionCard.tsx    # 集合卡片
│   ├── AddCasesModal.tsx     # 添加用例弹窗
│   └── CasesList.tsx         # 集合用例列表
├── File/
│   ├── FileUploader.tsx      # 文件上传
│   ├── FileList.tsx          # 文件列表
│   ├── FileCard.tsx          # 文件卡片
│   └── FigmaEmbed.tsx        # Figma 预览
└── AI/
    ├── AIGenerateForm.tsx    # AI 生成表单
    ├── TaskMonitor.tsx       # 任务监控
    └── AnalysisResult.tsx    # 分析结果展示
```

---

## 7. 部署方案

### 7.1 本地部署

#### 7.1.1 Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL
  postgres:
    image: postgres:15
    container_name: heka-postgres
    environment:
      POSTGRES_DB: heka
      POSTGRES_USER: heka
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - heka-network

  # Redis
  redis:
    image: redis:7-alpine
    container_name: heka-redis
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    networks:
      - heka-network

  # Milvus (使用 Milvus Standalone)
  milvus:
    image: milvusdb/milvus:v2.3.0
    container_name: heka-milvus
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    volumes:
      - milvus-data:/var/lib/milvus
    ports:
      - "19530:19530"
    depends_on:
      - etcd
      - minio
    networks:
      - heka-network

  # Milvus 依赖
  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    container_name: heka-etcd
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
    volumes:
      - etcd-data:/etcd
    networks:
      - heka-network

  minio:
    image: minio/minio:RELEASE.2023-03-20T20-16-18Z
    container_name: heka-minio
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    volumes:
      - minio-data:/minio_data
    command: minio server /minio_data
    networks:
      - heka-network

  # Heka Backend
  heka-backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: heka-backend
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_NAME=heka
      - DATABASE_USER=heka
      - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - MILVUS_HOST=milvus
      - MILVUS_PORT=19530
      - JWT_SECRET=${JWT_SECRET}
      - AI_CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - AI_OPENAI_API_KEY=${OPENAI_API_KEY}
    volumes:
      - heka-uploads:/app/uploads
      - heka-logs:/app/logs
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - milvus
    networks:
      - heka-network

  # Heka Frontend
  heka-frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: heka-frontend
    ports:
      - "3000:80"
    depends_on:
      - heka-backend
    networks:
      - heka-network

  # Nginx (可选，作为反向代理)
  nginx:
    image: nginx:alpine
    container_name: heka-nginx
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - heka-backend
      - heka-frontend
    networks:
      - heka-network

volumes:
  postgres-data:
  redis-data:
  milvus-data:
  etcd-data:
  minio-data:
  heka-uploads:
  heka-logs:

networks:
  heka-network:
    driver: bridge
```

#### 7.1.2 环境变量

```bash
# .env
POSTGRES_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret

# AI Provider API Keys
CLAUDE_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GEMINI_API_KEY=xxx

# 文件上传
MAX_UPLOAD_SIZE=104857600  # 100MB
UPLOAD_PATH=/app/uploads
```

### 7.2 生产部署

#### 7.2.1 系统服务配置

```ini
# /etc/systemd/system/heka.service
[Unit]
Description=Heka Backend Service
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=heka
WorkingDirectory=/opt/heka/backend
ExecStart=/opt/heka/backend/heka-server
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/etc/heka/config.yaml

[Install]
WantedBy=multi-user.target
```

#### 7.2.2 Nginx 配置

```nginx
# /etc/nginx/conf.d/heka.conf
upstream heka_backend {
    server 127.0.0.1:8080;
}

upstream heka_frontend {
    server 127.0.0.1:3000;
}

server {
    listen 80;
    server_name heka.example.com;

    # 前端
    location / {
        proxy_pass http://heka_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # API
    location /api/ {
        proxy_pass http://heka_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 文件上传大小限制
    client_max_body_size 100M;
}
```

---

## 8. 安全设计

### 8.1 认证授权

#### 8.1.1 JWT 认证

- 使用 JWT 进行用户认证
- Token 有效期：24 小时
- 支持 Token 刷新
- 密码使用 bcrypt 加密

#### 8.1.2 权限控制

- 项目成员才能访问项目数据
- 无角色区分，所有成员权限相同
- API 层进行权限验证

### 8.2 数据安全

- 敏感信息加密存储
- API 密钥使用环境变量
- 文件上传类型限制
- SQL 注入防护（参数化查询）

### 8.3 网络安全

- HTTPS 传输
- CORS 限制
- CSRF 防护
- 速率限制

---

## 9. 开发计划

### 9.1 第一阶段（MVP）

| 模块 | 功能 | 工作量 |
|------|------|--------|
| 用户与认证 | 登录、JWT | 2 天 |
| 项目管理 | 创建、列表、成员 | 3 天 |
| 测试用例 | CRUD、步骤、标签 | 5 天 |
| 测试计划 | 创建、关联用例 | 3 天 |
| 执行记录 | 提交结果、历史 | 3 天 |
| 测试报告 | 计划报告、覆盖度 | 3 天 |
| 文件管理 | 上传、列表、解析 | 5 天 |
| RAG 系统 | 分块、向量化、检索 | 5 天 |
| AI 用例生成 | 生成、批量创建 | 3 天 |
| 前端基础 | 框架、路由、布局 | 3 天 |
| 前端页面 | 核心页面实现 | 10 天 |

**总计**：约 45 个工作日

### 9.2 里程碑

| 里程碑 | 交付内容 | 时间 |
|--------|----------|------|
| M1 | 基础框架 + 用户认证 | 第 1 周 |
| M2 | 测试用例管理 | 第 3 周 |
| M3 | 测试计划与执行 | 第 5 周 |
| M4 | 文件管理 + RAG | 第 7 周 |
| M5 | AI 功能集成 | 第 8 周 |
| M6 | 前端完成 | 第 10 周 |
| M7 | 测试与优化 | 第 11 周 |
| M8 | 部署上线 | 第 12 周 |

---

## 10. 附录

### 10.1 错误码规范

#### 10.1.1 错误码格式

```
{domain}-{type}-{number}

domain: 领域缩写（2-4 字母）
  AUTH  - 认证鉴权
  TC    - 测试用例
  TP    - 测试计划
  EX    - 执行记录
  FILE  - 文件管理
  RAG   - RAG 检索
  AI    - AI 服务
  PROJ  - 项目管理
  USER  - 用户管理
  SYS   - 系统级

type: 错误类型（2 字母）
  NF    - Not Found
  VL    - Validation
  AU    - Auth
  CF    - Conflict
  IE    - Internal Error
  TE    - Timeout
  RT    - Rate Limit

number: 3 位序号（001-999）
```

#### 10.1.2 错误码定义

**认证鉴权（AUTH）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| AUTH-AU-001 | 401 | 未认证，缺少或无效 Token |
| AUTH-AU-002 | 401 | Token 已过期 |
| AUTH-AU-003 | 403 | 无权限访问该项目 |
| AUTH-VL-001 | 400 | 邮箱或密码错误 |

**测试用例（TC）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| TC-NF-001 | 404 | 测试用例不存在 |
| TC-VL-001 | 400 | 标题不能为空 |
| TC-VL-002 | 400 | 至少需要一个步骤 |
| TC-VL-003 | 400 | 无效的优先级值 |
| TC-VL-004 | 400 | 无效的状态转换 |
| TC-CF-001 | 409 | 版本冲突，数据已被修改 |
| TC-CF-002 | 409 | 用例标题重复 |

**测试计划（TP）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| TP-NF-001 | 404 | 测试计划不存在 |
| TP-VL-001 | 400 | 计划必须关联至少一个用例 |
| TP-VL-002 | 400 | 无效的计划状态转换 |
| TP-CF-001 | 409 | 已有执行中的测试 |

**执行记录（EX）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| EX-NF-001 | 404 | 执行记录不存在 |
| EX-VL-001 | 400 | 无效的执行状态 |
| EX-CF-001 | 409 | 该用例已提交执行结果 |

**文件管理（FILE）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| FILE-NF-001 | 404 | 文件不存在 |
| FILE-VL-001 | 400 | 不支持的文件类型 |
| FILE-VL-002 | 400 | 文件大小超限 |
| FILE-VL-003 | 400 | 文件正在索引中 |

**AI 服务（AI）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| AI-IE-001 | 503 | AI 服务不可用（所有 Provider 故障） |
| AI-TE-001 | 504 | AI 调用超时 |
| AI-RT-001 | 429 | AI 请求排队已满 |
| AI-IE-002 | 500 | AI 响应解析失败 |

**项目管理（PROJ）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| PROJ-NF-001 | 404 | 项目不存在 |
| PROJ-CF-001 | 409 | 项目名已存在 |

**系统级（SYS）**

| 错误码 | HTTP 状态码 | 含义 |
|--------|------------|------|
| SYS-IE-001 | 500 | 数据库错误 |
| SYS-IE-002 | 500 | 缓存服务错误 |
| SYS-IE-003 | 503 | 向量数据库不可用 |
| SYS-VL-001 | 400 | 参数验证失败 |
| SYS-RT-001 | 429 | 请求频率超限 |
```

#### 10.1.3 错误响应格式

```json
{
  "code": "TC-NF-001",
  "message": "测试用例不存在",
  "details": {}
}
```

#### 10.1.4 Domain 层错误映射

```go
// internal/domain/testcase/errors.go
package testcase

import "errors"

var (
    ErrNotFound            = errors.New("TC-NF-001:test case not found")
    ErrTitleRequired       = errors.New("TC-VL-001:title is required")
    ErrAtLeastOneStep      = errors.New("TC-VL-002:at least one step required")
    ErrInvalidPriority     = errors.New("TC-VL-003:invalid priority")
    ErrInvalidTransition   = errors.New("TC-VL-004:invalid status transition")
    ErrVersionConflict     = errors.New("TC-CF-001:version conflict")
)
```

### 10.2 配置项说明

```yaml
# config.yaml
server:
  port: 8080
  mode: release  # debug, release

database:
  host: localhost
  port: 5432
  name: heka
  user: heka
  password: ${DB_PASSWORD}
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0

milvus:
  host: localhost
  port: 19530
  collection_name: heka_chunks

jwt:
  secret: ${JWT_SECRET}
  expire: 24h

upload:
  max_size: 104857600  # 100MB
  allowed_types:
    - pdf
    - docx
    - xlsx
    - png
    - jpg
    - jpeg
  path: /var/heka/uploads

ai:
  pool_max_workers: 10
  pool_queue_size: 100
  retry_max_attempts: 3
  timeout_request: 60s
  providers:
    claude:
      api_key: ${CLAUDE_API_KEY}
      model: claude-3-5-sonnet-20241022
      priority: 1
      enabled: true
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4o
      priority: 2
      enabled: true
```

### 10.3 依赖服务版本

| 服务 | 版本 |
|------|------|
| PostgreSQL | 15+ |
| Redis | 7+ |
| Milvus | 2.3+ |
| Go | 1.21+ |
| Node.js | 18+ |

---

**文档版本**：v1.2
**最后更新**：2025-05-15
**作者**：Heka Team

**更新记录**：
- v1.2 (2025-05-15): 修正目录概念为业务模块（modules），将 test_case_folders 改为 modules 表；更新相关 API 和前端页面结构
- v1.1 (2025-05-15): 根据审阅意见更新，新增用例目录、集合管理、标签管理等 API；删除 AI 助手；增强 AI 分析功能；补充状态转换规则、并发控制策略、缓存策略等
- v1.0 (2025-05-15): 初始版本
