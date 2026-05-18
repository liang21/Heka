# Heka 前端原子化任务列表

> 基于 `spec.md` v1.1 + `plan.md` v1.1，按组件分层拆解
> 版本：v1.1
> 日期：2026-05-18

---

## 约定

- **[P]** = 可与同组其他 [P] 任务并行执行
- **Depends** = 前置任务 ID
- 每个任务只涉及**一个主要文件**的创建或修改（Verify 类型 Action 可省略 File 字段）
- 前端不采用 TDD，但每个任务包含自测验收要点

---

## Phase 1: Foundation（项目骨架 + 基础设施）

> Vite 项目启动、核心类型定义、API 层、Store、路由、布局、登录、项目列表

---

### 1.1 项目初始化

#### T001: Vite + React + TypeScript 脚手架
- **File:** `package.json`, `vite.config.ts`, `tsconfig.json`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `npm create vite@latest heka-frontend -- --template react-ts`
  - 配置 `vite.config.ts`：路径别名 `@/` → `src/`
  - 配置 `tsconfig.json`：`paths: { "@/*": ["src/*"] }`
  - 确保 `npm run dev` 启动无报错

#### T002: Tailwind CSS 集成
- **File:** `tailwind.config.ts`, `postcss.config.js`, `src/styles/globals.css`
- **Depends:** T001
- **Action:** Create
- **Details:**
  - 安装 `tailwindcss`, `postcss`, `autoprefixer`
  - `tailwind.config.ts` 配置 content 路径 + Shadcn/ui 的 `tailwindCSS` 配置
  - `globals.css` 引入 `@tailwind base/components/utilities` + CSS 变量（参照 spec.md 12.1）

#### T003: Shadcn/ui 初始化 + 基础组件
- **File:** `components.json`, `src/components/ui/*`
- **Depends:** T002
- **Action:** Create
- **Details:**
  - `npx shadcn-ui@latest init`
  - 引入基础组件：Button, Card, Input, Dialog, Form, Label, Skeleton, DropdownMenu, Avatar, Badge, Separator, ScrollArea, Tooltip
  - 确保 `import { Button } from "@/components/ui/button"` 可用

#### T004: [P] 目录结构搭建
- **File:** `src/` 下全部子目录
- **Depends:** T001
- **Action:** Create
- **Details:**
  - 创建 `app/`, `pages/`, `components/`, `hooks/`, `services/`, `stores/`, `types/`, `lib/`, `styles/` 目录
  - 每个目录放 `.gitkeep` 占位
  - 参照 spec.md 16 节完整目录结构

#### T005: [P] 环境变量文件
- **File:** `.env`, `.env.production`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `.env`: `VITE_API_BASE_URL=http://localhost:8080/api/v1`, `VITE_APP_TITLE=Heka AI 测试管理平台`
  - `.env.production`: `VITE_API_BASE_URL=/api/v1`

#### T006: [P] Lucide React 图标引入
- **File:** `package.json`
- **Depends:** T001
- **Action:** Modify
- **Details:**
  - 安装 `lucide-react`
  - 验证 `import { Home } from "lucide-react"` 可用

#### T007: Vite 代理配置
- **File:** `vite.config.ts`
- **Depends:** T001
- **Action:** Modify
- **Details:**
  - 添加 `server.proxy: { "/api": { target: "http://localhost:8080", changeOrigin: true } }`
  - 开发阶段即可联调后端 API

---

### 1.2 类型定义

#### T008: [P] 通用类型
- **File:** `src/types/api.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `ApiResponse<T>`, `PaginatedData<T>`, `PaginationParams`, `PaginatedResponse<T>`
  - 参照 spec.md 3.1

#### T009: [P] 认证类型
- **File:** `src/types/auth.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `User`, `LoginRequest`, `LoginResponse`
  - 参照 spec.md 3.2

#### T010: [P] 项目类型
- **File:** `src/types/project.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `Project`, `ProjectDetail`, `ProjectMember`, `ProjectStatistics`
  - 参照 spec.md 3.3

---

### 1.3 工具与常量

#### T011: [P] 工具函数（cn 等）
- **File:** `src/lib/utils.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `import { clsx, type ClassValue } from "clsx"; import { twMerge } from "tailwind-merge"`
  - `export function cn(...inputs: ClassValue[]) { return twMerge(clsx(inputs)) }`

#### T012: [P] 常量定义
- **File:** `src/lib/constants.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `MAX_UPLOAD_SIZE = 100 * 1024 * 1024`, `CHUNK_SIZE = 5 * 1024 * 1024`, `DEFAULT_PAGE_SIZE = 20`, `MAX_PAGE_SIZE = 100`

#### T012a: [P] 格式化工具
- **File:** `src/lib/format.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `formatDate(date, pattern)` — date-fns `format` 封装
  - `formatRelativeTime(date)` — 相对时间（"3 分钟前"、"昨天"）
  - `formatFileSize(bytes)` — 文件大小（"12.5 MB"）
  - 安装 `date-fns`

#### T013: [P] 错误码映射
- **File:** `src/lib/error-messages.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `export const errorMessages: Record<string, string>` 全量错误码 → 中文
  - 参照 spec.md 10.2 全部 ~25 个错误码

#### T014: [P] Zod Schema（loginSchema + createProjectSchema）
- **File:** `src/lib/schemas.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `loginSchema`: email(必填, email格式) + password(必填)
  - `createProjectSchema`: name(必填, max 255) + description(optional, max 5000)
  - 参照 spec.md 11 节

---

### 1.4 API 基础层

#### T015: Axios 实例 + 请求拦截器
- **File:** `src/services/api.ts`
- **Depends:** T008, T011
- **Action:** Create
- **Details:**
  - `axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL, timeout: 30000 })`
  - 请求拦截器：从 `useAuthStore.getState().token` 读取 → 注入 `Authorization: Bearer {token}`

#### T016: 响应拦截器（401 静默刷新 + 排队重试）
- **File:** `src/services/api.ts`（追加）
- **Depends:** T015
- **Action:** Modify
- **Details:**
  - 成功响应：返回 `response.data`
  - 401 首次：POST `/auth/refresh` 获取新 Token → 重试原请求
  - 并发请求排队：维护 `refreshSubscribers` 队列，刷新完成后逐个重试
  - 刷新失败：`clearAuth()` + 跳转 `/login`

#### T017: [P] Query Key 工厂（projects 域）
- **File:** `src/lib/query-keys.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `export const queryKeys = { projects: { all, detail(id), members(id) } }`
  - 后续阶段追加 modules, testCases, plans 等域

#### T018: [P] auth 服务
- **File:** `src/services/auth.ts`
- **Depends:** T015, T009
- **Action:** Create
- **Details:**
  - `login(data: LoginRequest)`, `getMe()`, `refresh()`
  - 参照 spec.md 5.2 auth.ts

#### T019: [P] project 服务
- **File:** `src/services/project.ts`
- **Depends:** T015, T010
- **Action:** Create
- **Details:**
  - `list()`, `create(data)`, `getById(id)`, `addMember(id, userId)`

---

### 1.5 Store

#### T020: [P] auth store
- **File:** `src/stores/auth.ts`
- **Depends:** T009
- **Action:** Create
- **Details:**
  - Zustand + `persist` 中间件 + `partialize` 仅保存 `token`
  - `interface AuthState { token, user, setAuth(token, user), clearAuth(), isAuthenticated() }`
  - user 信息由 `GET /auth/me` 刷新，不持久化

#### T021: [P] project store
- **File:** `src/stores/project.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - Zustand + `persist`
  - `interface ProjectState { currentProjectId, setCurrentProject(id), sidebarCollapsed, toggleSidebar() }`

---

### 1.6 路由与布局

#### T022: Providers 聚合
- **File:** `src/app/providers.tsx`
- **Depends:** T017
- **Action:** Create
- **Details:**
  - `QueryClientProvider` + QueryClient 配置（`staleTime: 2min`, `gcTime: 30min`, `retry: 1`）
  - 导出 `Providers` 组件

#### T023: 懒加载包装组件
- **File:** `src/components/shared/LazyLoad.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - `Suspense` 包裹 + Skeleton fallback
  - 导出 `withSuspense` HOC

#### T024: [P] ProjectSidebar 组件
- **File:** `src/components/layout/ProjectSidebar.tsx`
- **Depends:** T003, T021
- **Action:** Create
- **Details:**
  - 项目导航菜单：概览、模块、用例、计划、执行、报表、文件、AI
  - 折叠/展开交互（读取 `sidebarCollapsed`）
  - 使用 `useNavigate` + `useParams` 高亮当前路由

#### T025: [P] Breadcrumb 面包屑组件
- **File:** `src/components/layout/Breadcrumb.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 多级面包屑导航，支持动态参数（项目名、用例标题等）
  - 使用 Shadcn/ui Breadcrumb 组件

#### T026: AppLayout 布局
- **File:** `src/components/layout/AppLayout.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - TopBar（Logo + 用户头像 + 下拉菜单）+ 主内容区 `<Outlet />`
  - 全局宽度 100%，无侧边栏

#### T027: ProjectLayout 布局
- **File:** `src/components/layout/ProjectLayout.tsx`
- **Depends:** T024, T025
- **Action:** Create
- **Details:**
  - 引用 ProjectSidebar + Breadcrumb + `<Outlet />`
  - 左侧 ProjectSidebar(260px) + 右侧内容区

#### T028: useAuthInit Hook + AuthGuard
- **File:** `src/hooks/useAuthInit.ts`
- **Depends:** T020, T018
- **Action:** Create
- **Details:**
  - `useAuthInit`：应用启动时，若 localStorage 有 token → 调用 `GET /auth/me` 校验 + 恢复 user
  - `AuthGuard` 组件：未登录跳转 `/login`，已登录访问 `/login` 跳转 `/`

#### T029: React Router 配置
- **File:** `src/app/router.tsx`
- **Depends:** T023, T026, T027, T028
- **Action:** Create
- **Details:**
  - 阶段 1 路由：`/login` + `/`(AppLayout, children: `index→Projects`, `settings`, `project/:projectId→ProjectLayout`)
  - 所有页面组件用 `React.lazy` + `Suspense`

#### T030: App 入口
- **File:** `src/app/App.tsx`
- **Depends:** T022, T029
- **Action:** Create
- **Details:**
  - 挂载 `Providers` → `RouterProvider`
  - 引入全局样式 `globals.css`

---

### 1.7 页面组件

#### T031: 登录页
- **File:** `src/pages/Login/index.tsx`
- **Depends:** T003, T014, T018, T020
- **Action:** Create
- **Details:**
  - 居中卡片 `max-w-[400px]`，白色卡片 + 浅灰背景
  - React Hook Form + `loginSchema` 校验
  - `useMutation` 调用 `authApi.login` → `setAuth(token, user)` → `navigate("/")`
  - 错误码 `AUTH-VL-001` → 表单下方红色文字

#### T032: 项目列表页（首页）
- **File:** `src/pages/Projects/index.tsx`
- **Depends:** T003, T014, T019
- **Action:** Create
- **Details:**
  - `useQuery` 调用 `projectApi.list`
  - 卡片网格（响应式 1/2/3/4 列）
  - 新建项目 Dialog（React Hook Form + `createProjectSchema`）
  - 点击卡片 → `navigate(/project/${id})`
  - 空状态："还没有项目，创建一个开始吧"

#### T033: useProject Hook
- **File:** `src/hooks/useProject.ts`
- **Depends:** T019, T017
- **Action:** Create
- **Details:**
  - `useProjectDetail(projectId)` — `useQuery` 调用 `projectApi.getById`

#### T034: 项目概览页
- **File:** `src/pages/Overview/index.tsx`
- **Depends:** T033, T003
- **Action:** Create
- **Details:**
  - 统计卡片行（用例数、计划数、执行数、成员数）
  - 成员列表（头像 + 名称）
  - 快捷操作按钮（新建用例、新建计划、上传文件、AI 生成）
  - 参照 spec.md 7.3

#### T035: [P] 项目设置页
- **File:** `src/pages/Settings/ProjectSettings.tsx`
- **Depends:** T033, T003
- **Action:** Create
- **Details:**
  - 基本信息编辑表单（项目名称、描述）
  - 调用 `projectApi` 更新接口（需后端支持）

#### T036: [P] 系统设置页
- **File:** `src/pages/Settings/index.tsx`
- **Depends:** T020, T003
- **Action:** Create
- **Details:**
  - 用户信息展示（从 auth store 读取）
  - 主题切换入口（MVP 占位，暂不实现）

#### T037: [P] EmptyState 通用组件
- **File:** `src/components/shared/EmptyState.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - Props: `icon`, `title`, `description`, `actionLabel?`, `onAction?`
  - 居中布局 + 图标 + 文案 + 可选操作按钮

#### T038: [P] ConfirmDialog 通用组件
- **File:** `src/components/shared/ConfirmDialog.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - Props: `title`, `description`, `confirmText?`, `variant?`（default/destructive）
  - 基于 Shadcn/ui AlertDialog

#### T039: [P] useConfirm Hook
- **File:** `src/hooks/useConfirm.ts`
- **Depends:** T038
- **Action:** Create
- **Details:**
  - 命令式调用 `confirm({ title, description })` → 返回 `Promise<boolean>`
  - 内部管理 ConfirmDialog 的 open/close 状态

### 阶段 1 验收

- [ ] `npm run dev` 启动，页面正常渲染
- [ ] 登录 → 存储 Token → 跳转首页
- [ ] Token 过期自动刷新，刷新失败跳转登录
- [ ] 项目列表展示，可创建新项目
- [ ] 点击项目 → 概览页展示统计 + 成员
- [ ] 未登录自动跳转 `/login`

---

## Phase 2: Core Test Cases（核心用例管理）

> 模块树、标签、用例 CRUD、步骤编辑器、批量操作、集合

---

### 2.1 类型与服务准备

#### T040: [P] 模块类型
- **File:** `src/types/module.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `Module` 接口，参照 spec.md 3.4

#### T041: [P] 测试用例类型
- **File:** `src/types/testcase.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `TestCase`, `TestStep`, `TestCaseStatus`, `Priority`, `TestCaseListParams`，参照 spec.md 3.5

#### T042: [P] 标签类型
- **File:** `src/types/tag.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `Tag` 接口，参照 spec.md 3.10

#### T043: [P] 集合类型
- **File:** `src/types/collection.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `Collection` 接口，参照 spec.md 3.10

#### T044: [P] 筛选参数类型
- **File:** `src/types/filters.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `TestCaseFilters`, `CollectionCaseFilters`，参照 spec.md 3.11

#### T045: Query Key 补充（阶段 2 域）
- **File:** `src/lib/query-keys.ts`
- **Depends:** T017, T040, T041, T042, T043, T044
- **Action:** Modify
- **Details:**
  - 追加 modules, testCases, tags, collections 域
  - 参照 spec.md 4.2 queryKeys 完整定义

#### T046: [P] module 服务
- **File:** `src/services/module.ts`
- **Depends:** T015, T040
- **Action:** Create
- **Details:** `tree(projectId)`, `create(data)`, `update(id, data)`, `delete(id)`

#### T047: [P] testcase 服务
- **File:** `src/services/testcase.ts`
- **Depends:** T015, T041
- **Action:** Create
- **Details:** 8 个方法：`list`, `getById`, `create`, `update`, `delete`, `batchUpdateStatus`, `batchDelete`, `batchMove`

#### T048: [P] tag 服务
- **File:** `src/services/tag.ts`
- **Depends:** T015, T042
- **Action:** Create
- **Details:** `list(projectId)`, `create(data)`

#### T049: [P] collection 服务
- **File:** `src/services/collection.ts`
- **Depends:** T015, T043
- **Action:** Create
- **Details:** 6 个方法：`list`, `create`, `getById`, `getCases`, `addCases`, `removeCases`

---

### 2.2 模块管理

#### T050: ModuleTree 递归树组件
- **File:** `src/components/module/ModuleTree.tsx`
- **Depends:** T003, T040
- **Action:** Create
- **Details:**
  - 递归渲染模块树，支持展开/折叠
  - 节点渲染：名称 + 用例数 badge + 右键菜单（编辑/删除/新建子模块）
  - 选中高亮

#### T051: @dnd-kit 拖拽排序集成
- **File:** `src/components/module/ModuleTree.tsx`（追加）
- **Depends:** T050
- **Action:** Modify
- **Details:**
  - 安装 `@dnd-kit/core`, `@dnd-kit/sortable`, `@dnd-kit/utilities`
  - 集成 `@dnd-kit/core` + `@dnd-kit/sortable`
  - 支持同级排序 + 层级移动（拖入父节点/子节点）
  - 拖拽结束调用 `moduleApi.update(id, { parent_id, order_index })`

#### T052: ModuleForm 表单组件
- **File:** `src/components/module/ModuleForm.tsx`
- **Depends:** T003, T040
- **Action:** Create
- **Details:**
  - React Hook Form：名称(必填) + 描述 + 父模块选择下拉（从 ModuleTree 数据获取）

#### T053: useModules Hook
- **File:** `src/hooks/useModules.ts`
- **Depends:** T046, T045
- **Action:** Create
- **Details:**
  - `useModuleTree(projectId)` — useQuery
  - `useCreateModule` — useMutation + invalidate module tree
  - `useUpdateModule` — useMutation + invalidate
  - `useDeleteModule` — useMutation + invalidate

#### T054: 模块管理页面
- **File:** `src/pages/Modules/index.tsx`
- **Depends:** T050, T052, T053, T039
- **Action:** Create
- **Details:**
  - ModuleTree + 新建/编辑(Dialog with ModuleForm) / 删除(ConfirmDialog)
  - 删除确认文案："该模块下有 N 个用例，用例将移至'未分类'"

#### T055: [P] 创建模块页面
- **File:** `src/pages/Modules/CreateModule.tsx`
- **Depends:** T052, T053
- **Action:** Create
- **Details:** ModuleForm 独立页面 + 提交后跳转模块列表

---

### 2.3 测试用例列表

#### T056: [P] PriorityBadge 组件
- **File:** `src/components/testcase/PriorityBadge.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 4 级优先级 Badge：P0 灰色(低), P1 蓝色(中), P2 橙色(高), P3 红色(紧急)。参照 spec.md 12.3

#### T057: [P] StatusTag 组件
- **File:** `src/components/testcase/StatusTag.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 3 种状态标签：draft(灰), ready(绿), archived(黄)

#### T058: TestCaseTable 组件
- **File:** `src/components/testcase/TestCaseTable.tsx`
- **Depends:** T003, T056, T057
- **Action:** Create
- **Details:**
  - Shadcn/ui Table：checkbox 行选择 + 标题/优先级/状态/更新时间列 + 排序列头
  - 点击行 → `navigate(testcases/:id)`
  - 支持多选 + BatchActions 联动

#### T059: TestCaseFilters 筛选组件
- **File:** `src/components/testcase/TestCaseFilters.tsx`
- **Depends:** T003, T044
- **Action:** Create
- **Details:**
  - status Select + priority Select + keyword SearchInput + tags 多选下拉
  - 所有筛选值通过 props 回调给父组件

#### T060: [P] BatchActions 批量操作栏
- **File:** `src/components/testcase/BatchActions.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 底部固定操作栏：选中 N 项 → 改状态/移动到模块/删除
  - 批量删除使用 ConfirmDialog

#### T061: [P] Pagination 分页组件
- **File:** `src/components/shared/Pagination.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 页码按钮 + 上一页/下一页 + 每页条数 Select
  - Props: `page`, `pageSize`, `total`, `onChange`

#### T062: [P] SearchInput 组件
- **File:** `src/components/shared/SearchInput.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 搜索输入框 + 放大镜图标 + 300ms 防抖
  - Props: `value`, `onChange`, `placeholder`

#### T063: usePagination Hook
- **File:** `src/hooks/usePagination.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - URL 参数同步：`useSearchParams` 读写 `page` / `page_size`
  - 返回 `{ page, pageSize, setPage, setPageSize }`

#### T064: useTestCases Hook
- **File:** `src/hooks/useTestCases.ts`
- **Depends:** T047, T045
- **Action:** Create
- **Details:**
  - `useTestCaseList(projectId, filters)` — useQuery + `placeholderData: (prev) => prev`
  - `useCreateTestCase(projectId)` — useMutation + invalidate `["projects", projectId, "testcases"]`
  - `useBatchUpdateStatus`, `useBatchDelete`, `useBatchMove` — useMutation + invalidate

#### T065: useTestCaseFilters Hook
- **File:** `src/hooks/useTestCaseFilters.ts`
- **Depends:** T044
- **Action:** Create
- **Details:**
  - 筛选状态 ↔ URL query params 双向同步
  - 返回 `{ filters, setFilter, resetFilters }` + URL 自动更新

#### T066: 测试用例列表页
- **File:** `src/pages/TestCases/index.tsx`
- **Depends:** T050, T058, T059, T060, T061, T062, T063, T064, T065
- **Action:** Create
- **Details:**
  - 三栏布局：左侧模块树(260px, 可折叠) + 右侧筛选栏 + 表格 + 分页
  - 模块树点击 → 更新 `module_id` 筛选
  - 参照 spec.md 7.4

---

### 2.4 创建/编辑测试用例

#### T067: Zod Schema 追加（createTestCaseSchema）
- **File:** `src/lib/schemas.ts`（追加）
- **Depends:** T014, T041
- **Action:** Modify
- **Details:**
  - `createTestCaseSchema`: title(必填, max500) + description(optional, max10000) + module_id(optional) + steps(array, min1) + priority(0-3) + tags(optional)

#### T068: StepEditor 步骤编辑器
- **File:** `src/components/testcase/StepEditor.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - `useFieldArray` 动态数组：每行 = 序号 + 操作输入 + 预期结果输入 + 删除按钮
  - 支持增删行 + `@dnd-kit` 拖拽排序
  - 自动序号

#### T069: [P] TagInput 标签输入组件
- **File:** `src/components/testcase/TagInput.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 已有标签下拉选择 + 自定义输入创建
  - 选中标签显示为 Badge，可点击移除

#### T070: useUnsavedChanges Hook
- **File:** `src/hooks/useUnsavedChanges.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - React Router `useBlocker` + `window.beforeunload`
  - 参数 `isDirty: boolean`，为 true 时拦截离开

#### T071: 创建用例页面
- **File:** `src/pages/TestCases/CreateTestCase.tsx`
- **Depends:** T067, T068, T069, T064
- **Action:** Create
- **Details:**
  - React Hook Form + `createTestCaseSchema`
  - StepEditor + TagInput + 模块选择 + 优先级选择
  - 提交 → `useCreateTestCase` → 成功后跳转详情页

#### T072: 编辑用例页面（含 version 乐观锁）
- **File:** `src/pages/TestCases/EditTestCase.tsx`
- **Depends:** T071, T064, T070
- **Action:** Create
- **Details:**
  - 加载用例详情 → 记录 `version`
  - 提交时携带 `version`：`testcaseApi.update(id, { ...data, version })`
  - 收到 409 `TC-CF-001` → 冲突弹窗："该用例已被其他人修改"
  - `useUnsavedChanges(form.formState.isDirty)`

---

### 2.5 用例详情

#### T073: [P] StatusActions 状态流转按钮
- **File:** `src/components/testcase/StatusActions.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 根据当前 status 显示可用操作按钮
  - draft → [标记就绪] [归档]; ready → [改回草稿] [归档]; archived → [恢复]
  - 参照 spec.md 8.1

#### T074: 用例详情页
- **File:** `src/pages/TestCases/TestCaseDetail.tsx`
- **Depends:** T033, T073, T068, T025
- **Action:** Create
- **Details:**
  - Breadcrumb + 标题区 + 描述 + 步骤表格 + 执行历史列表
  - StatusActions 按钮
  - 点击"编辑"→ 标题/描述/步骤切换为可编辑模式
  - 参照 spec.md 7.6

---

### 2.6 标签管理

#### T075: useTags Hook
- **File:** `src/hooks/useTags.ts`
- **Depends:** T048, T045
- **Action:** Create
- **Details:**
  - `useTagList(projectId)` — useQuery
  - `useCreateTag` — useMutation + invalidate tags list

---

### 2.7 用例集合

#### T076: [P] CollectionCard 组件
- **File:** `src/components/collection/CollectionCard.tsx`
- **Depends:** T003, T043
- **Action:** Create
- **Details:** 集合卡片：名称 + 描述 + 用例数 + 创建时间

#### T077: [P] AddCasesDialog 组件
- **File:** `src/components/collection/AddCasesDialog.tsx`
- **Depends:** T003, T058
- **Action:** Create
- **Details:**
  - Dialog 弹出用例列表（复用 TestCaseTable，checkbox 多选）
  - 确认后调用 `collectionApi.addCases`

#### T078: [P] CollectionCasesList 组件
- **File:** `src/components/collection/CollectionCasesList.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 集合内用例列表 + 每行移除按钮

#### T079: useCollections Hook
- **File:** `src/hooks/useCollections.ts`
- **Depends:** T049, T045
- **Action:** Create
- **Details:**
  - `useCollectionList`, `useCreateCollection`, `useCollectionDetail`, `useAddCases`, `useRemoveCases`

#### T080: 集合列表页
- **File:** `src/pages/TestCases/Collections.tsx`
- **Depends:** T076, T077, T078, T079
- **Action:** Create
- **Details:**
  - 集合卡片网格 + 新建 Dialog
  - 点击集合 → 展示 CollectionCasesList + AddCasesDialog

#### T081: 阶段 2 路由注册
- **File:** `src/app/router.tsx`（追加）
- **Depends:** T029, T054, T066, T071, T072, T074, T080
- **Action:** Modify
- **Details:**
  - 在 `project/:projectId` children 中追加：modules, modules/create, testcases, testcases/create, testcases/collections, testcases/:id 路由
  - **注意**：`testcases/collections` 必须在 `testcases/:id` 之前

### 阶段 2 验收

- [ ] 模块树 CRUD + 拖拽排序
- [ ] 用例列表分页筛选（URL 参数同步）
- [ ] 创建/编辑用例（步骤编辑器 + 标签 + 乐观锁）
- [ ] 用例详情 + 状态流转 + 执行历史
- [ ] 批量操作（改状态/移动/删除）
- [ ] 集合 CRUD + 用例关联
- [ ] 离开编辑页未保存提示

---

## Phase 3: Plan & Execution（计划与执行）

---

### 3.1 类型与服务准备

#### T082: [P] 计划类型
- **File:** `src/types/plan.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `TestPlan`, `PlanDetail`, `PlanTestCase`, `PlanProgress`, `ExecutionSummary`, `PlanStatus`，参照 spec.md 3.6

#### T083: [P] 执行类型
- **File:** `src/types/execution.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `Execution`, `ExecutionDetail`, `ExecutionResult`, `ExecutionStatus`, `ResultStatus`, `SubmitResultRequest`，参照 spec.md 3.6

#### T084: [P] 筛选参数补充
- **File:** `src/types/filters.ts`（追加）
- **Depends:** T044
- **Action:** Modify
- **Details:** 追加 `PlanFilters`, `ExecutionFilters`

#### T085: Query Key 补充（plans + executions 域）
- **File:** `src/lib/query-keys.ts`（追加）
- **Depends:** T045, T082, T083
- **Action:** Modify
- **Details:** 追加 plans, executions 域

#### T086: Zod Schema 追加（createPlanSchema）
- **File:** `src/lib/schemas.ts`（追加）
- **Depends:** T014, T082
- **Action:** Modify
- **Details:** `createPlanSchema`: name(必填, max255) + description(optional) + test_case_ids(array, min1) + assignments(optional)

#### T087: [P] plan 服务
- **File:** `src/services/plan.ts`
- **Depends:** T015, T082
- **Action:** Create
- **Details:** 8 个方法：`list`, `create`, `getById`, `start`, `pause`, `resume`, `complete`, `cancel`

#### T088: [P] execution 服务
- **File:** `src/services/execution.ts`
- **Depends:** T015, T083
- **Action:** Create
- **Details:** `getById`, `submitResult`, `submitBatchResults`

---

### 3.2 测试计划

#### T089: usePlans Hook
- **File:** `src/hooks/usePlans.ts`
- **Depends:** T087, T085
- **Action:** Create
- **Details:**
  - `usePlanList`, `usePlanDetail`, `useCreatePlan`
  - `usePlanAction` — useMutation 封装 start/pause/resume/complete/cancel + invalidate

#### T090: 测试计划列表页
- **File:** `src/pages/TestPlans/index.tsx`
- **Depends:** T089, T003
- **Action:** Create
- **Details:**
  - 计划表格（名称/状态/用例数/创建时间）+ 状态筛选 + 新建按钮
  - 状态列使用 StatusTag 组件

#### T091: [P] AssignSelector 执行人选择组件
- **File:** `src/components/plan/AssignSelector.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 项目成员下拉列表 + 用例-执行人映射表格

#### T092: 创建计划页面
- **File:** `src/pages/TestPlans/CreatePlan.tsx`
- **Depends:** T086, T089, T091
- **Action:** Create
- **Details:**
  - React Hook Form + `createPlanSchema`
  - 用例多选（复用 TestCaseTable，checkbox）+ AssignSelector 执行人分配
  - 参照 spec.md 7.7

#### T092a: PlanStatusActions 计划状态流转按钮
- **File:** `src/components/plan/PlanStatusActions.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 根据当前 PlanStatus 显示可用操作按钮
  - draft → [开始执行] [编辑] [取消计划]; active → [暂停] [查看执行]; paused → [恢复执行] [取消计划]; completed → [查看报告]; cancelled → 仅查看
  - 参照 spec.md 8.2

#### T093: 计划详情页
- **File:** `src/pages/TestPlans/PlanDetail.tsx`
- **Depends:** T089, T092a
- **Action:** Create
- **Details:**
  - PlanStatusActions 状态按钮
  - 用例列表表格（PlanTestCase[]）+ 添加/移除用例 + 分配执行人
  - 执行历史列表
  - 并发执行预检：点击"开始执行"前检查 `current_execution`，有 in_progress 则弹窗选择"继续上次"或"取消并新建"

---

### 3.3 执行管理

#### T094: ExecutionPanel 执行面板
- **File:** `src/components/plan/ExecutionPanel.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 展示当前用例步骤详情
  - 4 个结果按钮：通过(绿)/失败(红)/阻塞(黄)/跳过(灰)
  - 选择"失败"时展开 Bug ID 输入框 + Notes 输入

#### T095: [P] ResultSummary 结果汇总
- **File:** `src/components/plan/ResultSummary.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** passed/failed/blocked/skipped 四格统计 + 进度条

#### T096: useExecution Hook
- **File:** `src/hooks/useExecution.ts`
- **Depends:** T088, T085
- **Action:** Create
- **Details:**
  - `useExecutionDetail` — useQuery
  - `useSubmitResult` — useMutation + invalidate execution detail
  - `useSubmitBatchResults` — useMutation + invalidate

#### T097: useKeyboard Hook（执行页快捷键）
- **File:** `src/hooks/useKeyboard.ts`
- **Depends:** —
- **Action:** Create
- **Details:**
  - `useEffect` 监听 keydown：P=通过, F=失败, B=阻塞, S=跳过, Enter=确认并下一个
  - 仅在执行页面激活（通过 enabled 参数控制）
  - 全局快捷键（Cmd+K 等）在阶段 5 扩展

#### T098: 执行页面（左右分栏）
- **File:** `src/pages/TestPlans/ExecutePlan.tsx`
- **Depends:** T094, T095, T096, T097
- **Action:** Create
- **Details:**
  - 左 40%：用例列表（显示执行结果图标，当前高亮）
  - 右 60%：ExecutionPanel + Bug ID 扩展
  - 底部：进度条 + ResultSummary
  - 提交结果后自动跳转下一个未执行用例
  - 参照 spec.md 7.8

---

### 3.4 执行记录

#### T099: 执行记录列表页
- **File:** `src/pages/Executions/index.tsx`
- **Depends:** T096, T003
- **Action:** Create
- **Details:**
  - 执行记录表格 + 状态筛选（in_progress/completed/cancelled）
  - 点击 → 跳转执行详情

#### T100: 执行记录详情页
- **File:** `src/pages/Executions/ExecutionDetail.tsx`
- **Depends:** T096, T095
- **Action:** Create
- **Details:**
  - 执行信息（计划名/执行人/时间/状态）
  - ResultSummary + 结果列表（按用例分组）

#### T101: 阶段 3 路由注册
- **File:** `src/app/router.tsx`（追加）
- **Depends:** T081, T090, T092, T093, T098, T099, T100
- **Action:** Modify
- **Details:** 追加 plans, plans/create, plans/:id, plans/:id/execute, executions, executions/:id 路由

### 阶段 3 验收

- [ ] 创建计划 + 关联用例 + 分配执行人
- [ ] 计划状态流转按钮动态显示
- [ ] 并发执行预检 + 冲突弹窗
- [ ] 执行页面左右分栏 + 键盘快捷键
- [ ] 提交结果自动跳转下一个
- [ ] 执行记录列表 + 详情

---

## Phase 4: Files & AI（文件与 AI）

---

### 4.1 类型与服务准备

#### T102: [P] 文件类型
- **File:** `src/types/file.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `ProjectFile`, `FileSourceType`, `FileIndexStatus`，参照 spec.md 3.7

#### T103: [P] AI 类型
- **File:** `src/types/ai.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `AITask`, `AITaskStatus`, `AIGeneratedCase`, `AIAnalysisResult`，参照 spec.md 3.8

#### T104: [P] 筛选参数补充
- **File:** `src/types/filters.ts`（追加）
- **Depends:** T084
- **Action:** Modify
- **Details:** 追加 `FileFilters`

#### T105: Query Key 补充（files + ai 域）
- **File:** `src/lib/query-keys.ts`（追加）
- **Depends:** T085, T102, T103
- **Action:** Modify
- **Details:** 追加 files, ai 域

#### T106: Zod Schema 追加（aiGenerateSchema）
- **File:** `src/lib/schemas.ts`（追加）
- **Depends:** T014, T103
- **Action:** Modify
- **Details:** `aiGenerateSchema`: project_id + file_id(单文件UUID) + query + options(count, priority, include_negative)

#### T107: [P] file 服务
- **File:** `src/services/file.ts`
- **Depends:** T015, T102
- **Action:** Create
- **Details:** 8 个方法：`list`, `upload`, `uploadLarge`(分块), `addFigmaLink`, `getById`, `getIndexStatus`, `reindex`, `delete`

#### T108: [P] ai 服务
- **File:** `src/services/ai.ts`
- **Depends:** T015, T103
- **Action:** Create
- **Details:** `generate`, `getTaskStatus`, `analyze`

---

### 4.2 文件管理

#### T109: FileUploader 组件（≤10MB）
- **File:** `src/components/file/FileUploader.tsx`
- **Depends:** T003, T012
- **Action:** Create
- **Details:**
  - 文件选择 + 大小校验(`MAX_UPLOAD_SIZE`) + MIME 白名单校验
  - 上传进度条（Axios `onUploadProgress`）
  - 支持 drag & drop

#### T110: ChunkUploader 组件（>10MB 分块）
- **File:** `src/components/file/ChunkUploader.tsx`
- **Depends:** T003, T012
- **Action:** Create
- **Details:**
  - init → chunk × N(每块 5MB) → complete 三步流程
  - 每块上传进度汇总为总进度百分比
  - **注意：分块上传端点待后端实现**

#### T111: [P] IndexStatusBadge 组件
- **File:** `src/components/file/IndexStatusBadge.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 4 种状态图标：pending(灰)/processing(蓝转圈)/completed(绿✓)/failed(红✗)

#### T112: useFiles Hook
- **File:** `src/hooks/useFiles.ts`
- **Depends:** T107, T105
- **Action:** Create
- **Details:**
  - `useFileList`, `useFileDetail`, `useUploadFile`, `useDeleteFile`

#### T113: useFileIndexStatus Hook（3s 轮询）
- **File:** `src/hooks/useFileIndexStatus.ts`
- **Depends:** T107, T105
- **Action:** Create
- **Details:**
  - useQuery + `refetchInterval: 3000`，completed/failed 返回 false 停止

#### T114: 文件列表页
- **File:** `src/pages/Files/index.tsx`
- **Depends:** T109, T110, T112, T037
- **Action:** Create
- **Details:**
  - 文件表格 + 上传按钮（FileUploader/ChunkUploader 自动切换）+ Figma 链接入口
  - 空状态："还没有上传文件"

#### T115: 文件详情页
- **File:** `src/pages/Files/FileDetail.tsx`
- **Depends:** T112, T113, T111, T038
- **Action:** Create
- **Details:**
  - 基本信息 + IndexStatusBadge + 版本历史列表
  - 重新索引按钮（含 force 选项）+ 删除按钮（ConfirmDialog）

---

### 4.3 AI 功能

#### T116: GenerateForm 生成表单
- **File:** `src/components/ai/GenerateForm.tsx`
- **Depends:** T003, T106, T103
- **Action:** Create
- **Details:**
  - 文件选择下拉（仅已索引文件）+ query 文本域 + count 数字输入 + priority 选择 + include_negative 开关
  - React Hook Form + `aiGenerateSchema`

#### T117: TaskProgress 进度组件
- **File:** `src/components/ai/TaskProgress.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - 环形进度百分比 + 状态文字（pending/processing/completed/failed）
  - 脉冲动画

#### T118: useAITask Hook（2s 轮询）
- **File:** `src/hooks/useAITask.ts`
- **Depends:** T108, T105
- **Action:** Create
- **Details:**
  - useQuery + `refetchInterval: 2000`
  - `enabled: !!taskId`
  - completed/failed 返回 false 停止轮询

#### T119: AI 生成页面
- **File:** `src/pages/AIGenerate/index.tsx`
- **Depends:** T116, T117, T118, T064
- **Action:** Create
- **Details:**
  - 流程：GenerateForm 提交 → 获得 task_id → TaskProgress 轮询 → 结果列表
  - 结果列表：勾选 + 预览弹窗 + 全选/取消
  - 批量创建：`Promise.allSettled`（max 5 并发）→ 汇总成功/失败数
  - 参照 spec.md 7.9

#### T120: AI 分析页面
- **File:** `src/pages/AIAnalysis/index.tsx`
- **Depends:** T108, T003
- **Action:** Create
- **Details:**
  - 变更描述文本域 + 分析按钮
  - 结果展示：受影响用例列表（用例ID + 标题 + 原因）+ 建议

#### T121: 阶段 4 路由注册
- **File:** `src/app/router.tsx`（追加）
- **Depends:** T101, T114, T115, T119, T120
- **Action:** Modify
- **Details:** 追加 files, files/:id, ai-generate, ai-analysis 路由

### 阶段 4 验收

- [ ] 文件上传（≤10MB / >10MB 分块）
- [ ] 上传后自动轮询索引状态
- [ ] Figma 链接导入
- [ ] AI 生成：提交 → 轮询 → 结果 → 批量创建
- [ ] AI 分析：输入 → 受影响用例 + 建议

---

## Phase 5: Reports & Polish（报表与打磨）

---

### 5.1 报表页面

#### T122: [P] 报表类型
- **File:** `src/types/report.ts`
- **Depends:** —
- **Action:** Create
- **Details:** `PlanReport`, `CoverageReport`, `TrendReport`, `BugReport`, `WorkloadReport`，参照 spec.md 3.9

#### T123: [P] 筛选参数补充
- **File:** `src/types/filters.ts`（追加）
- **Depends:** T104
- **Action:** Modify
- **Details:** 追加 `BugReportParams`, `WorkloadParams`

#### T124: report 服务
- **File:** `src/services/report.ts`
- **Depends:** T015, T122
- **Action:** Create
- **Details:** `plan`, `coverage`, `trend(days=30)`, `bugs`, `workload`

#### T125: Query Key 补充（reports 域）
- **File:** `src/lib/query-keys.ts`（追加）
- **Depends:** T105, T122
- **Action:** Modify
- **Details:** 追加 reports 域（plan, coverage, trend, bugs, workload）

#### T126: useReports Hook
- **File:** `src/hooks/useReports.ts`
- **Depends:** T124, T125
- **Action:** Create
- **Details:**
  - `usePlanReport`, `useCoverageReport`, `useTrendReport`, `useBugReport`, `useWorkloadReport`

#### T127: [P] PassRateChart 饼图组件
- **File:** `src/components/report/PassRateChart.tsx`
- **Depends:** —
- **Action:** Create
- **Details:** Recharts PieChart — 通过/失败/阻塞/跳过占比

#### T128: [P] TrendChart 折线图组件
- **File:** `src/components/report/TrendChart.tsx`
- **Depends:** —
- **Action:** Create
- **Details:** Recharts LineChart — 每日通过/失败数，双线

#### T129: [P] CoverageChart 柱状图组件
- **File:** `src/components/report/CoverageChart.tsx`
- **Depends:** —
- **Action:** Create
- **Details:** Recharts BarChart — 按模块/优先级覆盖度

#### T130: 计划报告页
- **File:** `src/pages/Reports/PlanReport.tsx`
- **Depends:** T127, T126
- **Action:** Create
- **Details:** PassRateChart + 失败用例列表

#### T131: 覆盖度报告页
- **File:** `src/pages/Reports/Coverage.tsx`
- **Depends:** T129, T126
- **Action:** Create
- **Details:** CoverageChart + 按模块/优先级维度切换

#### T132: 趋势报告页
- **File:** `src/pages/Reports/Trend.tsx`
- **Depends:** T128, T126
- **Action:** Create
- **Details:** TrendChart + days 筛选器（默认 30 天，可选 7/30/90）

#### T133: 缺陷分布报告页
- **File:** `src/pages/Reports/BugDistribution.tsx`
- **Depends:** T126
- **Action:** Create
- **Details:** Recharts BarChart + top 失败用例列表

#### T134: 工作量报告页
- **File:** `src/pages/Reports/Workload.tsx`
- **Depends:** T126
- **Action:** Create
- **Details:** Recharts BarChart（每日执行量）+ 用户筛选

---

### 5.2 通用交互

#### T135: CommandPalette 命令面板
- **File:** `src/components/shared/CommandPalette.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - Cmd+K 唤起 + 搜索框 + 快捷导航列表 + 快捷操作
  - 使用 Shadcn/ui Command 组件
  - 导航：各页面路由；操作：新建用例、新建计划

#### T136: ErrorBoundary 错误边界
- **File:** `src/components/shared/ErrorBoundary.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - `componentDidCatch` 捕获渲染异常
  - 降级 UI：错误图标 + "页面出错了" + 重试按钮
  - 包裹在路由级别

#### T137: 全局 Toast 通知 + 错误处理统一
- **File:** `src/services/api.ts`（追加）
- **Depends:** T016
- **Action:** Modify
- **Details:**
  - 响应拦截器中非 401 错误统一使用 Shadcn/ui `toast`
  - 错误码自动映射 `errorMessages`
  - 400 → 表单错误由页面处理；403/404/409/429/500 → Toast

#### T138: [P] 404 页面
- **File:** `src/pages/NotFound.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:** 错误图标 + "页面不存在" + 返回首页按钮

#### T139: [P] KeyboardHelp 快捷键帮助面板
- **File:** `src/components/shared/KeyboardHelp.tsx`
- **Depends:** T003
- **Action:** Create
- **Details:**
  - `?` 键唤起 Dialog，展示全部快捷键列表
  - 分组：全局快捷键、执行页快捷键

#### T139a: 空状态完善（各页面应用）
- **File:** 各页面组件（追加）
- **Depends:** T037, T066, T080, T114
- **Action:** Modify
- **Details:**
  - 项目列表：已有（T032）
  - 用例列表："该模块下还没有用例" + 创建 / AI 生成按钮
  - 集合列表："还没有集合" + 新建按钮
  - 文件列表：已有（T114）
  - 报表页面："暂无数据"

#### T139b: 骨架屏完善（各页面定制）
- **File:** 各页面组件（追加）
- **Depends:** T023
- **Action:** Modify
- **Details:**
  - 列表页骨架屏：表格行 Skeleton（3-5 行）
  - 详情页骨架屏：标题 + 描述 + 步骤表格 Skeleton
  - 概览页骨架屏：统计卡片 + 成员列表 Skeleton
  - 在 `React.lazy` fallback 中使用对应页面骨架屏（替换通用 Skeleton）

---

### 5.3 响应式与性能

#### T140: ProjectSidebar 响应式折叠改造
- **File:** `src/components/layout/ProjectSidebar.tsx`（改造）
- **Depends:** T024
- **Action:** Modify
- **Details:**
  - `useMediaQuery("(min-width: 1024px)")` 检测屏幕宽度
  - 1024px 以下自动折叠为图标模式
  - 手动折叠/展开优化动画

#### T141: TestCaseTable 响应式适配
- **File:** `src/components/testcase/TestCaseTable.tsx`（追加）
- **Depends:** T058
- **Action:** Modify
- **Details:** 窄屏隐藏次要列（更新时间、创建人等），保留标题/状态/优先级核心列

#### T142: Vite chunk 分割优化 + Recharts 按需加载
- **File:** `vite.config.ts`（追加）
- **Depends:** T007
- **Action:** Modify
- **Details:**
  - `build.rollupOptions.output.manualChunks`: `{ recharts: ["recharts"], "dnd-kit": ["@dnd-kit/core", "@dnd-kit/sortable"] }`
  - 确保报表页面通过 `const Recharts = lazy(() => import("recharts"))` 动态导入，recharts 不进入主 bundle

#### T143: 最终路由注册
- **File:** `src/app/router.tsx`（最终版）
- **Depends:** T121, T130-T134, T135, T136, T138
- **Action:** Modify
- **Details:**
  - 追加 reports/plan, reports/coverage, reports/trend, reports/bugs, reports/workload 路由
  - 追加全局 `*` → NotFound 路由
  - ErrorBoundary 包裹各路由

---

### 5.4 构建部署验证

#### T143a: [P] Dockerfile（多阶段构建）
- **File:** `Dockerfile`
- **Depends:** —
- **Action:** Create
- **Details:**
  - Stage 1: `node:18-alpine` — `npm ci && npm run build`
  - Stage 2: `nginx:alpine` — 拷贝 dist + nginx.conf
  - 最终镜像 < 50MB
  - 参照 spec.md 14.3

#### T144: Nginx SPA 路由回退
- **File:** `nginx.conf`
- **Depends:** —
- **Action:** Create
- **Details:** `location / { try_files $uri $uri/ /index.html; }` + API 反向代理

#### T145: 生产构建验证
- **Depends:** T143, T144
- **Action:** Verify
- **Details:**
  - `npm run build` 无报错
  - 验证 chunk 分割：recharts/dnd-kit 独立 chunk
  - 验证 gzip 后首屏 JS < 200KB

### 阶段 5 验收

- [ ] 5 个报表页面图表正确展示
- [ ] 趋势报告默认 30 天，可切换
- [ ] Cmd+K 命令面板搜索导航
- [ ] ErrorBoundary 捕获渲染异常
- [ ] 全局错误 Toast 自动弹出
- [ ] `npm run build` 通过，无报错
- [ ] Recharts 独立 chunk

---

## 任务统计

| Phase | 任务数 | 说明 |
|-------|--------|------|
| Phase 1: Foundation | 40 | 项目骨架 + 类型 + 工具 + API + Store + 路由 + 登录 + 项目列表 |
| Phase 2: Core Test Cases | 42 | 模块 + 标签 + 用例 CRUD + 步骤 + 批量 + 集合 |
| Phase 3: Plan & Execution | 21 | 计划 + 执行 + 键盘快捷键 + 执行记录 |
| Phase 4: Files & AI | 20 | 文件上传/分块 + AI 生成/分析 |
| Phase 5: Reports & Polish | 27 | 报表 + 命令面板 + ErrorBoundary + 响应式 + 构建 |
| **合计** | **150** | |

---

## 依赖关系总览

```
Phase 1 (T001-T039)
  ↓
Phase 2 (T040-T081) — 依赖 Phase 1 的 API 层、Store、布局
  ↓
Phase 3 (T082-T101) — 依赖 Phase 2 的用例组件
  ↓
Phase 4 (T102-T121) — 依赖 Phase 1 的 API 层 + Phase 2 的用例 Hook（T119→T064）
  ↓
Phase 5 (T122-T145) — 依赖 Phase 3 的执行数据 + Phase 4 的文件/AI
```

**Phase 内并行性**：
- Phase 1: T008-T010 之间 [P], T011-T014 之间 [P], T018-T019 之间 [P], T020-T021 之间 [P], T035-T039 之间 [P]
- Phase 2: T040-T044 之间 [P], T046-T049 之间 [P], T056-T057 之间 [P], T060-T062 之间 [P], T076-T078 之间 [P]
- Phase 3: T082-T083 之间 [P], T087-T088 之间 [P]
- Phase 4: T102-T104 之间 [P], T107-T108 之间 [P]
- Phase 5: T122-T123 之间 [P], T127-T129 之间 [P], T138-T139 之间 [P], T143a [P]

---

**文档版本**：v1.1
**最后更新**：2026-05-18
