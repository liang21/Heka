# Heka 前端技术设计与 UI/UX 设计规范

> 版本：v1.1
> 日期：2025-05-18

---

## 1. 技术选型与决策依据

### 1.1 核心技术栈

| 分类 | 技术 | 版本 | 决策依据 |
|------|------|------|----------|
| 框架 | React | 18.x | 生态成熟，社区活跃，团队熟悉度高 |
| 语言 | TypeScript | 5.x | 类型安全，IDE 支持好，减少运行时错误 |
| 构建工具 | Vite | 5.x | 开发体验优于 CRA，HMR 速度快 |
| 样式方案 | Tailwind CSS | 3.x | 原子化 CSS，开发效率高，与 Shadcn/ui 深度集成 |
| 路由 | React Router | 6.x | 声明式路由，支持嵌套路由和数据加载 |
| 状态管理 | Zustand | 4.x | 轻量，无 Provider 嵌套，适合中小型项目 |
| 服务端状态 | TanStack Query | 5.x | 请求缓存、自动重试、乐观更新、分页支持 |
| UI 组件库 | Shadcn/ui | latest | 可定制性强，基于 Radix UI 无障碍支持好，非 npm 依赖 |
| 表单 | React Hook Form + Zod | latest | 非受控模式性能好，Zod 类型推导与 TS 完美集成 |
| HTTP 客户端 | Axios | 1.x | 拦截器机制，请求取消，超时控制 |
| 图表 | Recharts | 2.x | 基于 React 的声明式图表，轻量 |
| 拖拽 | @dnd-kit/core | 6.x | 现代化拖拽库，支持列表排序和看板 |
| 图标 | Lucide React | latest | 与 Shadcn/ui 默认集成，风格统一 |
| 日期处理 | date-fns | 3.x | 函数式、Tree-shakable、轻量 |

### 1.2 不选用的技术及原因

| 技术 | 不选用原因 |
|------|-----------|
| Redux / Redux Toolkit | 20-50 人团队规模不需要 Redux 的复杂度，Zustand 更轻量 |
| emotion / styled-components | Tailwind CSS 已满足需求，CSS-in-JS 增加运行时开销 |
| Next.js | MVP 阶段不需要 SSR，纯 SPA 部署更简单 |
| Ant Design / Material UI | 定制性差，样式覆盖成本高；Shadcn/ui 源码可控 |

---

## 2. 项目目录结构

```
heka-frontend/
├── public/
│   └── favicon.svg
├── src/
│   ├── app/                          # 应用入口
│   │   ├── App.tsx                   # 根组件
│   │   ├── router.tsx                # 路由配置
│   │   └── providers.tsx             # 全局 Provider
│   ├── pages/                        # 页面组件
│   │   ├── Login/
│   │   │   └── index.tsx
│   │   ├── Projects/                 # 项目列表（首页）
│   │   │   └── index.tsx
│   │   ├── ProjectLayout/            # 项目布局壳
│   │   │   └── index.tsx
│   │   ├── Overview/                 # 项目概览
│   │   │   └── index.tsx
│   │   ├── Modules/                  # 模块管理
│   │   │   ├── index.tsx
│   │   │   └── CreateModule.tsx
│   │   ├── TestCases/                # 测试用例
│   │   │   ├── index.tsx             # 用例列表
│   │   │   ├── CreateTestCase.tsx    # 创建用例
│   │   │   ├── TestCaseDetail.tsx    # 用例详情
│   │   │   └── Collections.tsx       # 用例集合
│   │   ├── TestPlans/                # 测试计划
│   │   │   ├── index.tsx             # 计划列表
│   │   │   ├── CreatePlan.tsx        # 创建计划
│   │   │   ├── PlanDetail.tsx        # 计划详情
│   │   │   └── ExecutePlan.tsx       # 执行页面
│   │   ├── Executions/               # 执行记录
│   │   │   ├── index.tsx
│   │   │   └── ExecutionDetail.tsx
│   │   ├── Reports/                  # 测试报告
│   │   │   ├── PlanReport.tsx
│   │   │   ├── Coverage.tsx
│   │   │   ├── Trend.tsx
│   │   │   ├── BugDistribution.tsx
│   │   │   └── Workload.tsx
│   │   ├── Files/                    # 文件管理
│   │   │   ├── index.tsx
│   │   │   └── FileDetail.tsx
│   │   ├── AIGenerate/               # AI 用例生成
│   │   │   └── index.tsx
│   │   ├── AIAnalysis/               # AI 智能分析
│   │   │   └── index.tsx
│   │   └── Settings/                 # 项目设置
│   │       └── index.tsx
│   ├── components/                   # 通用组件
│   │   ├── ui/                       # Shadcn/ui 基础组件
│   │   │   ├── button.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── input.tsx
│   │   │   ├── select.tsx
│   │   │   ├── table.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── toast.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── card.tsx
│   │   │   ├── skeleton.tsx
│   │   │   ├── tooltip.tsx
│   │   │   ├── command.tsx           # 命令面板 (Cmd+K)
│   │   │   └── ...
│   │   ├── layout/                   # 布局组件
│   │   │   ├── AppLayout.tsx         # 全局布局
│   │   │   ├── ProjectSidebar.tsx    # 项目侧边栏
│   │   │   ├── ProjectHeader.tsx     # 项目头部
│   │   │   └── Breadcrumb.tsx        # 面包屑
│   │   ├── module/                   # 模块组件
│   │   │   ├── ModuleTree.tsx        # 模块树
│   │   │   ├── ModuleForm.tsx        # 模块表单
│   │   │   └── ModuleTreeNode.tsx    # 树节点
│   │   ├── testcase/                 # 用例组件
│   │   │   ├── TestCaseTable.tsx     # 用例表格
│   │   │   ├── TestCaseCard.tsx      # 用例卡片
│   │   │   ├── TestCaseForm.tsx      # 用例表单
│   │   │   ├── StepEditor.tsx        # 步骤编辑器
│   │   │   ├── TestCaseFilters.tsx   # 筛选器
│   │   │   ├── BatchActions.tsx      # 批量操作
│   │   │   └── PriorityBadge.tsx     # 优先级标签
│   │   ├── plan/                     # 计划组件
│   │   │   ├── PlanCard.tsx
│   │   │   ├── ExecutionPanel.tsx    # 执行面板
│   │   │   ├── ResultSummary.tsx     # 结果摘要
│   │   │   └── AssignSelector.tsx    # 分配选择器
│   │   ├── collection/               # 集合组件
│   │   │   ├── CollectionCard.tsx
│   │   │   ├── AddCasesDialog.tsx
│   │   │   └── CollectionCasesList.tsx
│   │   ├── file/                     # 文件组件
│   │   │   ├── FileUploader.tsx      # 文件上传
│   │   │   ├── FileList.tsx          # 文件列表
│   │   │   ├── FileCard.tsx          # 文件卡片
│   │   │   └── IndexStatusBadge.tsx  # 索引状态
│   │   │   └── ChunkUploader.tsx     # 大文件分块上传（>10MB）
│   │   ├── ai/                       # AI 组件
│   │   │   ├── GenerateForm.tsx      # 生成表单
│   │   │   ├── TaskProgress.tsx      # 任务进度
│   │   │   ├── AnalysisResult.tsx    # 分析结果
│   │   │   └── ProviderStatus.tsx    # AI 服务状态
│   │   ├── report/                   # 报表组件
│   │   │   ├── PassRateChart.tsx
│   │   │   ├── TrendChart.tsx
│   │   │   ├── CoverageChart.tsx
│   │   │   └── BugDistributionChart.tsx
│   │   └── shared/                   # 共享组件
│   │       ├── ConfirmDialog.tsx     # 确认弹窗
│   │       ├── EmptyState.tsx        # 空状态
│   │       ├── ErrorBoundary.tsx     # 错误边界
│   │       ├── LoadingSpinner.tsx    # 加载动画
│   │       ├── PageSkeleton.tsx      # 页面骨架屏
│   │       ├── Pagination.tsx        # 分页器
│   │       ├── SearchInput.tsx       # 搜索输入
│   │       ├── SortableHeader.tsx    # 可排序表头
│   │       ├── StatusTag.tsx         # 状态标签
│   │       ├── UserAvatar.tsx        # 用户头像
│   │       └── TagInput.tsx          # 标签输入
│   ├── hooks/                        # 自定义 Hooks
│   │   ├── useAuth.ts                # 认证初始化（/auth/me）
│   │   ├── useProject.ts             # 当前项目
│   │   ├── usePagination.ts          # 分页逻辑
│   │   ├── useConfirm.ts             # 确认弹窗
│   │   ├── useAITask.ts              # AI 任务轮询
│   │   ├── useFileIndexStatus.ts     # 文件索引状态轮询
│   │   ├── useUnsavedChanges.ts      # 未保存修改拦截
│   │   └── useKeyboard.ts            # 键盘快捷键
│   ├── services/                     # API 服务层
│   │   ├── api.ts                    # Axios 实例和拦截器
│   │   ├── auth.ts                   # 认证 API
│   │   ├── project.ts                # 项目 API
│   │   ├── module.ts                 # 模块 API
│   │   ├── testcase.ts               # 用例 API
│   │   ├── collection.ts             # 集合 API
│   │   ├── plan.ts                   # 计划 API
│   │   ├── execution.ts              # 执行 API
│   │   ├── file.ts                   # 文件 API
│   │   ├── ai.ts                     # AI API
│   │   ├── report.ts                 # 报表 API
│   │   └── tag.ts                    # 标签 API
│   ├── stores/                       # Zustand 状态
│   │   ├── auth.ts                   # 认证状态
│   │   └── project.ts                # 项目切换状态
│   ├── types/                        # TypeScript 类型定义
│   │   ├── api.ts                    # API 通用类型
│   │   ├── auth.ts                   # 认证类型
│   │   ├── project.ts                # 项目类型
│   │   ├── module.ts                 # 模块类型
│   │   ├── testcase.ts               # 用例类型
│   │   ├── plan.ts                   # 计划类型
│   │   ├── execution.ts              # 执行类型
│   │   ├── file.ts                   # 文件类型
│   │   ├── ai.ts                     # AI 类型
│   │   └── report.ts                 # 报表类型
│   ├── lib/                          # 工具函数
│   │   ├── utils.ts                  # 通用工具（cn 等）
│   │   ├── constants.ts              # 常量定义
│   │   ├── format.ts                 # 格式化（日期、文件大小）
│   │   └── query-keys.ts            # TanStack Query Key 管理
│   └── styles/                       # 全局样式
│       └── globals.css               # Tailwind 入口 + CSS 变量
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.ts
├── components.json                   # Shadcn/ui 配置
├── .env                              # 环境变量
├── .env.production
└── package.json
```

---

## 3. 路由设计

### 3.1 路由表

```tsx
// src/app/router.tsx
const routes = [
  {
    path: "/login",
    element: <Login />,
  },
  {
    path: "/",
    element: <AppLayout />,         // 含侧边栏和顶部栏
    children: [
      {
        index: true,
        element: <Projects />,       // 项目列表
      },
      {
        path: "settings",
        element: <Settings />,       // 系统设置
      },
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

### 3.2 路由守卫

```tsx
// 未登录 → 跳转登录页
// 已登录但无项目 → 跳转项目列表
// 项目成员校验由后端 API 层处理（403 响应）
```

---

## 4. 状态管理策略

### 4.1 状态分层

```
┌─────────────────────────────────────────────────────────────┐
│                     URL 状态（Router）                       │
│  - 当前页面路径                                              │
│  - 分页参数（?page=1&page_size=20）                         │
│  - 筛选条件（?status=ready&priority=2）                     │
│  - 排序参数（?sort_by=created_at&sort_desc=true）           │
├─────────────────────────────────────────────────────────────┤
│                  服务端状态（TanStack Query）                 │
│  - 测试用例列表、详情                                        │
│  - 测试计划、执行记录                                        │
│  - 文件列表、项目列表                                        │
│  - AI 任务状态                                               │
│  - 报表数据                                                  │
├─────────────────────────────────────────────────────────────┤
│                  客户端状态（Zustand）                        │
│  - 认证信息（token, user）                                   │
│  - 当前选中项目 ID                                           │
│  - 侧边栏折叠状态                                            │
│  - 主题偏好                                                  │
├─────────────────────────────────────────────────────────────┤
│                  组件状态（useState/useReducer）              │
│  - 表单输入值                                                │
│  - 弹窗开关                                                  │
│  - 局部 UI 状态（选中行、编辑模式等）                        │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Zustand Store 设计

```tsx
// src/stores/auth.ts
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AuthState {
  token: string | null;
  user: User | null;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
      clearAuth: () => set({ token: null, user: null }),
      isAuthenticated: () => !!get().token,
    }),
    {
      name: "heka-auth",
      // 仅持久化 token，user 信息由 /auth/me 接口刷新
      partialize: (state) => ({ token: state.token }),
    }
  )
);
```

```tsx
// src/stores/project.ts
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface ProjectState {
  currentProjectId: string | null;
  setCurrentProject: (id: string) => void;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}

export const useProjectStore = create<ProjectState>()(
  persist(
    (set) => ({
      currentProjectId: null,
      setCurrentProject: (id) => set({ currentProjectId: id }),
      sidebarCollapsed: false,
      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
    }),
    {
      name: "heka-project",
      partialize: (state) => ({
        currentProjectId: state.currentProjectId,
        sidebarCollapsed: state.sidebarCollapsed,
      }),
    }
  )
);
```

### 4.3 TanStack Query Key 规范

```tsx
// src/lib/query-keys.ts
export const queryKeys = {
  // 项目
  projects: {
    all: ["projects"] as const,
    detail: (id: string) => ["projects", id] as const,
    members: (id: string) => ["projects", id, "members"] as const,
  },
  // 模块
  modules: {
    tree: (projectId: string) => ["projects", projectId, "modules"] as const,
  },
  // 测试用例
  testCases: {
    list: (projectId: string, filters: TestCaseFilters) =>
      ["projects", projectId, "testcases", filters] as const,
    detail: (id: string) => ["testcases", id] as const,
  },
  // 测试计划
  plans: {
    list: (projectId: string, filters?: PlanFilters) =>
      ["projects", projectId, "plans", filters] as const,
    detail: (id: string) => ["plans", id] as const,
  },
  // 执行记录
  executions: {
    list: (projectId: string, filters?: ExecutionFilters) =>
      ["projects", projectId, "executions", filters] as const,
    detail: (id: string) => ["executions", id] as const,
  },
  // 项目成员
  members: {
    list: (projectId: string) => ["projects", projectId, "members"] as const,
  },
  // 文件
  files: {
    list: (projectId: string, filters?: FileFilters) =>
      ["projects", projectId, "files", filters] as const,
    detail: (id: string) => ["files", id] as const,
    indexStatus: (id: string) => ["files", id, "index-status"] as const,
    reindexTask: (id: string) => ["files", id, "reindex"] as const,
  },
  // AI
  ai: {
    task: (taskId: string) => ["ai", "tasks", taskId] as const,
  },
  // 报表
  reports: {
    plan: (planId: string) => ["reports", "plan", planId] as const,
    coverage: (projectId: string) => ["reports", "coverage", projectId] as const,
    trend: (projectId: string, days: number) =>
      ["reports", "trend", projectId, days] as const,
    bugs: (projectId: string, params: BugReportParams) =>
      ["reports", "bugs", projectId, params] as const,
    workload: (userId: string, params: WorkloadParams) =>
      ["reports", "workload", userId, params] as const,
  },
  // 标签
  tags: {
    list: (projectId: string) => ["projects", projectId, "tags"] as const,
  },
  // 集合
  collections: {
    list: (projectId: string) => ["projects", projectId, "collections"] as const,
    detail: (id: string) => ["collections", id] as const,
    cases: (id: string, filters?: CollectionCaseFilters) =>
      ["collections", id, "cases", filters] as const,
  },
};
```

---

## 5. API 服务层设计

### 5.1 Axios 实例

```tsx
// src/services/api.ts
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
  timeout: 30000,
  headers: { "Content-Type": "application/json" },
});

// 是否正在刷新 Token（防止并发刷新）
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (token) resolve(token);
    else reject(error);
  });
  failedQueue = [];
}

// 请求拦截器：注入 Token
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器：统一错误处理 + Token 刷新
api.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const originalRequest = error.config;

    // 401 且未重试过 → 尝试刷新 Token
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // 其他请求等待刷新完成
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              resolve(api(originalRequest));
            },
            reject,
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // 用当前 Token 尝试刷新（后端 refresh 接口）
        const { data } = await axios.post(
          `${api.defaults.baseURL}/auth/refresh`,
          {},
          { headers: { Authorization: `Bearer ${useAuthStore.getState().token}` } }
        );
        const newToken = data.data.token;
        const user = useAuthStore.getState().user;

        useAuthStore.getState().setAuth(newToken, user!);
        processQueue(null, newToken);

        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return api(originalRequest);
      } catch {
        // 刷新失败 → 清除登录态，跳转登录页
        processQueue(error, null);
        useAuthStore.getState().clearAuth();
        window.location.href = "/login";
        return Promise.reject(error);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default api;
```

#### 应用启动初始化

```tsx
// src/hooks/useAuth.ts
// 应用启动时，若 localStorage 有 token，调用 /auth/me 校验并恢复用户信息
function useAuthInit() {
  const { token, setAuth, clearAuth } = useAuthStore();

  useEffect(() => {
    if (!token) return;

    api.get("/auth/me")
      .then((data) => {
        setAuth(token, data.data);
      })
      .catch(() => {
        clearAuth();
      });
  }, []);
}
```

### 5.2 API 服务示例

```tsx
// src/services/testcase.ts
import { api } from "./api";
import type { TestCase, TestCaseListResponse, CreateTestCaseRequest } from "@/types/testcase";

export const testcaseApi = {
  list: (params: TestCaseListParams) =>
    api.get<TestCaseListResponse>("/testcases", { params }),

  getById: (id: string) =>
    api.get<{ data: TestCase }>(`/testcases/${id}`),

  create: (data: CreateTestCaseRequest) =>
    api.post<{ data: { id: string; created_at: string } }>("/testcases", data),

  update: (id: string, data: UpdateTestCaseRequest) =>
    api.put<{ data: { id: string; updated_at: string } }>(`/testcases/${id}`, data),

  delete: (id: string) =>
    api.delete(`/testcases/${id}`),

  batchUpdateStatus: (ids: string[], status: string) =>
    api.put("/testcases/batch/status", { ids, status }),

  batchDelete: (ids: string[]) =>
    api.delete("/testcases/batch", { data: { ids } }),

  batchMove: (ids: string[], folderId: string) =>
    api.put("/testcases/batch/move", { ids, folder_id: folderId }),
};
```

### 5.3 Query Hook 封装模式

```tsx
// src/hooks/useTestCases.ts（模式示例，实际放在各页面或 hooks 目录）
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { testcaseApi } from "@/services/testcase";
import { queryKeys } from "@/lib/query-keys";

function useTestCases(projectId: string, filters: TestCaseFilters) {
  return useQuery({
    queryKey: queryKeys.testCases.list(projectId, filters),
    queryFn: () => testcaseApi.list({ project_id: projectId, ...filters }),
    placeholderData: (prev) => prev, // 分页保持旧数据
  });
}

function useCreateTestCase() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: testcaseApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] }); // 刷新列表
    },
  });
}
```

### 5.4 文件上传服务

小文件（≤10MB）直接上传，大文件使用分块上传：

```tsx
// src/services/file.ts
export const fileApi = {
  // 小文件：直接上传
  upload: (projectId: string, file: File) => {
    const formData = new FormData();
    formData.append("project_id", projectId);
    formData.append("file", file);
    return api.post("/files/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },

  // 大文件：分块上传
  uploadLarge: async (projectId: string, file: File, onProgress?: (pct: number) => void) => {
    const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    // 1. 初始化上传
    const { data } = await api.post("/files/upload/init", {
      project_id: projectId,
      file_name: file.name,
      file_size: file.size,
      total_chunks: totalChunks,
    });
    const { upload_id } = data;

    // 2. 逐块上传
    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE;
      const chunk = file.slice(start, start + CHUNK_SIZE);
      const chunkForm = new FormData();
      chunkForm.append("upload_id", upload_id);
      chunkForm.append("chunk_index", String(i));
      chunkForm.append("file", chunk);

      await api.post("/files/upload/chunk", chunkForm, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      onProgress?.(Math.round(((i + 1) / totalChunks) * 100));
    }

    // 3. 完成上传（后端合并 chunks）
    return api.post("/files/upload/complete", { upload_id });
  },

  // Figma 链接
  addFigmaLink: (projectId: string, figmaUrl: string, name: string) =>
    api.post("/files/figma", { project_id: projectId, figma_url: figmaUrl, name }),

  // 索引状态
  getIndexStatus: (fileId: string) =>
    api.get(`/files/${fileId}/index-status`),

  // 重新索引
  reindex: (fileId: string, force: boolean = false) =>
    api.post(`/files/${fileId}/reindex`, { force }),
};
```

```tsx
// src/components/file/FileUploader.tsx — 统一上传组件
// 内部自动判断文件大小：
// file.size > 10MB → fileApi.uploadLarge（分块上传，显示进度条）
// file.size <= 10MB → fileApi.upload（直接上传）
// 支持多文件选择，并发上传（最多 3 个同时）
// 上传完成后自动轮询索引状态
```

---

## 6. 页面布局与 UI 设计

### 6.1 全局布局层级

```
┌──────────────────────────────────────────────────────────────────┐
│  AppLayout                                                       │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  TopBar  [Logo] [项目切换器]  [搜索]  [用户头像▼]         │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌──────────┬───────────────────────────────────────────────┐    │
│  │ Sidebar  │  Content Area                                 │    │
│  │          │                                               │    │
│  │ 概览     │  ┌─────────────────────────────────────┐      │    │
│  │ 模块     │  │  PageHeader                         │      │    │
│  │ 用例     │  │  [面包屑] [标题] [操作按钮]         │      │    │
│  │ 计划     │  └─────────────────────────────────────┘      │    │
│  │ 执行     │  ┌─────────────────────────────────────┐      │    │
│  │ 报告     │  │  PageContent                        │      │    │
│  │ 文件     │  │                                     │      │    │
│  │ AI 生成  │  │                                     │      │    │
│  │ AI 分析  │  │                                     │      │    │
│  │ 设置     │  │                                     │      │    │
│  │          │  └─────────────────────────────────────┘      │    │
│  └──────────┴───────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

### 6.2 配色系统

基于 Shadcn/ui 的 CSS 变量系统，使用 HSL 定义：

```css
/* 浅色主题 */
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  --primary: 221.2 83.2% 53.3%;       /* 主色：蓝色系 */
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --border: 214.3 31.8% 91.4%;
  --ring: 221.2 83.2% 53.3%;
  --radius: 0.5rem;
}

/* 暗色主题（预留，MVP 不实现切换） */
.dark {
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --card: 222.2 84% 4.9%;
  --card-foreground: 210 40% 98%;
  --popover: 222.2 84% 4.9%;
  --popover-foreground: 210 40% 98%;
  --primary: 217.2 91.2% 59.8%;
  --primary-foreground: 222.2 47.4% 11.2%;
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  --accent: 217.2 32.6% 17.5%;
  --accent-foreground: 210 40% 98%;
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 210 40% 98%;
  --border: 217.2 32.6% 17.5%;
  --ring: 224.3 76.3% 48%;
}
```

### 6.3 语义化配色

| 语义 | 用途 | 色值参考 |
|------|------|----------|
| `primary` | 主操作按钮、活跃链接、选中态 | Blue 600 |
| `success` | 执行通过、状态正常 | Green 500 |
| `warning` | 中等优先级、需注意 | Amber 500 |
| `destructive` | 删除操作、执行失败 | Red 500 |
| `info` | 低优先级、辅助信息 | Sky 500 |

### 6.4 优先级可视化

| 优先级 | 标签样式 | 文案 |
|--------|----------|------|
| 低 (0) | 灰底灰字 | 低 |
| 中 (1) | 蓝底蓝字 | 中 |
| 高 (2) | 橙底橙字 | 高 |
| 紧急 (3) | 红底红字 | 紧急 |

### 6.5 状态可视化

| 状态 | 用例状态 | 计划状态 | 执行状态 |
|------|----------|----------|----------|
| 草稿 | 灰色圆点 | 灰色圆点 | - |
| 就绪/活跃 | 绿色圆点 | 蓝色圆点 | - |
| 已归档/完成 | 默认文字 | 绿色圆点 | - |
| 执行中 | - | 蓝色脉动圆点 | 蓝色脉动圆点 |
| 暂停 | - | 黄色圆点 | 黄色圆点 |
| 已取消 | - | 灰色圆点 | 灰色圆点 |
| 通过 | - | - | 绿色勾 |
| 失败 | - | - | 红色叉 |
| 阻塞 | - | - | 黄色叹号 |
| 跳过 | - | - | 灰色横线 |

---

## 7. 核心页面 UI 设计

### 7.1 登录页

```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│                   ┌──────────────────────┐                   │
│                   │      Heka Logo       │                   │
│                   │     AI 测试管理       │                   │
│                   ├──────────────────────┤                   │
│                   │  邮箱                │                   │
│                   │  ┌──────────────┐    │                   │
│                   │  │              │    │                   │
│                   │  └──────────────┘    │                   │
│                   │  密码                │                   │
│                   │  ┌──────────────┐    │                   │
│                   │  │              │    │                   │
│                   │  └──────────────┘    │                   │
│                   │                      │                   │
│                   │  ┌──────────────┐    │                   │
│                   │  │    登 录     │    │                   │
│                   │  └──────────────┘    │                   │
│                   └──────────────────────┘                   │
│                                                              │
└──────────────────────────────────────────────────────────────┘

说明：
- 居中卡片布局，宽度 400px
- 白色卡片 + 浅灰背景
- Logo + 标题 + 副标题
- 错误信息在表单下方内联显示
```

### 7.2 项目列表（首页）

```
┌──────────────────────────────────────────────────────────────┐
│ TopBar: [Heka] [用户名 ▼]                                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  我的项目                              [+ 新建项目]          │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │ 项目名称        │  │ 项目名称        │  │ 项目名称        │ │
│  │ 描述文本...     │  │ 描述文本...     │  │ 描述文本...     │ │
│  │                │  │                │  │                │ │
│  │ 5 成员 · 150用例│  │ 3 成员 · 80用例 │  │ 8 成员 · 200用例│ │
│  │ 创建于 5/10    │  │ 创建于 5/12    │  │ 创建于 5/15    │ │
│  └────────────────┘  └────────────────┘  └────────────────┘ │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐                      │
│  │ ...             │  │ ...             │                      │
│  └────────────────┘  └────────────────┘                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘

说明：
- 卡片网格布局，响应式：1/2/3/4 列
- 每张卡片显示：名称、描述、成员数、用例数、创建时间
- 卡片点击进入项目主页
- 空状态：引导创建第一个项目
```

### 7.3 测试用例列表（三栏布局）

```
┌──────────────────────────────────────────────────────────────────┐
│ TopBar: [Heka] [项目名 ▼] [搜索] [张三 ▼]                      │
├────────┬─────────────────────────────────────────────────────────┤
│ Sidebar│  面包屑: 项目名 / 测试用例                               │
│        │  ┌───────────────────────────────────────────────────┐  │
│ 概览    │  │ 搜索框  [状态▼] [优先级▼] [标签▼]  [+创建] [AI生成]│  │
│ 模块    │  ├───────────────────────────────────────────────────┤  │
│ 用例 ◉ │  │ ☐  标题        模块   状态   优先级  标签   创建人  │  │
│ 计划    │  │ ─────────────────────────────────────────────────│  │
│ 执行    │  │ ☐  用户登录成功  认证   就绪   中    登录   张三   │  │
│ 报告    │  │ ☐  用户注册       认证   草稿   高    注册   李四   │  │
│ 文件    │  │ ☐  密码重置       认证   就绪   中    密码   张三   │  │
│ AI生成  │  │ ...                                              │  │
│ AI分析  │  ├───────────────────────────────────────────────────┤  │
│ 设置    │  │  已选 3 项 [批量改状态] [批量移动] [批量删除]       │  │
│        │  │                          ← 1 2 3 ... 10 →         │  │
│        │  └───────────────────────────────────────────────────┘  │
│ ─────  │                                                         │
│ 模块树  │                                                         │
│ ▼ 认证  │                                                         │
│   登录  │                                                         │
│   注册  │                                                         │
│ ▼ 支付  │                                                         │
│   下单  │                                                         │
│        │                                                         │
└────────┴─────────────────────────────────────────────────────────┘

说明：
- 左侧固定宽度 260px 的模块树面板（可折叠）
- 中间是主表格区域
- 表格支持行选择、排序、列自定义
- 底部固定批量操作栏（选中时显示）
- 分页器在底部
- 点击模块树节点筛选对应模块下的用例
```

### 7.4 用例详情页

```
┌──────────────────────────────────────────────────────────────────┐
│ ← 返回列表  /  认证模块 / 用户登录成功                           │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  用户登录成功                                [编辑] [更多▼]       │
│  就绪 · 中优先级 · 张三创建于 2025-05-15                         │
│  标签: [登录] [功能测试]                                          │
│                                                                  │
│  ── 描述 ────────────────────────────────────────────────────── │
│  验证用户使用正确的用户名和密码可以成功登录                        │
│                                                                  │
│  ── 测试步骤 ────────────────────────────────────────────────── │
│                                                                  │
│  ┌─────┬───────────────────┬──────────────────────┐             │
│  │  #  │  操作              │  预期结果             │             │
│  ├─────┼───────────────────┼──────────────────────┤             │
│  │  1  │  打开登录页面      │  显示登录表单         │             │
│  │  2  │  输入正确的用户名  │  输入框显示输入内容   │             │
│  │  3  │  点击登录按钮      │  登录成功，跳转到首页 │             │
│  └─────┴───────────────────┴──────────────────────┘             │
│                                                                  │
│  ── 执行历史 ────────────────────────────────────────────────── │
│  ┌──────────┬──────┬────────┬────────┬──────┐                   │
│  │ 执行时间 │ 计划  │ 执行人  │ 结果   │ 备注 │                   │
│  ├──────────┼──────┼────────┼────────┼──────┤                   │
│  │ 5/15     │ V1.0 │ 张三    │ ✓ 通过 │      │                   │
│  │ 5/14     │ V1.0 │ 李四    │ ✗ 失败 │ ...  │                   │
│  └──────────┴──────┴────────┴────────┴──────┘                   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

说明：
- 顶部返回导航 + 面包屑
- 标题区：标题 + 状态标签 + 优先级标签 + 操作按钮
- 编辑模式：点击"编辑"后标题、描述、步骤变为可编辑态
- 步骤表格支持拖拽排序（编辑模式下）
- 执行历史从后端查询关联数据
```

### 7.5 测试计划执行页面

```
┌──────────────────────────────────────────────────────────────────┐
│ ← 返回计划  V1.0 回归测试                                        │
│ 活跃 · 50 用例 · 执行中                                          │
├───────────────────────┬──────────────────────────────────────────┤
│                       │                                          │
│  用例列表              │  执行面板                                │
│  ┌─────────────────┐  │                                          │
│  │  ✓ 用户登录成功  │  │  用户登录成功                            │
│  │  ✗ 用户注册      │  │  优先级: 高  分配给: 张三                │
│  │  → 密码重置      │  │                                          │
│  │  ○ 找回密码      │  │  ── 步骤 ──────────────────────────     │
│  │  ○ 修改密码      │  │  1. 打开登录页面                         │
│  │  ...            │  │     预期: 显示登录表单                    │
│  │                 │  │  2. 输入正确的用户名和密码                 │
│  │                 │  │     预期: 输入框显示输入内容               │
│  │                 │  │  3. 点击登录按钮                          │
│  │                 │  │     预期: 登录成功，跳转到首页            │
│  │                 │  │                                          │
│  │                 │  │  ── 执行结果 ────────────────────────    │
│  │                 │  │                                          │
│  │                 │  │  [✓ 通过]  [✗ 失败]  [⚠ 阻塞]  [○ 跳过] │
│  │                 │  │                                          │
│  │                 │  │  Bug ID: [____________]                  │
│  │                 │  │  备注:   [____________]                  │
│  │                 │  │                                          │
│  │                 │  │  [确认并下一个 →]                         │
│  │                 │  │                                          │
│  ├─────────────────┤  ├──────────────────────────────────────────┤
│  │ 进度: 30/50 60% │  │  通过 25  失败 3  阻塞 1  跳过 1       │
│  └─────────────────┘  └──────────────────────────────────────────┘
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

说明：
- 左右分栏，左 40% 右 60%
- 左侧用例列表，当前执行的用例高亮
- 右侧执行面板，展示步骤详情
- 底部四个快捷按钮标记结果
- 失败时展开 Bug ID 输入框
- 键盘快捷键：P=通过, F=失败, B=阻塞, S=跳过, Enter=下一个
```

### 7.6 AI 用例生成页面

```
┌──────────────────────────────────────────────────────────────────┐
│ AI 生成测试用例                                                   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ── 选择来源文件 ─────────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  📄 需求文档.pdf   100 页   已索引 ✓                     │   │
│  │  📄 接口设计.docx   30 页   已索引 ✓                     │   │
│  │  📄 UI 设计稿      Figma   已索引 ✓                      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ── 生成设置 ─────────────────────────────────────────────────── │
│  查询描述:  [用户登录功能的测试用例                        ]      │
│  生成数量:  [10 ▼]    优先级: [中 ▼]    包含反向用例: [✓]       │
│                                                                  │
│  [✨ 生成测试用例]                                                │
│                                                                  │
│  ── 生成结果 ─────────────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  ████████████████████░░░░  80%  正在生成...              │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  已生成 8/10:                                                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ 用户登录成功 - 正确凭据登录                    [预览]  │   │
│  │ ☑ 用户登录失败 - 错误密码                        [预览]  │   │
│  │ ☑ 用户登录失败 - 账号锁定                        [预览]  │   │
│  │ ☑ 用户登录失败 - 空密码                          [预览]  │   │
│  │ ...                                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  [全选] [取消全选]                        [批量创建选中用例 →]    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

说明：
- 文件选择列表（已索引文件）
- 表单：查询描述 + 生成选项
- 进度条通过轮询实时更新
- 生成结果列表支持逐条预览和勾选
- 批量创建按钮
```

### 7.7 模块管理页面

```
┌──────────────────────────────────────────────────────────────────┐
│ TopBar: [Heka] [项目名 ▼] [搜索] [张三 ▼]                      │
├────────┬─────────────────────────────────────────────────────────┤
│ Sidebar│  面包屑: 项目名 / 模块管理                               │
│        │  ┌───────────────────────────────────────────────────┐  │
│ 概览    │  │ [+ 新建模块]                  [展开全部] [折叠]  │  │
│ 模块 ◉ │  ├───────────────────────────────────────────────────┤  │
│ 用例    │  │                                                   │  │
│ 计划    │  │  ▼ 📁 认证模块                    [编辑] [删除]  │  │
│ 执行    │  │    │  📄 登录模块                  [编辑] [删除]  │  │
│ 报告    │  │    │  📄 注册模块                  [编辑] [删除]  │  │
│ 文件    │  │    └  📄 密码重置                  [编辑] [删除]  │  │
│ AI生成  │  │                                                   │  │
│ AI分析  │  │  ▼ 📁 支付模块                    [编辑] [删除]  │  │
│ 设置    │  │    │  📄 下单流程                  [编辑] [删除]  │  │
│        │  │    └  📄 退款流程                  [编辑] [删除]  │  │
│        │  │                                                   │  │
│        │  │  📁 未分类                         [编辑] [删除]  │  │
│        │  │                                                   │  │
│        │  └───────────────────────────────────────────────────┘  │
│        │                                                         │
└────────┴─────────────────────────────────────────────────────────┘

说明：
- 树形结构展示模块层级
- 支持拖拽排序和移动（使用 @dnd-kit）
- 每个节点显示：名称 + 用例数 + 操作按钮
- 新建/编辑弹出 Dialog 表单（名称、描述、父模块选择）
- 删除模块时确认：「该模块下有 N 个用例，用例将移至"未分类"」
```

### 7.8 测试计划详情页

```
┌──────────────────────────────────────────────────────────────────┐
│ ← 返回列表  /  V1.0 回归测试                                     │
│ 活跃 · 创建于 2025-05-15 · 创建人 张三                           │
├──────────────────────────────────────────────────────────────────┤
│  [开始执行]  [暂停]  [编辑]  [更多▼]                              │
│                                                                   │
│  ── 基本信息 ─────────────────────────────────────────────────── │
│  描述: 本次回归测试覆盖认证模块和支付模块的所有核心功能            │
│                                                                   │
│  ── 用例列表 ─────────────────────────────────────────────────── │
│  已添加 50 个用例                              [+ 添加用例]       │
│  ┌──────────────────┬──────┬──────┬──────────┬──────────┐       │
│  │ 标题              │ 优先级│ 状态  │ 分配给   │ 操作     │       │
│  ├──────────────────┼──────┼──────┼──────────┼──────────┤       │
│  │ 用户登录成功      │ 高   │ 就绪 │ 张三      │ [移除]   │       │
│  │ 用户注册          │ 高   │ 就绪 │ 李四      │ [移除]   │       │
│  │ 密码重置          │ 中   │ 就绪 │ 张三      │ [移除]   │       │
│  │ ...              │      │      │           │          │       │
│  └──────────────────┴──────┴──────┴──────────┴──────────┘       │
│                                                                   │
│  ── 执行历史 ─────────────────────────────────────────────────── │
│  ┌──────────┬────────┬──────┬──────────────┬──────────┐         │
│  │ 执行名称  │ 执行人  │ 状态  │ 进度         │ 时间     │         │
│  ├──────────┼────────┼──────┼──────────────┼──────────┤         │
│  │ 第1轮    │ 张三    │ 完成 │ 50/50 100%   │ 5/15     │         │
│  │ 第2轮    │ 李四    │ 进行 │ 30/50 60%    │ 5/16     │ [继续]  │
│  └──────────┴────────┴──────┴──────────────┴──────────┘         │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘

说明：
- 顶部操作按钮根据计划状态动态显示（参考第 9.2 节状态流转）
- 用例列表支持批量添加、拖拽排序、分配执行人
- 执行历史显示所有执行记录，进行中的可点击继续
```

### 7.9 文件详情页

```
┌──────────────────────────────────────────────────────────────────┐
│ ← 返回文件列表                                                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  📄 需求规格说明书.docx                                          │
│  Word · 2.3MB · 上传于 2025-05-15 · 上传人 张三                   │
│                                                                   │
│  ── 索引状态 ─────────────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  状态: ✅ 已索引                                          │   │
│  │  分块数: 156  ·  索引时间: 2025-05-15 14:30:00           │   │
│  │  向量模型: text-embedding-3-small                        │   │
│  │                                        [重新索引 ▼]      │   │
│  │                                                         │   │
│  │  重新索引选项:                                           │   │
│  │  [ ] 强制重新索引（忽略缓存）                             │   │
│  │                                       [确认重新索引]     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ── 版本历史 ─────────────────────────────────────────────────── │
│  ┌──────────┬──────┬────────┬──────────┐                        │
│  │ 版本      │ 大小  │ 上传人  │ 时间     │                        │
│  ├──────────┼──────┼────────┼──────────┤                        │
│  │ v3       │ 2.3MB│ 张三    │ 5/15     │ ← 当前               │
│  │ v2       │ 2.1MB│ 李四    │ 5/12     │                       │
│  │ v1       │ 1.8MB│ 张三    │ 5/10     │                       │
│  └──────────┴──────┴────────┴──────────┘                        │
│                                                                   │
│  ── 预览 ────────────────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  文档内容预览区域（前 500 字）                            │   │
│  │  ...                                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  [下载文件]  [删除文件]                                           │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘

说明：
- 索引状态实时显示（轮询 /files/:id/index-status）
- 重新索引按钮带确认和「强制」选项
- 版本历史支持查看和回退
- 删除文件需要输入文件名确认
```

---

## 8. 交互规范

### 8.1 键盘快捷键

| 快捷键 | 上下文 | 动作 |
|--------|--------|------|
| `Cmd/Ctrl + K` | 全局 | 打开命令面板 |
| `Cmd/Ctrl + N` | 用例列表 | 新建测试用例 |
| `Cmd/Ctrl + S` | 编辑页面 | 保存 |
| `Escape` | 弹窗/编辑态 | 关闭/取消 |
| `P` | 执行页面 | 标记通过 |
| `F` | 执行页面 | 标记失败 |
| `B` | 执行页面 | 标记阻塞 |
| `S` | 执行页面 | 标记跳过 |
| `Enter` | 执行页面 | 确认并下一个 |
| `?` | 全局 | 显示快捷键帮助 |
| `Cmd/Ctrl + /` | 侧边栏 | 折叠/展开 |

### 8.2 命令面板（Cmd+K）

命令面板支持以下操作：

| 分类 | 操作 | 说明 |
|------|------|------|
| 导航 | 跳转项目列表 | `/` |
| 导航 | 跳转用例列表 | 当前项目 → 用例 |
| 导航 | 跳转测试计划 | 当前项目 → 计划 |
| 导航 | 跳转执行记录 | 当前项目 → 执行 |
| 导航 | 跳转文件管理 | 当前项目 → 文件 |
| 导航 | 跳转 AI 生成 | 当前项目 → AI 生成 |
| 搜索 | 搜索用例 | 按标题模糊搜索 |
| 搜索 | 搜索计划 | 按名称模糊搜索 |
| 操作 | 新建用例 | 快捷创建 |
| 操作 | 新建计划 | 快捷创建 |
| 操作 | 切换项目 | 项目列表选择 |
| 操作 | 切换主题 | 浅色/深色（预留） |

实现基于 Shadcn/ui 的 `Command` 组件（cmdk），搜索通过本地缓存 + API 混合实现。

### 8.3 确认弹窗规范

| 操作 | 确认方式 | 说明 |
|------|----------|------|
| 删除单个用例 | ConfirmDialog | "确定删除该用例？此操作不可恢复" |
| 批量删除 | ConfirmDialog + 输入数量 | "输入删除数量以确认：__" |
| 删除项目 | ConfirmDialog + 输入项目名 | "输入项目名称以确认删除" |
| 状态变更 | 直接操作 | 通过 Toast 提示 |
| 离开编辑页面 | unsaved_changes 提示 | "有未保存的修改，确定离开？" |

### 8.4 加载状态

| 场景 | 处理方式 |
|------|----------|
| 页面首次加载 | 骨架屏（Skeleton） |
| 列表翻页 | 保持旧数据 + 表格顶部进度条 |
| 提交表单 | 按钮禁用 + Spinner |
| 删除操作 | 乐观更新 + 失败回滚 |
| AI 生成 | 进度条 + SSE 实时更新 |

### 8.5 空状态设计

| 页面 | 空状态文案 | 操作按钮 |
|------|-----------|----------|
| 项目列表 | "还没有项目，创建一个开始吧" | + 新建项目 |
| 用例列表 | "该模块下还没有用例" | + 创建用例 / AI 生成 |
| 执行记录 | "暂无执行记录" | - |
| 文件列表 | "还没有上传文件" | + 上传文件 |
| 报表 | "暂无数据，执行测试后生成报告" | - |

### 8.6 未保存修改提示

编辑页面（用例编辑、计划编辑等）在用户有未保存修改时，离开页面需要确认。

**实现方式**：

```tsx
// 使用 React Router v6 的 useBlocker
import { useBlocker } from "react-router-dom";

function useUnsavedChanges(hasChanges: boolean) {
  const blocker = useBlocker(hasChanges);

  useEffect(() => {
    // 浏览器关闭/刷新时的提示
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasChanges) {
        e.preventDefault();
      }
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [hasChanges]);

  return blocker;
}

// 在编辑页面中使用
function TestCaseEditPage() {
  const formMethods = useForm({...});
  const isDirty = formMethods.formState.isDirty;
  const blocker = useUnsavedChanges(isDirty);

  if (blocker.state === "blocked") {
    return (
      <ConfirmDialog
        open
        title="未保存的修改"
        description="有未保存的修改，确定离开？"
        onConfirm={() => blocker.proceed()}
        onCancel={() => blocker.reset()}
      />
    );
  }
  // ...
}
```

---

## 9. 状态流转与 UI 控制

### 9.1 测试用例状态流转

```
  ┌───────┐    标记就绪    ┌───────┐    归档     ┌──────────┐
  │ draft │ ────────────→ │ ready │ ──────────→ │ archived │
  │ 草稿  │ ←──────────── │ 就绪  │ ←────────── │ 已归档   │
  └───────┘    改回草稿    └───────┘    恢复     └──────────┘
```

| 当前状态 | 可用操作 | UI 按钮 |
|---------|---------|---------|
| `draft` | 编辑、删除、标记就绪、归档 | [标记就绪] [归档] [编辑] [删除] |
| `ready` | 编辑、删除、改回草稿、归档 | [改回草稿] [归档] [编辑] [删除] |
| `archived` | 恢复就绪、删除 | [恢复] [删除] |

**前端预校验**：按钮可见性根据当前状态控制，非法状态变更后端会返回 400。

### 9.2 测试计划状态流转

```
  ┌───────┐     开始执行    ┌────────┐    暂停    ┌────────┐
  │ draft │ ─────────────→ │ active │ ────────→ │ paused │
  │ 草稿  │                 │ 活跃   │ ←──────── │ 已暂停 │
  └───────┘                 └────┬───┘    恢复    └───┬────┘
       │                     │                    │
       │   取消              │ 完成               │ 取消
       ↓                     ↓                    ↓
  ┌──────────┐         ┌───────────┐        ┌──────────┐
  │ cancelled│         │ completed │        │ cancelled│
  │ 已取消   │         │ 已完成    │        │ 已取消   │
  └──────────┘         └───────────┘        └──────────┘
```

| 当前状态 | 可用操作 | UI 按钮 |
|---------|---------|---------|
| `draft` | 编辑、添加/移除用例、开始执行、取消 | [开始执行] [编辑] [取消计划] |
| `active` | 暂停执行、查看执行进度 | [暂停] [查看执行] |
| `paused` | 恢复执行、取消 | [恢复执行] [取消计划] |
| `completed` | 查看报告、查看执行详情 | [查看报告] |
| `cancelled` | 仅查看 | — |

### 9.3 执行并发控制

后端限制：同一测试计划同时只能有一个 `in_progress` 的执行。

**前端处理**：
- 点击「开始执行」前，先查询是否有未完成的执行
- 如果有，弹窗提示：「当前有进行中的执行：{执行名称}（已完成 {n}/{total}），请先完成或取消」
- 提供「继续上次执行」和「取消并新建」两个选项
- 后端兜底：违反规则返回 `TP-CF-001` (409)

### 9.4 乐观锁（版本控制）

测试用例使用 `version` 字段做乐观锁，防止并发编辑覆盖：

**前端处理流程**：
1. 进入编辑页时记录当前 `version`
2. 提交更新时携带 `version` 字段
3. 收到 `TC-CF-001` (409) 错误时：
   - 弹窗提示「该用例已被其他人修改」
   - 提供「查看最新版本」和「强制覆盖」两个选项
   - 「查看最新版本」：丢弃本地修改，重新加载
   - 「强制覆盖」：用最新 version 重新提交

---

## 10. 异步任务与实时通信设计

### 9.1 设计决策

> **AI 任务进度采用轮询模式（2s 间隔），不使用 SSE。**
>
> 原因：浏览器原生 `EventSource` API 不支持自定义 HTTP Header，无法携带 Bearer Token。
> 对内部工具（20-50 人）场景，2 秒轮询延迟完全可接受，且实现更简单、断线恢复更可靠。
> 后端 SSE 端点 `GET /ai/tasks/{id}/events` 保留，后续如需增强实时性可改用 fetch-based SSE。

### 9.2 AI 任务轮询 Hook

AI 生成采用 **异步任务 + 轮询** 模式：

```tsx
// src/hooks/useAITask.ts
import { useQuery } from "@tanstack/react-query";
import { aiApi } from "@/services/ai";
import { queryKeys } from "@/lib/query-keys";

// 流程：提交生成请求 → 获得 task_id → 启动轮询 → 任务完成后停止
function useAITask(taskId: string | null) {
  return useQuery({
    queryKey: queryKeys.ai.task(taskId!),
    queryFn: () => aiApi.getTaskStatus(taskId!),
    enabled: !!taskId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === "completed" || status === "failed") return false;
      return 2000; // 2 秒轮询
    },
    // 任务完成后保留数据 30 分钟
    staleTime: 30 * 60 * 1000,
  });
}
```

### 9.3 文件索引状态轮询

文件上传后的 RAG 索引同样采用轮询：

```tsx
// src/hooks/useFileIndexStatus.ts
function useFileIndexStatus(fileId: string | null) {
  return useQuery({
    queryKey: queryKeys.files.indexStatus(fileId!),
    queryFn: () => fileApi.getIndexStatus(fileId!),
    enabled: !!fileId,
    refetchInterval: (query) => {
      const status = query.state.data?.index_status;
      if (status === "completed" || status === "failed") return false;
      return 3000; // 3 秒轮询
    },
  });
}
```

### 9.4 未来升级路径

如后续需要实时性更强的场景（多人协作通知等），可升级为 fetch-based SSE：

```tsx
// 预留方案：fetch + ReadableStream 实现 SSE
// 后端端点 GET /ai/tasks/{id}/events 保持不变
async function* streamSSE(url: string, token: string) {
  const resp = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const reader = resp.body!.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    // 解析 SSE 格式的 data: 行
    const lines = buffer.split("\n");
    buffer = lines.pop() || "";
    for (const line of lines) {
      if (line.startsWith("data: ")) {
        yield JSON.parse(line.slice(6));
      }
    }
  }
}
```

---

## 11. 表单设计规范

### 10.1 Zod Schema 与类型推导

```tsx
// src/types/testcase.ts
import { z } from "zod";

export const createTestCaseSchema = z.object({
  title: z.string().min(1, "标题不能为空").max(500, "标题最多500字"),
  description: z.string().max(10000, "描述最多10000字").optional(),
  module_id: z.string().uuid().optional().nullable(),
  steps: z.array(
    z.object({
      action: z.string().min(1, "操作不能为空"),
      expected: z.string().min(1, "预期结果不能为空"),
    })
  ).min(1, "至少需要一个步骤"),
  priority: z.number().min(0).max(3),
  tags: z.array(z.string()).optional(),
});

export type CreateTestCaseInput = z.infer<typeof createTestCaseSchema>;
```

### 10.2 核心类型定义

```tsx
// src/types/api.ts — 通用 API 类型
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
```

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

```tsx
// src/types/testcase.ts
type TestCaseStatus = "draft" | "ready" | "archived";
type Priority = 0 | 1 | 2 | 3; // 0=低 1=中 2=高 3=紧急

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
  version: number; // 乐观锁版本号
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

interface PlanAssignment {
  test_case_id: string;
  assigned_to: string;
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
  summary: {
    total_bugs: number;
    by_source: Record<string, number>;
  };
  top_failed_cases: ExecutionResult[];
}

interface WorkloadReport {
  user_id: string;
  user_name: string;
  summary: {
    total_executed: number;
    passed: number;
    failed: number;
  };
  daily_breakdown: Array<{
    date: string;
    executed: number;
    passed: number;
    failed: number;
  }>;
}
```

```tsx
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

```tsx
// src/types/tag.ts
interface Tag {
  id: string;
  project_id: string;
  name: string;
  color: string;
  usage_count: number;
}
```

### 10.2 表单布局规范

- 标签在输入框上方，左对齐
- 必填项用 `*` 标记
- 错误信息在输入框下方红色文字
- 长表单使用分组卡片
- 底部固定操作栏（取消 + 提交）

---

## 12. 错误处理

### 11.1 错误边界

```tsx
// 页面级 ErrorBoundary：捕获渲染错误，显示友好错误页面
// 组件级 ErrorBoundary：局部组件出错时显示降级 UI
```

### 11.2 API 错误处理策略

| HTTP 状态码 | 处理方式 |
|-------------|----------|
| 400 | 表单字段高亮 + 错误信息 |
| 401 | 清除登录态 → 跳转登录页 |
| 403 | Toast 提示"无权限" |
| 404 | 显示 404 页面 |
| 409 | Toast 提示冲突信息（如版本冲突） |
| 429 | Toast 提示"操作过于频繁，请稍后" |
| 500 | Toast 提示"服务器错误" |
| 网络错误 | Toast 提示"网络连接失败" |

### 11.3 错误码映射

```tsx
// src/lib/error-messages.ts
const errorMessages: Record<string, string> = {
  // 认证
  "AUTH-AU-001": "登录已过期，请重新登录",
  "AUTH-AU-002": "登录已过期，请重新登录",
  "AUTH-AU-003": "无权限访问该项目",
  "AUTH-VL-001": "邮箱或密码错误",
  // 测试用例
  "TC-NF-001": "测试用例不存在",
  "TC-VL-001": "用例标题不能为空",
  "TC-VL-002": "至少需要一个测试步骤",
  "TC-CF-001": "该用例已被其他人修改，请刷新后重试",
  // 测试计划
  "TP-NF-001": "测试计划不存在",
  "TP-VL-001": "测试计划至少需要包含一个用例",
  "TP-CF-001": "已有执行中的测试，请先完成或取消当前执行",
  // 执行
  "EX-NF-001": "执行记录不存在",
  "EX-CF-001": "执行状态冲突，请刷新后重试",
  // 文件
  "FILE-VL-001": "不支持的文件类型",
  "FILE-VL-002": "文件大小超过限制（最大100MB）",
  "FILE-NF-001": "文件不存在",
  // RAG
  "RAG-IE-001": "RAG 检索服务异常，请稍后重试",
  "RAG-NF-001": "文件尚未完成索引，请等待索引完成",
  // AI
  "AI-IE-001": "AI 服务暂时不可用，请稍后重试",
  "AI-RT-001": "AI 请求排队已满，请稍后重试",
  "AI-IE-002": "AI 输出校验失败，请重试",
  // 项目
  "PROJ-NF-001": "项目不存在",
  "PROJ-AU-001": "无权限操作该项目",
  // 系统
  "SYS-VL-001": "参数格式错误（无效的 UUID）",
  "SYS-IE-001": "服务器内部错误，请稍后重试",
};
```

---

## 13. 性能优化

### 13.1 代码分割

```tsx
// 路由级懒加载
const TestCases = lazy(() => import("@/pages/TestCases"));
const TestPlans = lazy(() => import("@/pages/TestPlans"));
const AIGenerate = lazy(() => import("@/pages/AIGenerate"));
// ...
```

### 13.2 列表虚拟化

测试用例列表数据量可能较大（>1000），使用 `@tanstack/react-virtual` 虚拟滚动：

```tsx
// 仅在单页数据 > 100 条时启用虚拟滚动
// 默认分页模式（page_size=20）不需要虚拟滚动
```

### 13.3 缓存策略

| 数据类型 | TanStack Query 配置 | 说明 |
|----------|-------------------|------|
| 用户信息 | staleTime: 30min | 配合 /auth/me 刷新 |
| 项目列表 | staleTime: 5min | 不频繁变化 |
| 模块树 | staleTime: 10min | 结构稳定 |
| 用例列表 | staleTime: 2min | 可能被多人修改 |
| 用例详情 | staleTime: 5min | |
| 文件列表 | staleTime: 2min | |
| AI 任务 | staleTime: 0 | 轮询模式 |
| 报表数据 | staleTime: 10min | 历史数据不变化 |

### 13.4 分页策略

后端使用传统 offset 分页（`page` + `page_size`），前端统一使用此方案：

```tsx
// 分页参数
interface PaginationParams {
  page: number;      // 从 1 开始
  page_size: number; // 默认 20，最大 100
}

// 分页响应
interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
```

> 后端 `database-performance-spec.md` 建议大表使用 cursor 分页，
> 但 MVP 阶段数据量有限（单项目用例 < 10000），offset 分页性能够用。
> 如后续出现性能问题，后端可切换为 cursor 分页，前端仅需修改 `Pagination` 组件。

### 13.5 图片/文件优化

- 用户头像：使用缩略图 API
- 文件列表：仅显示元数据，不预加载文件内容
- 报表图表：按需加载 Recharts

---

## 14. 响应式设计

### 14.1 断点定义

| 断点 | 宽度 | 布局调整 |
|------|------|----------|
| `sm` | ≥640px | 移动端横屏 |
| `md` | ≥768px | 平板 |
| `lg` | ≥1024px | 小桌面（默认设计基准） |
| `xl` | ≥1280px | 桌面 |
| `2xl` | ≥1536px | 大桌面 |

### 14.2 响应式策略

MVP 阶段优先桌面端（1280px+），确保在 1024px 以上可用。

| 组件 | <1024px | ≥1024px |
|------|---------|---------|
| 侧边栏 | 自动折叠为图标模式 | 展开完整侧边栏 |
| 用例列表 | 隐藏模块树 | 三栏布局 |
| 执行页面 | 上下分栏 | 左右分栏 |
| 表格 | 横向滚动 | 正常展示 |

---

## 15. 国际化预留

MVP 阶段仅支持中文，但代码层面预留国际化能力：

- 文案集中在常量文件中，不硬编码在组件内
- 日期格式统一使用 `yyyy-MM-dd HH:mm`
- 数字使用中文习惯（千分位逗号）

---

## 16. 可访问性（a11y）

基于 Shadcn/ui（底层 Radix UI），默认支持：

- 键盘导航：所有交互组件支持 Tab / Enter / Escape
- ARIA 属性：Dialog、Dropdown、Tabs 等自动添加
- 焦点管理：弹窗打开时聚焦，关闭后恢复焦点
- 颜色对比度：满足 WCAG 2.1 AA 标准

---

## 17. 开发规范

### 16.1 命名约定

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `TestCaseTable.tsx` |
| Hook 文件 | camelCase, use 前缀 | `usePagination.ts` |
| 服务文件 | camelCase | `testcase.ts` |
| 类型文件 | camelCase | `testcase.ts` |
| Store 文件 | camelCase | `auth.ts` |
| CSS 类名 | Tailwind 原子类 | 直接使用，不自定义类名 |
| 组件导出 | 具名导出 | `export function TestCaseTable()` |
| 常量 | UPPER_SNAKE_CASE | `MAX_UPLOAD_SIZE` |

### 16.2 组件设计原则

- 单一职责：一个组件只做一件事
- 受控 vs 非受控：表单用 React Hook Form（非受控），简单交互用 useState
- Props 设计：回调用 `onXxx`，布尔用 `isXxx`/`hasXxx`
- 组合优于继承：使用 `children` 和 `render props`
- 组件粒度：页面 > Section > Component > Primitive

### 16.3 Git 规范

```
feat: 添加测试用例列表页面
fix: 修复执行页面键盘快捷键失效
refactor: 抽离分页逻辑到 usePagination hook
style: 统一 Badge 组件配色
chore: 升级 TanStack Query 到 v5
```

---

## 18. 环境变量

```bash
# .env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=Heka AI 测试管理平台

# .env.production
VITE_API_BASE_URL=/api/v1
VITE_APP_TITLE=Heka AI 测试管理平台
```

---

## 19. 构建部署

### 18.1 开发环境

```bash
npm run dev          # 启动 Vite 开发服务器 (localhost:5173)
# API 代理配置在 vite.config.ts 中
```

### 18.2 生产构建

```bash
npm run build        # 输出到 dist/
npm run preview      # 预览生产构建
```

### 18.3 Docker 构建

```dockerfile
# frontend/Dockerfile
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

### 18.4 Vite 代理配置

```ts
// vite.config.ts
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

---

## 20. 页面与 API 端点映射

### 19.1 认证

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 登录 | `/auth/login` | POST | 登录获取 Token |
| 应用初始化 | `/auth/me` | GET | 验证 Token，获取用户信息 |
| Token 刷新 | `/auth/refresh` | POST | 刷新即将过期的 Token |

### 19.2 项目管理

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 项目列表 | `/projects` | GET | 获取用户参与的项目列表 |
| 创建项目 | `/projects` | POST | 新建项目 |
| 项目概览 | `/projects/:id` | GET | 项目详情 + 统计数据 |
| 添加成员 | `/projects/:id/members` | POST | 邀请用户加入项目 |

### 19.3 模块管理

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 模块树 | `/modules?project_id=:id` | GET | 获取模块树结构 |
| 创建模块 | `/modules` | POST | 新建模块 |
| 编辑模块 | `/modules/:id` | PUT | 更新模块信息 |
| 删除模块 | `/modules/:id` | DELETE | 删除模块（用例移至未分类） |

### 19.4 测试用例

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 用例列表 | `/testcases?project_id=:id` | GET | 分页、筛选、排序 |
| 创建用例 | `/testcases` | POST | 新建用例 |
| 用例详情 | `/testcases/:id` | GET | 含步骤、版本号 |
| 编辑用例 | `/testcases/:id` | PUT | 携带 version 乐观锁 |
| 删除用例 | `/testcases/:id` | DELETE | 单个删除 |
| 批量改状态 | `/testcases/batch/status` | PUT | `{ids, status}` |
| 批量删除 | `/testcases/batch` | DELETE | `{ids}` |
| 批量移动 | `/testcases/batch/move` | PUT | `{ids, folder_id}` |

### 19.5 用例集合

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 集合列表 | `/collections?project_id=:id` | GET | 获取项目下所有集合 |
| 创建集合 | `/collections` | POST | 新建集合 |
| 集合详情 | `/collections/:id` | GET | 集合基本信息 |
| 集合用例列表 | `/collections/:id/cases` | GET | 分页获取集合内用例 |
| 添加用例到集合 | `/collections/:id/cases` | POST | `{test_case_ids}` |
| 从集合移除用例 | `/collections/:id/cases` | DELETE | `{test_case_ids}` |

### 19.6 测试计划

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 计划列表 | `/testplans?project_id=:id` | GET | 分页、状态筛选 |
| 创建计划 | `/testplans` | POST | 含用例分配 |
| 计划详情 | `/testplans/:id` | GET | 含用例列表、进度、当前执行 |
| 开始执行 | `/testplans/:id/start` | POST | `{name, executor_notes}` |
| 暂停执行 | `/testplans/:id/pause` | POST | 暂停当前执行 |
| 恢复执行 | `/testplans/:id/resume` | POST | 恢复暂停的执行 |
| 完成计划 | `/testplans/:id/complete` | POST | `{summary}` |
| 取消计划 | `/testplans/:id/cancel` | POST | `{reason}` |

### 19.7 执行记录

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 执行详情 | `/executions/:id` | GET | 含结果列表、摘要 |
| 提交单条结果 | `/executions/:id/results` | POST | `{test_case_id, status, ...}` |
| 批量提交结果 | `/executions/:id/results/batch` | POST | `{results: [...]}` |

### 19.8 文件管理

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 文件列表 | `/files?project_id=:id` | GET | 分页、类型筛选 |
| 上传文件 | `/files/upload` | POST | multipart/form-data, ≤100MB |
| Figma 链接 | `/files/figma` | POST | `{project_id, figma_url, name}` |
| 文件详情 | `/files/:id` | GET | 含索引状态、版本信息 |
| 索引状态 | `/files/:id/index-status` | GET | 轮询索引进度 |
| 重新索引 | `/files/:id/reindex` | POST | `{force: boolean}` |
| 删除文件 | `/files/:id` | DELETE | 输入文件名确认 |

### 19.9 AI 服务

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| AI 生成用例 | `/ai/generate-testcases` | POST | 异步，返回 task_id |
| 任务状态轮询 | `/ai/tasks/:id` | GET | 2s 间隔轮询 |
| AI 分析 | `/ai/analyze` | POST | 代码变更影响分析 |
| SSE 事件流 | `/ai/tasks/:id/events` | GET | 预留，当前使用轮询 |

### 19.10 报表

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 计划报告 | `/reports/plan/:plan_id` | GET | 含失败用例、执行历史 |
| 覆盖度报告 | `/reports/coverage?project_id=:id` | GET | 按状态/优先级/模块 |
| 趋势报告 | `/reports/trend?project_id=:id&days=30` | GET | 每日趋势 |
| 缺陷分布 | `/reports/bugs?project_id=:id` | GET | 含 top 失败用例 |
| 工作量 | `/reports/workload?user_id=:id` | GET | 按人统计 |

### 19.11 标签管理

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 标签列表 | `/tags?project_id=:id` | GET | 含使用计数 |
| 创建标签 | `/tags` | POST | `{project_id, name, color}` |

### 19.12 系统监控

| 页面 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 健康检查 | `/health` | GET | 服务状态 |
| AI 监控 | `/monitoring/ai` | GET | AI Provider 状态 |

---

**文档版本**：v1.0
**最后更新**：2025-05-18
