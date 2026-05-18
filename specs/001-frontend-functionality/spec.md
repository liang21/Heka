# Heka 前端功能实现规格

> 编译自 specs/frontend-design-spec.md v1.1，面向实现的统一规格
> 版本：v1.1
> 日期：2026-05-18

---

## 1. 概述

### 1.1 项目定位

Heka 前端是面向 20-50 人内部团队的 AI 测试管理平台 Web 界面，采用 SPA 架构，配合后端 RESTful API + JWT 认证。

**核心价值**：简洁高效的测试管理体验 | AI 辅助生成 | 响应式桌面端优先

### 1.2 技术栈

| 层级 | 技术选型 | 版本 |
|------|----------|------|
| 框架 | React | 18.x |
| 语言 | TypeScript | 5.x |
| 构建工具 | Vite | 5.x |
| 样式方案 | Tailwind CSS | 3.x |
| UI 组件库 | Shadcn/ui (Radix UI) | latest |
| 路由 | React Router | 6.x |
| 客户端状态 | Zustand | 4.x |
| 服务端状态 | TanStack Query | 5.x |
| 表单 | React Hook Form + Zod | latest |
| HTTP 客户端 | Axios | 1.x |
| 图表 | Recharts | 2.x |
| 拖拽 | @dnd-kit/core | 6.x |
| 日期 | date-fns | 3.x |
| 图标 | Lucide React | latest |

### 1.3 第一阶段功能范围

| 模块 | 页面/功能 |
|------|----------|
| 认证 | 登录页、Token 管理、自动刷新 |
| 项目管理 | 项目列表（首页）、项目概览、成员管理 |
| 模块管理 | 模块树 CRUD、拖拽排序 |
| 测试用例 | 用例列表（三栏布局）、创建/编辑用例、用例详情、步骤编辑器、批量操作 |
| 用例集合 | 集合列表、添加/移除用例 |
| 测试计划 | 计划列表、创建计划、计划详情、用例分配 |
| 执行管理 | 执行页面（左右分栏）、结果标记、键盘快捷键 |
| 执行记录 | 执行列表、执行详情 |
| 测试报告 | 计划报告、覆盖度、趋势、缺陷分布、工作量 |
| 文件管理 | 文件上传（含分块上传）、文件列表、文件详情、索引状态、Figma 链接 |
| AI 功能 | AI 用例生成（异步轮询）、AI 智能分析 |
| 通用 | 命令面板 (Cmd+K)、错误处理、加载状态、空状态 |

**不包含**：移动端适配、SSR、PWA、WebSocket 实时通信、国际化（仅预留）

---

## 2. 模块清单

| 模块 | 核心组件 | 自定义 Hooks | API 服务 | Zustand Store |
|------|---------|-------------|---------|--------------|
| `auth` | LoginPage | useAuth, useAuthInit | auth.ts | auth.ts |
| `project` | ProjectsPage, OverviewPage | useProject | project.ts | project.ts |
| `module` | ModulesPage, ModuleTree, ModuleForm | — | module.ts | — |
| `testcase` | TestCasesPage, CreateTestCase, TestCaseDetail, StepEditor, TestCaseFilters, BatchActions, PriorityBadge | — | testcase.ts | — |
| `collection` | CollectionsPage, AddCasesDialog, CollectionCasesList | — | collection.ts | — |
| `plan` | TestPlansPage, CreatePlan, PlanDetail | — | plan.ts | — |
| `execution` | ExecutePlan, ExecutionPanel, ResultSummary | — | execution.ts | — |
| `report` | PlanReport, Coverage, Trend, BugDistribution, Workload | — | report.ts | — |
| `file` | FilesPage, FileDetail, FileUploader, ChunkUploader, IndexStatusBadge | useFileIndexStatus | file.ts | — |
| `ai` | AIGeneratePage, AIAnalysisPage, GenerateForm, TaskProgress | useAITask | ai.ts | — |
| `shared` | ConfirmDialog, EmptyState, ErrorBoundary, Pagination, SearchInput, StatusTag, CommandPalette | usePagination, useConfirm, useUnsavedChanges, useKeyboard | api.ts (Axios) | — |

---

## 3. 类型系统

### 3.1 通用类型

```tsx
// src/types/api.ts
interface ApiResponse<T> {
  code: number;
  data: T;
  message: string;
}

interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

type PaginatedResponse<T> = ApiResponse<PaginatedData<T>>;

interface PaginationParams {
  page: number;
  page_size: number;
}
```

### 3.2 认证类型

```tsx
// src/types/auth.ts
interface User {
  id: string;
  name: string;
  email: string;
}

interface LoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  token: string;
  user: User;
}
```

### 3.3 项目类型

```tsx
// src/types/project.ts
interface Project {
  id: string;
  name: string;
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface ProjectDetail extends Project {
  members: ProjectMember[];
  statistics: ProjectStatistics;
}

interface ProjectMember {
  user_id: string;
  name: string;
  email: string;
  joined_at: string;
}

interface ProjectStatistics {
  test_cases: number;
  test_plans: number;
  executions: number;
}
```

### 3.4 模块类型

```tsx
// src/types/module.ts
interface Module {
  id: string;
  project_id: string;
  name: string;
  description: string;
  parent_id: string | null;
  order_index: number;
  case_count: number;
  children: Module[];
  created_at: string;
}
```

### 3.5 测试用例类型

```tsx
// src/types/testcase.ts
type TestCaseStatus = "draft" | "ready" | "archived";
type Priority = 0 | 1 | 2 | 3;

interface TestStep {
  id?: string;
  action: string;
  expected: string;
}

interface TestCase {
  id: string;
  project_id: string;
  module_id: string | null;
  title: string;
  description: string;
  status: TestCaseStatus;
  priority: Priority;
  tags: string[];
  steps: TestStep[];
  created_by: string;
  created_by_name: string;
  created_at: string;
  updated_by: string;
  updated_by_name: string;
  updated_at: string;
  version: number;
}

interface TestCaseListParams {
  project_id: string;
  page?: number;
  page_size?: number;
  status?: TestCaseStatus;
  priority?: Priority;
  module_id?: string;
  keyword?: string;
  tags?: string[];
  sort_by?: string;
  sort_desc?: boolean;
}
```

### 3.6 计划与执行类型

```tsx
// src/types/plan.ts
type PlanStatus = "draft" | "active" | "paused" | "completed" | "cancelled";

interface TestPlan {
  id: string;
  project_id: string;
  name: string;
  description: string;
  status: PlanStatus;
  started_at: string | null;
  paused_at: string | null;
  ended_at: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  current_execution_id: string | null;
}

interface PlanDetail extends TestPlan {
  test_cases: PlanTestCase[];
  progress: PlanProgress;
  current_execution: ExecutionSummary | null;
}

interface PlanTestCase {
  test_case_id: string;
  title: string;
  priority: Priority;
  assigned_to: string;
  assigned_to_name: string;
  order_index: number;
}

interface PlanProgress {
  total: number;
  passed: number;
  failed: number;
  blocked: number;
  skipped: number;
  not_executed: number;
}

interface ExecutionSummary {
  id: string;
  status: ExecutionStatus;
  executor_name: string;
  progress: PlanProgress;
}
```

```tsx
// src/types/execution.ts
type ExecutionStatus = "in_progress" | "paused" | "completed" | "cancelled";
type ResultStatus = "passed" | "failed" | "blocked" | "skipped";

interface Execution {
  id: string;
  plan_id: string;
  plan_name: string;
  status: ExecutionStatus;
  executor_id: string;
  executor_name: string;
  started_at: string;
  paused_at: string | null;
  completed_at: string | null;
  notes: string;
}

interface ExecutionDetail extends Execution {
  summary: PlanProgress;
  results: ExecutionResult[];
}

interface ExecutionResult {
  id: string;
  test_case_id: string;
  test_case_title: string;
  executor_id: string;
  executor_name: string;
  status: ResultStatus;
  bug_id: string | null;
  bug_url: string | null;
  notes: string | null;
  executed_at: string;
}

interface SubmitResultRequest {
  test_case_id: string;
  status: ResultStatus;
  bug_id?: string;
  notes?: string;
}
```

### 3.7 文件类型

```tsx
// src/types/file.ts
type FileSourceType = "upload" | "figma";
type FileIndexStatus = "pending" | "processing" | "completed" | "failed";

interface ProjectFile {
  id: string;
  project_id: string;
  name: string;
  type: string;
  size: number;
  source_type: FileSourceType;
  source_url: string | null;
  is_indexed: boolean;
  index_status: FileIndexStatus;
  index_error: string | null;
  indexed_at: string | null;
  chunk_count: number;
  uploaded_by: string;
  uploaded_by_name: string;
  uploaded_at: string;
  version: number;
}
```

### 3.8 AI 类型

```tsx
// src/types/ai.ts
type AITaskStatus = "pending" | "processing" | "completed" | "failed";

interface AITask {
  task_id: string;
  status: AITaskStatus;
  progress: {
    current: number;
    total: number;
    message: string;
  };
  result: {
    test_cases: AIGeneratedCase[];
    count: number;
  } | null;
  error: string | null;
  estimated_time: number | null;
}

interface AIGeneratedCase {
  title: string;
  description: string;
  steps: TestStep[];
  priority: Priority;
  tags: string[];
}

interface AIAnalysisResult {
  affected_cases: Array<{
    test_case_id: string;
    title: string;
    reason: string;
  }>;
  summary: {
    total_affected: number;
    by_priority: Record<string, number>;
  };
  recommendations: string[];
}
```

### 3.9 报表类型

```tsx
// src/types/report.ts
interface PlanReport {
  plan: TestPlan;
  summary: PlanProgress;
  failed_cases: ExecutionResult[];
  execution_history: Execution[];
}

interface CoverageReport {
  project_id: string;
  total_cases: number;
  by_status: Record<TestCaseStatus, number>;
  by_priority: Record<string, number>;
  by_module: Array<{ module_name: string; total: number; executed: number }>;
}

interface TrendReport {
  project_id: string;
  daily_trend: Array<{
    date: string;
    total: number;
    passed: number;
    failed: number;
  }>;
}

interface BugReport {
  project_id: string;
  summary: { total_bugs: number; by_source: Record<string, number> };
  top_failed_cases: ExecutionResult[];
}

interface WorkloadReport {
  user_id: string;
  user_name: string;
  summary: { total_executed: number; passed: number; failed: number };
  daily_breakdown: Array<{
    date: string;
    executed: number;
    passed: number;
    failed: number;
  }>;
}
```

### 3.10 标签与集合类型

```tsx
// src/types/tag.ts
interface Tag {
  id: string;
  project_id: string;
  name: string;
  color: string;
  usage_count: number;
}

// src/types/collection.ts
interface Collection {
  id: string;
  project_id: string;
  name: string;
  description: string;
  created_by: string;
  created_at: string;
}
```

### 3.11 筛选与查询参数类型

```tsx
// src/types/filters.ts
interface TestCaseFilters {
  status?: TestCaseStatus;
  priority?: Priority;
  module_id?: string;
  keyword?: string;
  tags?: string[];
}

interface PlanFilters {
  status?: PlanStatus;
}

interface ExecutionFilters {
  status?: ExecutionStatus;
}

interface FileFilters {
  type?: string;
  indexed?: boolean;
}

interface CollectionCaseFilters {
  page?: number;
  page_size?: number;
}

interface BugReportParams {
  start_date?: string;
  end_date?: string;
}

interface WorkloadParams {
  start_date?: string;
  end_date?: string;
}
```

---

## 4. 状态管理规格

### 4.1 Zustand Store

**auth store（持久化）**：

```tsx
// src/stores/auth.ts
// 持久化 token 到 localStorage
// user 信息由 /auth/me 接口刷新，不持久化
interface AuthState {
  token: string | null;
  user: User | null;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  isAuthenticated: () => boolean;
}
// 使用 persist 中间件，partialize 仅保存 token
```

**project store（持久化）**：

```tsx
// src/stores/project.ts
// 持久化 currentProjectId 和 sidebarCollapsed
interface ProjectState {
  currentProjectId: string | null;
  setCurrentProject: (id: string) => void;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}
```

### 4.2 TanStack Query Key 规范

```tsx
// src/lib/query-keys.ts
export const queryKeys = {
  projects: {
    all: ["projects"],
    detail: (id: string) => ["projects", id],
    members: (id: string) => ["projects", id, "members"],
  },
  modules: {
    tree: (projectId: string) => ["projects", projectId, "modules"],
  },
  testCases: {
    list: (projectId: string, filters: TestCaseFilters) =>
      ["projects", projectId, "testcases", filters],
    detail: (id: string) => ["testcases", id],
  },
  plans: {
    list: (projectId: string, filters?: PlanFilters) =>
      ["projects", projectId, "plans", filters],
    detail: (id: string) => ["plans", id],
  },
  executions: {
    list: (projectId: string, filters?: ExecutionFilters) =>
      ["projects", projectId, "executions", filters],
    detail: (id: string) => ["executions", id],
  },
  files: {
    list: (projectId: string, filters?: FileFilters) =>
      ["projects", projectId, "files", filters],
    detail: (id: string) => ["files", id],
    indexStatus: (id: string) => ["files", id, "index-status"],
    reindexTask: (id: string) => ["files", id, "reindex"],
  },
  ai: {
    task: (taskId: string) => ["ai", "tasks", taskId],
  },
  reports: {
    plan: (planId: string) => ["reports", "plan", planId],
    coverage: (projectId: string) => ["reports", "coverage", projectId],
    trend: (projectId: string, days: number = 30) => ["reports", "trend", projectId, days],
    bugs: (projectId: string, params: BugReportParams) => ["reports", "bugs", projectId, params],
    workload: (userId: string, params: WorkloadParams) => ["reports", "workload", userId, params],
  },
  tags: {
    list: (projectId: string) => ["projects", projectId, "tags"],
  },
  collections: {
    list: (projectId: string) => ["projects", projectId, "collections"],
    detail: (id: string) => ["collections", id],
    cases: (id: string, filters?: CollectionCaseFilters) => ["collections", id, "cases", filters],
  },
};
```

### 4.3 缓存策略

| 数据类型 | staleTime | gcTime | 说明 |
|----------|-----------|--------|------|
| 用户信息 | 30min | 1h | 配合 /auth/me 刷新 |
| 项目列表 | 5min | 30min | 不频繁变化 |
| 模块树 | 10min | 1h | 结构稳定 |
| 用例列表 | 2min | 30min | 多人可能修改 |
| 用例详情 | 5min | 30min | |
| 文件列表 | 2min | 30min | |
| AI 任务 | 0 | 30min | 轮询模式，完成后 30min |
| 报表数据 | 10min | 1h | 历史数据不变化 |

---

## 5. API 服务层规格

### 5.1 Axios 实例配置

```tsx
// src/services/api.ts
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
  timeout: 30000,
  headers: { "Content-Type": "application/json" },
});
```

**请求拦截器**：从 `useAuthStore.getState().token` 读取 Token，注入 `Authorization: Bearer {token}`。

**响应拦截器**：
1. 成功：返回 `response.data`（解包 Axios 层）
2. 401 错误：
   - 首次 401：尝试 POST `/auth/refresh`，用新 Token 重试原请求
   - 并发请求：排队等待刷新完成
   - 刷新失败：清除登录态，跳转 `/login`

**应用初始化**：`useAuthInit` Hook 在应用启动时，若 localStorage 有 token，调用 `GET /auth/me` 校验并恢复 user 信息。

### 5.2 API 服务清单

每个服务模块导出一个对象，方法返回 Axios Promise：

**auth.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `login(data)` | `POST /auth/login` | 登录 |
| `getMe()` | `GET /auth/me` | 获取当前用户 |
| `refresh()` | `POST /auth/refresh` | 刷新 Token |

**project.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list()` | `GET /projects` | 项目列表 |
| `create(data)` | `POST /projects` | 创建项目 |
| `getById(id)` | `GET /projects/:id` | 项目详情 |
| `addMember(id, userId)` | `POST /projects/:id/members` | 添加成员 |

**module.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `tree(projectId)` | `GET /modules?project_id=:id` | 模块树 |
| `create(data)` | `POST /modules` | 创建模块 |
| `update(id, data)` | `PUT /modules/:id` | 更新模块 |
| `delete(id)` | `DELETE /modules/:id` | 删除模块 |

**testcase.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list(params)` | `GET /testcases` | 用例列表 |
| `getById(id)` | `GET /testcases/:id` | 用例详情 |
| `create(data)` | `POST /testcases` | 创建用例 |
| `update(id, data)` | `PUT /testcases/:id` | 更新用例（含 version） |
| `delete(id)` | `DELETE /testcases/:id` | 删除用例 |
| `batchUpdateStatus(ids, status)` | `PUT /testcases/batch/status` | 批量改状态 |
| `batchDelete(ids)` | `DELETE /testcases/batch` | 批量删除 |
| `batchMove(ids, folderId)` | `PUT /testcases/batch/move` | 批量移动 |

**collection.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list(projectId)` | `GET /collections?project_id=:id` | 集合列表 |
| `create(data)` | `POST /collections` | 创建集合 |
| `getById(id)` | `GET /collections/:id` | 集合详情 |
| `getCases(id, params)` | `GET /collections/:id/cases` | 集合用例 |
| `addCases(id, ids)` | `POST /collections/:id/cases` | 添加用例 |
| `removeCases(id, ids)` | `DELETE /collections/:id/cases` | 移除用例 |

**plan.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list(projectId, filters)` | `GET /testplans` | 计划列表 |
| `create(data)` | `POST /testplans` | 创建计划 |
| `getById(id)` | `GET /testplans/:id` | 计划详情 |
| `start(id, data)` | `POST /testplans/:id/start` | 开始执行 |
| `pause(id)` | `POST /testplans/:id/pause` | 暂停 |
| `resume(id)` | `POST /testplans/:id/resume` | 恢复 |
| `complete(id, data)` | `POST /testplans/:id/complete` | 完成 |
| `cancel(id, data)` | `POST /testplans/:id/cancel` | 取消 |

**execution.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `getById(id)` | `GET /executions/:id` | 执行详情 |
| `submitResult(id, data)` | `POST /executions/:id/results` | 提交结果 |
| `submitBatchResults(id, results)` | `POST /executions/:id/results/batch` | 批量提交 |

**file.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list(projectId, filters)` | `GET /files` | 文件列表 |
| `upload(projectId, file)` | `POST /files/upload` | 小文件上传（≤10MB） |
| `uploadLarge(projectId, file, onProgress)` | 分块上传（>10MB） | `POST /files/upload/init` → `POST /files/upload/chunk` × N → `POST /files/upload/complete`。**注意：这三个端点待后端实现，当前后端仅支持 `POST /files/upload` 单次上传** |
| `addFigmaLink(projectId, url, name)` | `POST /files/figma` | Figma 链接 |
| `getById(id)` | `GET /files/:id` | 文件详情 |
| `getIndexStatus(id)` | `GET /files/:id/index-status` | 索引状态 |
| `reindex(id, force)` | `POST /files/:id/reindex` | 重新索引 |
| `delete(id)` | `DELETE /files/:id` | 删除文件 |

**ai.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `generate(data)` | `POST /ai/generate-testcases` | 发起 AI 生成（返回 task_id） |
| `getTaskStatus(taskId)` | `GET /ai/tasks/:id` | 查询任务状态（轮询用） |
| `analyze(data)` | `POST /ai/analyze` | AI 分析 |

**report.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `plan(planId)` | `GET /reports/plan/:plan_id` | 计划报告 |
| `coverage(projectId)` | `GET /reports/coverage` | 覆盖度 |
| `trend(projectId, days?)` | `GET /reports/trend` | 趋势（`days` 默认 30） |
| `bugs(projectId, params)` | `GET /reports/bugs` | 缺陷分布 |
| `workload(userId, params)` | `GET /reports/workload` | 工作量 |

**tag.ts**：

| 方法 | 端点 | 说明 |
|------|------|------|
| `list(projectId)` | `GET /tags` | 标签列表 |
| `create(data)` | `POST /tags` | 创建标签 |

### 5.3 Query Hook 封装模式

```tsx
// 列表查询：使用 useQuery + placeholderData 保持旧数据
function useTestCases(projectId: string, filters: TestCaseFilters) {
  return useQuery({
    queryKey: queryKeys.testCases.list(projectId, filters),
    queryFn: () => testcaseApi.list({ project_id: projectId, ...filters }),
    placeholderData: (prev) => prev,
  });
}

// 创建/更新/删除：使用 useMutation + invalidateQueries 精确刷新
function useCreateTestCase(projectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: testcaseApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["projects", projectId, "testcases"],
      });
    },
  });
}

// AI 任务轮询：使用 refetchInterval 实现自动轮询
function useAITask(taskId: string | null) {
  return useQuery({
    queryKey: queryKeys.ai.task(taskId!),
    queryFn: () => aiApi.getTaskStatus(taskId!),
    enabled: !!taskId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === "completed" || status === "failed") return false;
      return 2000;
    },
    staleTime: 30 * 60 * 1000,
  });
}

// 文件索引状态轮询
function useFileIndexStatus(fileId: string | null) {
  return useQuery({
    queryKey: queryKeys.files.indexStatus(fileId!),
    queryFn: () => fileApi.getIndexStatus(fileId!),
    enabled: !!fileId,
    refetchInterval: (query) => {
      const status = query.state.data?.index_status;
      if (status === "completed" || status === "failed") return false;
      return 3000;
    },
  });
}
```

---

## 6. 路由规格

### 6.1 路由表

```tsx
const routes = [
  { path: "/login", element: <Login /> },
  {
    path: "/",
    element: <AppLayout />,    // 含 TopBar + Sidebar
    children: [
      { index: true, element: <Projects /> },
      { path: "settings", element: <Settings /> },
      {
        path: "project/:projectId",
        element: <ProjectLayout />,  // 项目布局壳
        children: [
          { index: true, element: <Navigate to="overview" replace /> },
          { path: "overview", element: <Overview /> },
          { path: "modules", element: <Modules /> },
          { path: "modules/create", element: <CreateModule /> },
          { path: "testcases", element: <TestCases /> },
          { path: "testcases/create", element: <CreateTestCase /> },
          { path: "testcases/collections", element: <Collections /> },
          { path: "testcases/:id", element: <TestCaseDetail /> },
          { path: "plans", element: <TestPlans /> },
          { path: "plans/create", element: <CreatePlan /> },
          { path: "plans/:id", element: <PlanDetail /> },
          { path: "plans/:id/execute", element: <ExecutePlan /> },
          { path: "executions", element: <Executions /> },
          { path: "executions/:id", element: <ExecutionDetail /> },
          { path: "reports/plan", element: <PlanReport /> },
          { path: "reports/coverage", element: <Coverage /> },
          { path: "reports/trend", element: <Trend /> },
          { path: "reports/bugs", element: <BugDistribution /> },
          { path: "reports/workload", element: <Workload /> },
          { path: "files", element: <Files /> },
          { path: "files/:id", element: <FileDetail /> },
          { path: "ai-generate", element: <AIGenerate /> },
          { path: "ai-analysis", element: <AIAnalysis /> },
          { path: "settings", element: <ProjectSettings /> },
        ],
      },
    ],
  },
];
```

**注意**：`testcases/collections` 必须在 `testcases/:id` 之前声明，避免 `collections` 被匹配为动态参数。

### 6.2 路由守卫

- 未登录访问任何页面 → 跳转 `/login`
- 已登录访问 `/login` → 跳转 `/`
- 项目成员校验由后端 API 层处理（403 响应）

### 6.3 路由级懒加载

所有页面组件使用 `React.lazy` + `Suspense`：

```tsx
const TestCases = lazy(() => import("@/pages/TestCases"));
const TestPlans = lazy(() => import("@/pages/TestPlans"));
// ...
```

---

## 7. 页面组件规格

### 7.1 登录页

| 元素 | 实现 |
|------|------|
| 布局 | 居中卡片，max-w-[400px]，白色卡片 + 浅灰背景 |
| 表单 | React Hook Form + Zod 校验（email 必填、password 必填） |
| 提交 | `useMutation` 调用 `authApi.login`，成功后 `setAuth` + 跳转 `/` |
| 错误处理 | 后端错误码 `AUTH-VL-001` → 表单下方红色文字 |

### 7.2 项目列表（首页）

| 元素 | 实现 |
|------|------|
| 布局 | 卡片网格，响应式 1/2/3/4 列 |
| 数据 | `useQuery` 调用 `projectApi.list` |
| 卡片内容 | name, description, member_count, test_case_count, created_at |
| 操作 | 点击卡片 → navigate(`/project/${id}`)；新建按钮 → Dialog |
| 空状态 | "还没有项目，创建一个开始吧" + 新建按钮 |

### 7.3 项目概览

| 元素 | 实现 |
|------|------|
| 布局 | 统计卡片行 + 最近活动列表 + 快捷操作 |
| 数据 | `useQuery` 调用 `projectApi.getById(projectId)` 获取 ProjectDetail |
| 统计卡片 | 用例总数、计划总数、执行总数、成员数（来自 statistics 字段） |
| 最近活动 | 最近更新的用例/计划/执行记录（各取前 5 条） |
| 快捷操作 | 新建用例、新建计划、上传文件、AI 生成 |
| 成员列表 | 展示项目成员头像和名称（来自 members 字段） |

### 7.4 模块管理

| 元素 | 实现 |
|------|------|
| 布局 | 项目布局 + 内容区树形列表 |
| 组件 | ModuleTree（递归渲染，支持展开/折叠） |
| 拖拽 | @dnd-kit 实现排序和移动 |
| CRUD | 新建/编辑 → Dialog（ModuleForm），删除 → ConfirmDialog |
| 删除确认 | "该模块下有 N 个用例，用例将移至'未分类'" |
| 数据刷新 | 操作成功后 `invalidateQueries` 刷新模块树 |

### 7.4 测试用例列表（三栏布局）

| 元素 | 实现 |
|------|------|
| 布局 | 左侧模块树（260px，可折叠）+ 右侧表格区 |
| 筛选 | URL 参数同步：`?page=1&status=ready&priority=2&keyword=登录` |
| 表格 | 行选择（checkbox）、排序列头、PriorityBadge、StatusTag |
| 分页 | 底部分页器，使用 `usePagination` Hook |
| 批量操作 | 选中行后底部固定操作栏：改状态、移动、删除 |
| 模块筛选 | 点击模块树节点 → 更新 URL 的 `module_id` 参数 |

### 7.5 创建/编辑测试用例

| 元素 | 实现 |
|------|------|
| 表单 | React Hook Form + Zod（title 必填 max 500, steps ≥ 1） |
| 步骤编辑器 | StepEditor 组件：动态数组，支持增删、拖拽排序 |
| 标签输入 | TagInput 组件：从已有标签选择 + 自定义输入 |
| 提交 | 创建 → `testcaseApi.create`；编辑 → `testcaseApi.update(id, { ...data, version })` |
| 乐观锁 | 编辑页加载时记录 `version`，提交时携带；收到 409 TC-CF-001 → 冲突提示 |
| 离开拦截 | `useUnsavedChanges(form.formState.isDirty)` |

### 7.6 用例详情

| 元素 | 实现 |
|------|------|
| 布局 | 面包屑 + 标题区 + 描述 + 步骤表格 + 执行历史 |
| 状态按钮 | 根据状态流转（第 9.1 节）动态显示 |
| 编辑模式 | 点击"编辑"→ 标题、描述、步骤变为可编辑 |
| 执行历史 | `useQuery` 查询关联的 ExecutionResult[] |

### 7.7 测试计划详情

| 元素 | 实现 |
|------|------|
| 操作按钮 | 根据计划状态流转（第 9.2 节）动态显示 |
| 用例列表 | PlanTestCase[] 表格，支持批量添加、移除、分配执行人 |
| 执行历史 | 该计划下所有 Execution 记录 |
| 执行并发 | 点击"开始执行"前检查是否有 in_progress 的执行（第 9.3 节） |

### 7.8 执行页面（左右分栏）

| 元素 | 实现 |
|------|------|
| 布局 | 左 40% 用例列表，右 60% 执行面板 |
| 用例列表 | 显示执行结果图标，当前用例高亮 |
| 执行面板 | 展示步骤详情 + 4 个结果按钮（通过/失败/阻塞/跳过） |
| 失败扩展 | 选择"失败"时展开 Bug ID 输入框 |
| 提交 | `executionApi.submitResult` → 刷新执行详情 → 自动跳转下一个 |
| 键盘快捷键 | P=通过, F=失败, B=阻塞, S=跳过, Enter=确认并下一个 |
| 进度 | 底部进度条 + 统计摘要 |

### 7.9 AI 用例生成

| 元素 | 实现 |
|------|------|
| 来源文件 | 文件选择列表（仅显示已索引文件），多选 |
| 生成表单 | query 描述 + count + priority + include_negative |
| 提交流程 | `aiApi.generate` → 获得 task_id → `useAITask(taskId)` 轮询 |
| 进度展示 | TaskProgress 组件显示进度百分比 |
| 结果列表 | 勾选 + 预览弹窗，全选/取消全选 |
| 批量创建 | 使用 `Promise.allSettled` 并发创建（最多 5 个并发），全部完成后汇总成功/失败数。若后端后续提供批量创建接口则切换为单次调用 |

### 7.10 文件管理

| 元素 | 实现 |
|------|------|
| 上传 | FileUploader：≤10MB 直接上传，>10MB 分块上传（ChunkUploader） |
| 分块上传 | init → chunk × N → complete，每块 5MB，显示进度条 |
| 索引状态 | 上传完成后自动 `useFileIndexStatus` 轮询 |
| 文件详情 | 索引状态、版本历史、重新索引（含 force 选项）、删除 |
| Figma | 输入 Figma URL + 名称，调用 `fileApi.addFigmaLink` |

### 7.11 报表页面

| 页面 | 图表 | 数据源 |
|------|------|--------|
| 计划报告 | Recharts PieChart（通过率）+ 失败用例列表 | `reportApi.plan` |
| 覆盖度 | Recharts BarChart（按模块/优先级） | `reportApi.coverage` |
| 趋势 | Recharts LineChart（每日通过/失败） | `reportApi.trend` |
| 缺陷分布 | Recharts BarChart + top 失败用例 | `reportApi.bugs` |
| 工作量 | Recharts BarChart（每日执行量） | `reportApi.workload` |

---

## 8. 状态流转与 UI 控制

### 8.1 测试用例状态

```
draft → ready → archived
  ↑       ↓        ↓
  └───────┘    ←───┘
```

| 当前状态 | 可用操作 |
|---------|---------|
| `draft` | [标记就绪] [归档] [编辑] [删除] |
| `ready` | [改回草稿] [归档] [编辑] [删除] |
| `archived` | [恢复] [删除] |

### 8.2 测试计划状态

```
draft → active → completed
  ↓       ↓
  ↓     paused → active / cancelled
  ↓       ↓
cancelled ← cancelled
```

| 当前状态 | 可用操作 |
|---------|---------|
| `draft` | [开始执行] [编辑] [取消计划] |
| `active` | [暂停] [查看执行] |
| `paused` | [恢复执行] [取消计划] |
| `completed` | [查看报告] |
| `cancelled` | 仅查看 |

### 8.3 执行并发控制

- 同一计划同时只能有一个 `in_progress` 执行
- 前端预检：点击"开始执行"时查询是否有未完成执行
- 冲突提示：提供"继续上次执行"或"取消并新建"
- 后端兜底：`TP-CF-001` (409)

### 8.4 乐观锁（版本控制）

- 用例编辑页记录 `version`，提交时携带
- 收到 `TC-CF-001` (409) → 弹窗提供"查看最新版本"或"强制覆盖"

---

## 9. 交互规格

### 9.1 键盘快捷键

| 快捷键 | 上下文 | 动作 |
|--------|--------|------|
| `Cmd/Ctrl + K` | 全局 | 命令面板 |
| `Cmd/Ctrl + N` | 用例列表 | 新建用例 |
| `Cmd/Ctrl + S` | 编辑页面 | 保存 |
| `Escape` | 弹窗/编辑态 | 关闭/取消 |
| `P` | 执行页面 | 标记通过 |
| `F` | 执行页面 | 标记失败 |
| `B` | 执行页面 | 标记阻塞 |
| `S` | 执行页面 | 标记跳过 |
| `Enter` | 执行页面 | 确认并下一个 |
| `?` | 全局 | 快捷键帮助 |
| `Cmd/Ctrl + /` | 侧边栏 | 折叠/展开 |

### 9.2 确认弹窗

| 操作 | 确认方式 |
|------|----------|
| 删除单个用例 | ConfirmDialog："确定删除？" |
| 批量删除 | ConfirmDialog + 输入数量 |
| 删除项目 | ConfirmDialog + 输入项目名 |
| 状态变更 | 直接操作 + Toast |
| 离开编辑页 | unsaved_changes 提示（`useBlocker` + `beforeunload`） |

### 9.3 加载状态

| 场景 | 处理方式 |
|------|----------|
| 页面首次加载 | 骨架屏（Skeleton） |
| 列表翻页 | 保持旧数据 + 顶部进度条 |
| 提交表单 | 按钮禁用 + Spinner |
| 删除操作 | 乐观更新 + 失败回滚 |
| AI 生成 | 进度条 + 轮询更新 |

### 9.4 空状态

| 页面 | 文案 | 操作按钮 |
|------|------|----------|
| 项目列表 | "还没有项目，创建一个开始吧" | + 新建项目 |
| 用例列表 | "该模块下还没有用例" | + 创建 / AI 生成 |
| 文件列表 | "还没有上传文件" | + 上传文件 |
| 报表 | "暂无数据" | — |

---

## 10. 错误处理规格

### 10.1 HTTP 错误处理策略

| HTTP 状态码 | 处理方式 |
|-------------|----------|
| 400 | 表单字段高亮 + 错误信息 |
| 401 | 静默刷新 Token → 失败则跳转登录 |
| 403 | Toast "无权限" |
| 404 | 404 页面 |
| 409 | 冲突信息弹窗（版本冲突/执行冲突） |
| 429 | Toast "操作过于频繁" |
| 500 | Toast "服务器错误" |
| 网络错误 | Toast "网络连接失败" |

### 10.2 错误码映射

```tsx
const errorMessages: Record<string, string> = {
  "AUTH-AU-001": "登录已过期，请重新登录",
  "AUTH-AU-002": "登录已过期，请重新登录",
  "AUTH-AU-003": "无权限访问该项目",
  "AUTH-VL-001": "邮箱或密码错误",
  "TC-NF-001": "测试用例不存在",
  "TC-VL-001": "用例标题不能为空",
  "TC-VL-002": "至少需要一个测试步骤",
  "TC-CF-001": "该用例已被其他人修改，请刷新后重试",
  "TP-NF-001": "测试计划不存在",
  "TP-VL-001": "测试计划至少需要包含一个用例",
  "TP-CF-001": "已有执行中的测试，请先完成或取消当前执行",
  "EX-NF-001": "执行记录不存在",
  "EX-CF-001": "执行状态冲突，请刷新后重试",
  "FILE-VL-001": "不支持的文件类型",
  "FILE-VL-002": "文件大小超过限制（最大100MB）",
  "FILE-NF-001": "文件不存在",
  "RAG-IE-001": "RAG 检索服务异常，请稍后重试",
  "RAG-NF-001": "文件尚未完成索引，请等待索引完成",
  "AI-IE-001": "AI 服务暂时不可用，请稍后重试",
  "AI-RT-001": "AI 请求排队已满，请稍后重试",
  "AI-IE-002": "AI 输出校验失败，请重试",
  "PROJ-NF-001": "项目不存在",
  "PROJ-AU-001": "无权限操作该项目",
  "SYS-VL-001": "参数格式错误（无效的 UUID）",
  "SYS-IE-001": "服务器内部错误，请稍后重试",
};
```

---

## 11. 表单校验规格（Zod Schema）

```tsx
// 创建用例
const createTestCaseSchema = z.object({
  title: z.string().min(1, "标题不能为空").max(500),
  description: z.string().max(10000).optional(),
  module_id: z.string().uuid().optional().nullable(),
  steps: z.array(z.object({
    action: z.string().min(1, "操作不能为空"),
    expected: z.string().min(1, "预期结果不能为空"),
  })).min(1, "至少需要一个步骤"),
  priority: z.number().min(0).max(3),
  tags: z.array(z.string()).optional(),
});

// 创建计划
const createPlanSchema = z.object({
  name: z.string().min(1, "计划名称不能为空").max(255),
  description: z.string().max(5000).optional(),
  test_case_ids: z.array(z.string().uuid()).min(1, "至少选择一个用例"),
  assignments: z.array(z.object({
    test_case_id: z.string().uuid(),
    assigned_to: z.string().uuid(),
  })).optional(),
});

// 登录
const loginSchema = z.object({
  email: z.string().email("邮箱格式不正确"),
  password: z.string().min(1, "密码不能为空"),
});

// 创建项目
const createProjectSchema = z.object({
  name: z.string().min(1, "项目名称不能为空").max(255),
  description: z.string().max(5000).optional(),
});

// AI 生成
// 注意：file_id 为单文件 UUID，与设计稿多文件选择对齐需等后端支持多文件接口
const aiGenerateSchema = z.object({
  project_id: z.string().uuid(),
  file_id: z.string().uuid("请选择一个文件作为 AI 生成来源"),
  query: z.string().min(1, "查询描述不能为空").max(2000),
  options: z.object({
    count: z.number().min(1).max(50).default(10),
    priority: z.number().min(0).max(3).default(1),
    include_negative: z.boolean().default(true),
  }),
});
```

---

## 12. 配色系统

### 12.1 CSS 变量（浅色主题）

```css
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 221.2 83.2% 53.3%;       /* 蓝色系 */
  --primary-foreground: 210 40% 98%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --border: 214.3 31.8% 91.4%;
  --ring: 221.2 83.2% 53.3%;
  --radius: 0.5rem;
  /* ...其余变量参照 Shadcn/ui 默认值 */
}
```

### 12.2 语义化配色

| 语义 | Tailwind class | 用途 |
|------|---------------|------|
| primary | `bg-primary` | 主按钮、活跃链接 |
| success | `bg-green-500` | 通过、正常 |
| warning | `bg-amber-500` | 中优先级、需注意 |
| destructive | `bg-destructive` | 删除、失败 |
| info | `bg-sky-500` | 低优先级、辅助 |

### 12.3 优先级可视化

| Priority | 样式 | 文案 |
|----------|------|------|
| 0 | `bg-gray-100 text-gray-600` | 低 |
| 1 | `bg-blue-100 text-blue-600` | 中 |
| 2 | `bg-orange-100 text-orange-600` | 高 |
| 3 | `bg-red-100 text-red-600` | 紧急 |

### 12.4 暗色主题

已预留 `.dark` CSS 变量定义，MVP 不实现切换。后续可通过 Zustand 存储 theme 偏好。

---

## 13. 性能优化规格

### 13.1 代码分割

- 路由级懒加载：所有 pages/ 下的组件使用 `React.lazy`
- Recharts 按需加载：仅报表页面引入
- Vite 自动 chunk 分割：`node_modules` 独立 chunk

### 13.2 分页策略

- 使用 offset 分页：`page` + `page_size`
- 默认 `page_size=20`，最大 100
- TanStack Query `placeholderData: (prev) => prev` 翻页保持旧数据

### 13.3 列表虚拟化

MVP 阶段不启用。如果单页数据超过 100 条（`page_size` 增大时），使用 `@tanstack/react-virtual`。

---

## 14. 构建部署规格

### 14.1 环境变量

```bash
# .env（开发）
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=Heka AI 测试管理平台

# .env.production（生产）
VITE_API_BASE_URL=/api/v1
VITE_APP_TITLE=Heka AI 测试管理平台
```

### 14.2 Vite 代理配置

```ts
export default defineConfig({
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
```

### 14.3 Docker 构建

```dockerfile
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

---

## 15. 编码规范（关键条目）

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `TestCaseTable.tsx` |
| Hook 文件 | camelCase, `use` 前缀 | `usePagination.ts` |
| 服务文件 | camelCase | `testcase.ts` |
| 类型文件 | camelCase | `testcase.ts` |
| Store 文件 | camelCase | `auth.ts` |
| 组件导出 | 具名导出 | `export function TestCaseTable()` |
| 常量 | UPPER_SNAKE_CASE | `MAX_UPLOAD_SIZE` |
| Git 提交 | Conventional Commits | `feat: 添加用例列表页面` |

**组件设计原则**：
- 单一职责
- 表单用 React Hook Form（非受控），简单交互用 useState
- Props：回调用 `onXxx`，布尔用 `isXxx`/`hasXxx`
- 组合优于继承

---

## 16. 目录结构

```
heka-frontend/
├── public/
│   └── favicon.svg
├── src/
│   ├── app/                          # 应用入口
│   │   ├── App.tsx
│   │   ├── router.tsx                # 路由配置
│   │   └── providers.tsx             # QueryClientProvider, etc.
│   ├── pages/                        # 页面组件（路由级懒加载）
│   │   ├── Login/
│   │   ├── Projects/
│   │   ├── ProjectLayout/
│   │   ├── Overview/
│   │   ├── Modules/
│   │   ├── TestCases/
│   │   ├── TestPlans/
│   │   ├── Executions/
│   │   ├── Reports/
│   │   ├── Files/
│   │   ├── AIGenerate/
│   │   ├── AIAnalysis/
│   │   └── Settings/
│   ├── components/                   # 通用组件
│   │   ├── ui/                       # Shadcn/ui 基础组件
│   │   ├── layout/                   # 布局：AppLayout, ProjectSidebar, Breadcrumb
│   │   ├── module/                   # ModuleTree, ModuleForm
│   │   ├── testcase/                 # TestCaseTable, StepEditor, PriorityBadge, etc.
│   │   ├── plan/                     # ExecutionPanel, ResultSummary, AssignSelector
│   │   ├── collection/               # CollectionCard, AddCasesDialog
│   │   ├── file/                     # FileUploader, ChunkUploader, IndexStatusBadge
│   │   ├── ai/                       # GenerateForm, TaskProgress, ProviderStatus
│   │   ├── report/                   # PassRateChart, TrendChart, CoverageChart
│   │   └── shared/                   # ConfirmDialog, EmptyState, ErrorBoundary, Pagination
│   ├── hooks/                        # 自定义 Hooks
│   │   ├── useAuth.ts
│   │   ├── useProject.ts
│   │   ├── usePagination.ts
│   │   ├── useConfirm.ts
│   │   ├── useAITask.ts
│   │   ├── useFileIndexStatus.ts
│   │   ├── useUnsavedChanges.ts
│   │   └── useKeyboard.ts
│   ├── services/                     # API 服务层
│   │   ├── api.ts                    # Axios 实例 + 拦截器
│   │   ├── auth.ts
│   │   ├── project.ts
│   │   ├── module.ts
│   │   ├── testcase.ts
│   │   ├── collection.ts
│   │   ├── plan.ts
│   │   ├── execution.ts
│   │   ├── file.ts
│   │   ├── ai.ts
│   │   ├── report.ts
│   │   └── tag.ts
│   ├── stores/                       # Zustand Store
│   │   ├── auth.ts                   # 持久化 token
│   │   └── project.ts                # 持久化 currentProjectId
│   ├── types/                        # TypeScript 类型
│   │   ├── api.ts
│   │   ├── auth.ts
│   │   ├── project.ts
│   │   ├── module.ts
│   │   ├── testcase.ts
│   │   ├── plan.ts
│   │   ├── execution.ts
│   │   ├── file.ts
│   │   ├── ai.ts
│   │   ├── report.ts
│   │   ├── collection.ts
│   │   └── tag.ts
│   ├── lib/                          # 工具
│   │   ├── utils.ts                  # cn() 等
│   │   ├── constants.ts
│   │   ├── format.ts
│   │   ├── query-keys.ts
│   │   └── error-messages.ts
│   └── styles/
│       └── globals.css               # Tailwind 入口 + CSS 变量
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.ts
├── components.json
├── .env
├── .env.production
└── package.json
```

---

## 17. 实现路线图

### Phase 1：基础骨架（1 周）

> 交付目标：Vite 项目启动、登录可用、项目列表与概览展示

- 项目初始化（Vite + React + TypeScript + Tailwind + Shadcn/ui）
- 目录结构搭建
- Axios 实例 + 拦截器 + Token 刷新
- Zustand auth/project store
- 路由配置 + 路由守卫
- 登录页
- 项目列表页
- 项目概览页（统计卡片 + 最近活动）
- 项目设置页（基本信息编辑）
- 系统设置页（用户信息、主题切换入口）

### Phase 2：核心用例管理（2 周）

> 交付目标：用例 CRUD 全流程、模块树、批量操作

- 模块管理（树 + 拖拽）
- 用例列表（三栏布局、筛选、分页）
- 创建/编辑用例（步骤编辑器、表单校验）
- 用例详情
- 批量操作
- 标签管理
- 用例集合

### Phase 3：计划与执行（1.5 周）

> 交付目标：计划创建、执行流程、结果记录

- 测试计划列表 + 创建
- 计划详情（用例分配）
- 执行页面（左右分栏 + 键盘快捷键）
- 执行记录列表 + 详情
- 状态流转 UI 控制

### Phase 4：文件与 AI（1.5 周）

> 交付目标：文件上传、索引状态、AI 生成用例

- 文件列表 + 上传（含分块上传）
- 文件详情（索引状态、版本历史）
- Figma 链接
- AI 用例生成（异步轮询 + 进度展示）
- AI 智能分析

### Phase 5：报表与打磨（1 周）

> 交付目标：完整报表、命令面板、错误处理完善

- 5 个报表页面
- 命令面板 (Cmd+K)
- 空状态 + 加载状态完善
- 错误边界
- 响应式适配（1024px+）

---

**文档版本**：v1.1
**最后更新**：2026-05-18
