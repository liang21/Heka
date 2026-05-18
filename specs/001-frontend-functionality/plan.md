# Heka 前端实现计划

> 基于 `specs/001-frontend-functionality/spec.md` v1.1 拆解的详细任务清单
> 版本：v1.1（review 修复）
> 日期：2026-05-18

---

## 说明

- 每个任务标注 **P**(Page) / **C**(Component) / **H**(Hook) / **S**(Service) / **T**(Type) / **St**(Store) / **L**(Lib) 层
- 同一小节内任务按表格从上到下的顺序即隐含依赖关系（先定义后实现，先底层后上层）
- 预估工时为单人开发估计，含自测
- 验收标准（AC）= 该阶段完成时可演示的功能

---

## 阶段 1：基础骨架（1 周）

> 交付目标：Vite 项目启动、登录可用、项目列表与概览展示

### 1.1 项目初始化 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.1.1 | Vite + React + TypeScript 脚手架 | — | `package.json`, `vite.config.ts`, `tsconfig.json` | 可运行的空项目 |
| 1.1.2 | Tailwind CSS + Shadcn/ui 集成 | L | `tailwind.config.ts`, `components.json`, `src/styles/globals.css` | Shadcn 组件可引入 |
| 1.1.3 | Shadcn/ui 基础组件引入 | C | `src/components/ui/*` | Button, Card, Input, Dialog, Form, Label, Skeleton 等 |
| 1.1.4 | 目录结构搭建 | — | `src/` 下全部子目录 | 空目录 + index 占位 |
| 1.1.5 | 路径别名 `@/` 配置 | L | `vite.config.ts`, `tsconfig.json` | `@/components/...` 可解析 |
| 1.1.6 | Lucide React 图标引入 | L | `package.json` | 图标可用 |
| 1.1.7 | 环境变量文件 | — | `.env`, `.env.production` | `VITE_API_BASE_URL`, `VITE_APP_TITLE` |
| 1.1.8 | Vite 代理配置（`/api` → 后端） | L | `vite.config.ts`（追加） | 开发环境 API 代理，阶段 1 即可联调 |
| 1.1.9 | Nginx 配置 + Dockerfile | — | `nginx.conf`, `Dockerfile` | Docker 构建就绪（开发阶段暂不使用） |

### 1.2 类型与工具 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.2.1 | 通用类型定义 | T | `src/types/api.ts` | `ApiResponse`, `PaginatedData`, `PaginationParams` |
| 1.2.2 | 认证类型 | T | `src/types/auth.ts` | `User`, `LoginRequest`, `LoginResponse` |
| 1.2.3 | 项目类型 | T | `src/types/project.ts` | `Project`, `ProjectDetail`, `ProjectMember`, `ProjectStatistics` |
| 1.2.4 | 工具函数 | L | `src/lib/utils.ts` | `cn()` (classnames + tailwind-merge) |
| 1.2.5 | 常量定义 | L | `src/lib/constants.ts` | `MAX_UPLOAD_SIZE`, `DEFAULT_PAGE_SIZE` 等 |
| 1.2.6 | 错误码映射 | L | `src/lib/error-messages.ts` | `errorMessages` Record 全量错误码 → 中文 |
| 1.2.7 | Zod Schema（loginSchema） | L | `src/lib/schemas.ts` | 登录表单校验（email + password） |
| 1.2.8 | Zod Schema（createProjectSchema） | L | `src/lib/schemas.ts`（追加） | 项目创建校验（name + description） |

### 1.3 API 基础层 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.3.1 | Axios 实例 + 请求拦截器（注入 Bearer Token） | S | `src/services/api.ts` | `api` 实例 |
| 1.3.2 | 响应拦截器（401 静默刷新 + 排队重试） | S | `src/services/api.ts`（追加） | Token 过期自动刷新，并发请求排队 |
| 1.3.3 | auth 服务 | S | `src/services/auth.ts` | `login`, `getMe`, `refresh` |
| 1.3.4 | project 服务 | S | `src/services/project.ts` | `list`, `create`, `getById`, `addMember` |
| 1.3.5 | Query Key 工厂 | L | `src/lib/query-keys.ts` | `queryKeys` 全量定义（阶段 1 仅 projects 域） |
| 1.3.6 | QueryClient 默认配置 | L | `src/app/providers.tsx` | `staleTime` / `gcTime` / `retry` 全局默认值 |

### 1.4 Store [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.4.1 | auth store（persist 中间件，partialize 仅 token） | St | `src/stores/auth.ts` | `useAuthStore`: token/user 状态 + setAuth/clearAuth |
| 1.4.2 | project store（persist，currentProjectId + sidebarCollapsed） | St | `src/stores/project.ts` | `useProjectStore`: 当前项目 + 侧边栏状态 |

### 1.5 路由与布局 [0.75d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.5.1 | React Router 配置骨架 | P | `src/app/router.tsx` | `/login` + `/` 嵌套路由结构 |
| 1.5.2 | 路由守卫（未登录跳转 /login） | H | `src/hooks/useAuthInit.ts` | `AuthGuard` 组件 + `useAuthInit` Hook |
| 1.5.3 | 懒加载包装组件 | C | `src/components/shared/LazyLoad.tsx` | `Suspense` + 骨架屏 fallback |
| 1.5.4 | AppLayout（TopBar + Sidebar 骨架） | C | `src/components/layout/AppLayout.tsx` | 应用布局壳 |
| 1.5.5 | ProjectSidebar 组件 | C | `src/components/layout/ProjectSidebar.tsx` | 项目导航菜单 + 折叠交互 |
| 1.5.6 | Breadcrumb 面包屑组件 | C | `src/components/layout/Breadcrumb.tsx` | 多级面包屑导航 |
| 1.5.7 | ProjectLayout（项目侧边栏 + 内容区） | C | `src/components/layout/ProjectLayout.tsx` | 引用 ProjectSidebar + Breadcrumb |
| 1.5.8 | Providers 聚合 | L | `src/app/providers.tsx` | `QueryClientProvider` + `BrowserRouter` |
| 1.5.9 | App 入口 | P | `src/app/App.tsx` | 挂载 Providers + Router |

### 1.6 页面组件 [1.25d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 1.6.1 | 登录页 | P | `src/pages/Login/index.tsx` | 表单 + Zod 校验（loginSchema）+ useMutation 登录 |
| 1.6.2 | 项目列表页（首页） | P | `src/pages/Projects/index.tsx` | 卡片网格 + 新建 Dialog（createProjectSchema）+ 空状态 |
| 1.6.3 | useProject Hook | H | `src/hooks/useProject.ts` | 获取当前项目详情 |
| 1.6.4 | 项目概览页 | P | `src/pages/Overview/index.tsx` | 统计卡片 + 成员列表 + 快捷操作 |
| 1.6.5 | 项目设置页 | P | `src/pages/Settings/ProjectSettings.tsx` | 基本信息编辑表单 |
| 1.6.6 | 系统设置页 | P | `src/pages/Settings/index.tsx` | 用户信息展示 + 主题切换入口（占位） |
| 1.6.7 | 空状态组件 | C | `src/components/shared/EmptyState.tsx` | 图标 + 文案 + 操作按钮 |
| 1.6.8 | ConfirmDialog 组件 | C | `src/components/shared/ConfirmDialog.tsx` | 通用确认弹窗 |
| 1.6.9 | useConfirm Hook | H | `src/hooks/useConfirm.ts` | 命令式调用 ConfirmDialog |

### 阶段 1 验收

- [ ] `npm run dev` 启动无报错，页面正常渲染
- [ ] 登录页：输入邮箱密码 → 调用 API → 存储 Token → 跳转首页
- [ ] Token 过期自动刷新，刷新失败跳转登录页
- [ ] 项目列表页：展示项目卡片，可创建新项目
- [ ] 点击项目卡片 → 进入项目概览页，展示统计信息和成员
- [ ] 项目设置页可编辑基本信息
- [ ] 系统设置页可查看用户信息
- [ ] 未登录访问任何页面自动跳转 `/login`
- [ ] 响应拦截器 401 处理正常（静默刷新 + 排队）

---

## 阶段 2：核心用例管理（2 周）

> 交付目标：用例 CRUD 全流程、模块树、批量操作、集合

### 2.1 类型与服务准备 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.1.1 | 模块类型 | T | `src/types/module.ts` | `Module` 接口 |
| 2.1.2 | 测试用例类型 | T | `src/types/testcase.ts` | `TestCase`, `TestStep`, `TestCaseStatus`, `Priority`, `TestCaseListParams` |
| 2.1.3 | 标签类型 | T | `src/types/tag.ts` | `Tag` 接口 |
| 2.1.4 | 集合类型 | T | `src/types/collection.ts` | `Collection` 接口 |
| 2.1.5 | 筛选参数类型 | T | `src/types/filters.ts` | `TestCaseFilters`, `CollectionCaseFilters` |
| 2.1.6 | module 服务 | S | `src/services/module.ts` | `tree`, `create`, `update`, `delete` |
| 2.1.7 | testcase 服务 | S | `src/services/testcase.ts` | 8 个方法（含 batch 操作） |
| 2.1.8 | tag 服务 | S | `src/services/tag.ts` | `list`, `create` |
| 2.1.9 | collection 服务 | S | `src/services/collection.ts` | 6 个方法 |
| 2.1.10 | Query Key 补充 | L | `src/lib/query-keys.ts`（追加） | modules, testCases, tags, collections 域 |

### 2.2 模块管理 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.2.1 | ModuleTree 递归树组件 | C | `src/components/module/ModuleTree.tsx` | 展开/折叠 + 节点渲染 + 选中高亮 |
| 2.2.2 | @dnd-kit 拖拽排序集成 | C | `src/components/module/ModuleTree.tsx`（追加） | 拖拽排序 + 层级移动 |
| 2.2.3 | ModuleForm 表单组件 | C | `src/components/module/ModuleForm.tsx` | 名称 + 描述 + 父模块选择 |
| 2.2.4 | useModules Hook | H | `src/hooks/useModules.ts` | `useModuleTree`, `useCreateModule`, `useUpdateModule`, `useDeleteModule` |
| 2.2.5 | 模块管理页面 | P | `src/pages/Modules/index.tsx` | ModuleTree + 新建/编辑/删除操作 |
| 2.2.6 | 创建模块页面 | P | `src/pages/Modules/CreateModule.tsx` | ModuleForm 独立页 |

### 2.3 测试用例列表 [2.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.3.1 | PriorityBadge 组件 | C | `src/components/testcase/PriorityBadge.tsx` | 4 级优先级可视化 |
| 2.3.2 | StatusTag 组件 | C | `src/components/testcase/StatusTag.tsx` | 用例状态标签 |
| 2.3.3 | TestCaseTable 组件 | C | `src/components/testcase/TestCaseTable.tsx` | 行选择 checkbox + 排序列头 + 状态/优先级列 |
| 2.3.4 | TestCaseFilters 筛选组件 | C | `src/components/testcase/TestCaseFilters.tsx` | status/priority/module/keyword/tags 筛选栏 |
| 2.3.5 | BatchActions 批量操作栏 | C | `src/components/testcase/BatchActions.tsx` | 批量改状态 / 移动 / 删除 |
| 2.3.6 | Pagination 分页组件 | C | `src/components/shared/Pagination.tsx` | 页码 + 每页条数选择 |
| 2.3.7 | usePagination Hook | H | `src/hooks/usePagination.ts` | URL 参数同步（page/page_size） |
| 2.3.8 | SearchInput 组件 | C | `src/components/shared/SearchInput.tsx` | 防抖搜索输入框 |
| 2.3.9 | useTestCases Hook | H | `src/hooks/useTestCases.ts` | `useTestCaseList` + `useBatchUpdateStatus` + `useBatchDelete` + `useBatchMove` |
| 2.3.10 | 测试用例列表页 | P | `src/pages/TestCases/index.tsx` | 三栏布局：左侧模块树(260px) + 筛选栏 + 表格 + 分页 |
| 2.3.11 | URL 参数双向同步 | H | `src/hooks/useTestCaseFilters.ts` | 筛选状态 ↔ URL query params |

### 2.4 创建/编辑测试用例 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.4.1 | Zod Schema（createTestCaseSchema） | L | `src/lib/schemas.ts` | 用例创建/编辑校验规则 |
| 2.4.2 | StepEditor 步骤编辑器 | C | `src/components/testcase/StepEditor.tsx` | 动态数组增删 + 拖拽排序 |
| 2.4.3 | TagInput 标签输入组件 | C | `src/components/testcase/TagInput.tsx` | 已有标签下拉 + 自定义输入 |
| 2.4.4 | useUnsavedChanges Hook | H | `src/hooks/useUnsavedChanges.ts` | `useBlocker` + `beforeunload` 离开拦截 |
| 2.4.5 | 创建用例页面 | P | `src/pages/TestCases/CreateTestCase.tsx` | React Hook Form + Zod + 提交 |
| 2.4.6 | 编辑用例页面（含 version 乐观锁） | P | `src/pages/TestCases/EditTestCase.tsx` | 加载时记录 version，提交携带，409 冲突提示 |

### 2.5 用例详情 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.5.1 | 用例详情页 | P | `src/pages/TestCases/TestCaseDetail.tsx` | 面包屑 + 标题 + 描述 + 步骤表格 + 执行历史 |
| 2.5.2 | 状态流转按钮（draft/ready/archived） | C | `src/components/testcase/StatusActions.tsx` | 根据当前状态显示可用操作 |
| 2.5.3 | 编辑模式切换 | C | `src/pages/TestCases/TestCaseDetail.tsx`（追加） | 点击"编辑"→ 标题/描述/步骤可编辑 |

### 2.6 标签管理 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.6.1 | useTags Hook | H | `src/hooks/useTags.ts` | `useTagList`, `useCreateTag` |
| 2.6.2 | 标签管理（集成到用例筛选栏） | C | `src/components/testcase/TestCaseFilters.tsx`（追加） | 标签下拉选择 + 新建标签 |

### 2.7 用例集合 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.7.1 | CollectionCard 组件 | C | `src/components/collection/CollectionCard.tsx` | 集合卡片展示 |
| 2.7.2 | AddCasesDialog 组件 | C | `src/components/collection/AddCasesDialog.tsx` | 从用例列表勾选添加 |
| 2.7.3 | CollectionCasesList 组件 | C | `src/components/collection/CollectionCasesList.tsx` | 集合内用例列表 + 移除操作 |
| 2.7.4 | useCollections Hook | H | `src/hooks/useCollections.ts` | CRUD + 添加/移除用例 |
| 2.7.5 | 集合列表页 | P | `src/pages/TestCases/Collections.tsx` | 集合卡片网格 + 新建 Dialog |

### 2.8 路由整合 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 2.8.1 | 阶段 2 路由注册 | P | `src/app/router.tsx`（追加） | modules / testcases / collections 路由 |

### 阶段 2 验收

- [ ] 模块树 CRUD，支持多级嵌套 + 拖拽排序
- [ ] 标签创建 + 项目标签列表下拉
- [ ] 用例列表：分页、筛选（status/priority/module/keyword/tags）、排序
- [ ] 用例创建：表单校验 + 步骤编辑器 + 标签输入
- [ ] 用例编辑：携带 version，409 冲突提示
- [ ] 用例详情：展示步骤 + 执行历史 + 状态流转按钮
- [ ] 批量状态更新、批量删除、批量移动
- [ ] 集合创建 + 添加/移除用例
- [ ] 离开编辑页未保存提示
- [ ] URL 参数同步筛选状态（刷新页面保持筛选条件）

---

## 阶段 3：计划与执行（1.5 周）

> 交付目标：计划创建 → 关联用例 → 执行 → 提交结果 → 状态流转

### 3.1 类型与服务准备 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.1.1 | 计划类型 | T | `src/types/plan.ts` | `TestPlan`, `PlanDetail`, `PlanTestCase`, `PlanProgress`, `ExecutionSummary`, `PlanStatus` |
| 3.1.2 | 执行类型 | T | `src/types/execution.ts` | `Execution`, `ExecutionDetail`, `ExecutionResult`, `ExecutionStatus`, `ResultStatus`, `SubmitResultRequest` |
| 3.1.3 | 筛选参数类型补充 | T | `src/types/filters.ts`（追加） | `PlanFilters`, `ExecutionFilters` |
| 3.1.4 | plan 服务 | S | `src/services/plan.ts` | 8 个方法（含状态流转端点） |
| 3.1.5 | execution 服务 | S | `src/services/execution.ts` | `getById`, `submitResult`, `submitBatchResults` |
| 3.1.6 | Query Key 补充 | L | `src/lib/query-keys.ts`（追加） | plans, executions 域 |
| 3.1.7 | Zod Schema（createPlanSchema） | L | `src/lib/schemas.ts`（追加） | 计划创建校验 |

### 3.2 测试计划 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.2.1 | usePlans Hook | H | `src/hooks/usePlans.ts` | `usePlanList`, `usePlanDetail`, `useCreatePlan`, `usePlanAction`(start/pause/resume/complete/cancel) |
| 3.2.2 | 测试计划列表页 | P | `src/pages/TestPlans/index.tsx` | 计划表格 + 状态筛选 + 新建按钮 |
| 3.2.3 | 创建计划页面 | P | `src/pages/TestPlans/CreatePlan.tsx` | 表单 + 用例多选 + 执行人分配 |
| 3.2.4 | AssignSelector 执行人选择组件 | C | `src/components/plan/AssignSelector.tsx` | 项目成员下拉列表 |
| 3.2.5 | 计划详情页 | P | `src/pages/TestPlans/PlanDetail.tsx` | 状态按钮（根据 9.2 节流转）+ 用例列表 + 执行历史 |
| 3.2.6 | 并发执行预检逻辑 | L | `src/pages/TestPlans/PlanDetail.tsx`（追加） | 点击"开始执行"前检查 in_progress 执行，冲突时弹窗提供选择 |

### 3.3 执行管理 [2.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.3.1 | ExecutionPanel 执行面板 | C | `src/components/plan/ExecutionPanel.tsx` | 步骤详情展示 + 4 个结果按钮 |
| 3.3.2 | ResultSummary 结果汇总 | C | `src/components/plan/ResultSummary.tsx` | passed/failed/blocked/skipped 统计 |
| 3.3.3 | useExecution Hook | H | `src/hooks/useExecution.ts` | `useExecutionDetail`, `useSubmitResult`, `useSubmitBatchResults` |
| 3.3.4 | useKeyboard Hook（执行页快捷键） | H | `src/hooks/useKeyboard.ts` | 执行页 P/F/B/S/Enter 快捷键注册（全局快捷键在阶段 5 扩展） |
| 3.3.5 | 执行页面（左右分栏） | P | `src/pages/TestPlans/ExecutePlan.tsx` | 左 40% 用例列表 + 右 60% 执行面板 + 底部进度条 |
| 3.3.6 | 键盘快捷键（P/F/B/S/Enter） | L | `src/pages/TestPlans/ExecutePlan.tsx`（追加） | 执行页快捷键绑定 |
| 3.3.7 | Bug ID 输入（失败扩展） | C | `src/components/plan/ExecutionPanel.tsx`（追加） | 选择"失败"时展开 Bug ID 输入框 |
| 3.3.8 | 自动跳转下一个用例 | L | `src/pages/TestPlans/ExecutePlan.tsx`（追加） | 提交结果后自动选中下一个未执行用例 |

### 3.4 执行记录 [1d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.4.1 | 执行记录列表页 | P | `src/pages/Executions/index.tsx` | 执行记录表格 + 状态筛选 |
| 3.4.2 | 执行记录详情页 | P | `src/pages/Executions/ExecutionDetail.tsx` | 执行信息 + 结果列表 + 汇总统计 |

### 3.5 路由整合 [0.25d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 3.5.1 | 阶段 3 路由注册 | P | `src/app/router.tsx`（追加） | plans / executions 路由 |

### 阶段 3 验收

- [ ] 创建测试计划并关联用例 + 分配执行人
- [ ] 计划状态流转按钮根据当前状态动态显示（draft → active → paused → completed/cancelled）
- [ ] 同一计划并发执行被拦截（前端预检 + 后端 409 兜底）
- [ ] 执行页面左右分栏，键盘快捷键（P/F/B/S/Enter）正常工作
- [ ] 提交结果后自动跳转下一个用例
- [ ] 失败时展开 Bug ID 输入框
- [ ] 执行记录列表 + 详情可查看
- [ ] 乐观锁冲突（409）弹出冲突提示

---

## 阶段 4：文件与 AI（1.5 周）

> 交付目标：文件上传、索引状态、AI 生成用例、智能分析

### 4.1 类型与服务准备 [0.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.1.1 | 文件类型 | T | `src/types/file.ts` | `ProjectFile`, `FileSourceType`, `FileIndexStatus` |
| 4.1.2 | AI 类型 | T | `src/types/ai.ts` | `AITask`, `AITaskStatus`, `AIGeneratedCase`, `AIAnalysisResult` |
| 4.1.3 | 筛选参数类型补充 | T | `src/types/filters.ts`（追加） | `FileFilters` |
| 4.1.4 | file 服务 | S | `src/services/file.ts` | 8 个方法（含分块上传） |
| 4.1.5 | ai 服务 | S | `src/services/ai.ts` | `generate`, `getTaskStatus`, `analyze` |
| 4.1.6 | Query Key 补充 | L | `src/lib/query-keys.ts`（追加） | files, ai 域 |
| 4.1.7 | Zod Schema（aiGenerateSchema） | L | `src/lib/schemas.ts`（追加） | AI 生成表单校验 |

### 4.2 文件管理 [2.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.2.1 | FileUploader 组件（≤10MB） | C | `src/components/file/FileUploader.tsx` | 文件选择 + 大小校验 + MIME 校验 + 进度条 |
| 4.2.2 | ChunkUploader 组件（>10MB 分块） | C | `src/components/file/ChunkUploader.tsx` | init → chunk × N → complete，5MB 分块，进度条 |
| 4.2.3 | IndexStatusBadge 组件 | C | `src/components/file/IndexStatusBadge.tsx` | pending/processing/completed/failed 状态图标 |
| 4.2.4 | useFiles Hook | H | `src/hooks/useFiles.ts` | `useFileList`, `useFileDetail`, `useUploadFile`, `useDeleteFile` |
| 4.2.5 | useFileIndexStatus Hook（轮询） | H | `src/hooks/useFileIndexStatus.ts` | 3s 间隔轮询，completed/failed 停止 |
| 4.2.6 | 文件列表页 | P | `src/pages/Files/index.tsx` | 文件表格 + 上传按钮 + Figma 链接入口 + 空状态 |
| 4.2.7 | 文件详情页 | P | `src/pages/Files/FileDetail.tsx` | 索引状态 + 版本历史 + 重新索引 + 删除 |

### 4.3 AI 功能 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.3.1 | GenerateForm 生成表单 | C | `src/components/ai/GenerateForm.tsx` | 文件选择 + query + count + priority + include_negative |
| 4.3.2 | TaskProgress 进度组件 | C | `src/components/ai/TaskProgress.tsx` | 进度百分比 + 状态文字 + 动画 |
| 4.3.3 | useAITask Hook（2s 轮询） | H | `src/hooks/useAITask.ts` | `refetchInterval` 轮询，completed/failed 停止 |
| 4.3.4 | AI 生成页面 | P | `src/pages/AIGenerate/index.tsx` | GenerateForm → 提交 → 轮询进度 → 结果列表 → 批量创建 |
| 4.3.5 | AI 结果预览与勾选 | C | `src/pages/AIGenerate/index.tsx`（追加） | 生成结果列表：勾选 + 预览弹窗 + 全选/取消 |
| 4.3.6 | 批量创建（Promise.allSettled，max 5 并发） | L | `src/pages/AIGenerate/index.tsx`（追加） | 并发创建用例 + 成功/失败汇总 |
| 4.3.7 | AI 分析页面 | P | `src/pages/AIAnalysis/index.tsx` | 变更描述输入 → 分析结果展示（受影响用例 + 建议） |

### 4.4 路由整合 [0.25d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 4.4.1 | 阶段 4 路由注册 | P | `src/app/router.tsx`（追加） | files / ai-generate / ai-analysis 路由 |

### 阶段 4 验收

- [ ] 文件上传：≤10MB 直接上传，>10MB 分块上传
- [ ] 上传后自动轮询索引状态，状态变化实时反映
- [ ] 文件详情页展示索引状态、版本历史
- [ ] Figma 链接导入
- [ ] AI 用例生成：提交 → 轮询进度（2s）→ 结果展示 → 批量创建
- [ ] 批量创建用例汇总成功/失败数量
- [ ] AI 分析：输入描述 → 返回受影响用例列表和建议

---

## 阶段 5：报表与打磨（1 周）

> 交付目标：完整报表、命令面板、错误处理完善、响应式

### 5.1 报表页面 [3d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.1.1 | 报表类型 | T | `src/types/report.ts` | `PlanReport`, `CoverageReport`, `TrendReport`, `BugReport`, `WorkloadReport` |
| 5.1.2 | 筛选参数类型补充 | T | `src/types/filters.ts`（追加） | `BugReportParams`, `WorkloadParams` |
| 5.1.3 | report 服务 | S | `src/services/report.ts` | `plan`, `coverage`, `trend`, `bugs`, `workload` |
| 5.1.4 | Query Key 补充 | L | `src/lib/query-keys.ts`（追加） | reports 域 |
| 5.1.5 | useReports Hook | H | `src/hooks/useReports.ts` | `usePlanReport`, `useCoverageReport`, `useTrendReport`, `useBugReport`, `useWorkloadReport` |
| 5.1.6 | PassRateChart 通过率饼图 | C | `src/components/report/PassRateChart.tsx` | Recharts PieChart |
| 5.1.7 | TrendChart 趋势折线图 | C | `src/components/report/TrendChart.tsx` | Recharts LineChart（每日通过/失败） |
| 5.1.8 | CoverageChart 覆盖度柱状图 | C | `src/components/report/CoverageChart.tsx` | Recharts BarChart（按模块/优先级） |
| 5.1.9 | 计划报告页 | P | `src/pages/Reports/PlanReport.tsx` | PassRateChart + 失败用例列表 |
| 5.1.10 | 覆盖度报告页 | P | `src/pages/Reports/Coverage.tsx` | CoverageChart |
| 5.1.11 | 趋势报告页 | P | `src/pages/Reports/Trend.tsx` | TrendChart + days 筛选（默认 30 天） |
| 5.1.12 | 缺陷分布报告页 | P | `src/pages/Reports/BugDistribution.tsx` | Recharts BarChart + top 失败用例 |
| 5.1.13 | 工作量报告页 | P | `src/pages/Reports/Workload.tsx` | Recharts BarChart（每日执行量） |

### 5.2 通用交互 [2d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.2.1 | CommandPalette 命令面板 | C | `src/components/shared/CommandPalette.tsx` | Cmd+K 唤起 + 搜索 + 快捷导航 + 快捷操作 |
| 5.2.2 | ErrorBoundary 错误边界 | C | `src/components/shared/ErrorBoundary.tsx` | 捕获渲染异常 + 降级 UI + 重试按钮 |
| 5.2.3 | 全局 Toast 通知 | C | `src/components/shared/ToastProvider.tsx` | Shadcn Toast + 错误码自动映射 |
| 5.2.4 | 全局错误处理统一 | L | `src/services/api.ts`（追加） | 非 401 错误统一 Toast 提示 |
| 5.2.5 | 404 页面 | P | `src/pages/NotFound.tsx` | 返回首页按钮 |
| 5.2.6 | 空状态完善 | C | 各页面（追加） | 项目列表/用例列表/文件列表/报表空状态 |
| 5.2.7 | 骨架屏完善 | C | 各页面（追加） | 首次加载骨架屏 |
| 5.2.8 | 快捷键帮助面板 | C | `src/components/shared/KeyboardHelp.tsx` | `?` 键唤起，展示全部快捷键 |

### 5.3 响应式与性能 [1.5d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.3.1 | 侧边栏响应式折叠改造 | C | `src/components/layout/ProjectSidebar.tsx`（改造） | 1024px 以下自动折叠 + 手动切换优化 |
| 5.3.2 | 表格响应式适配 | C | `src/components/testcase/TestCaseTable.tsx`（追加） | 窄屏隐藏次要列 |
| 5.3.3 | Vite chunk 分割优化 | L | `vite.config.ts`（追加） | `manualChunks`: recharts / dnd-kit 独立 chunk |
| 5.3.4 | Recharts 按需加载 | L | 报表页面 | 仅报表页引入，不进入主 bundle |
| 5.3.5 | 路由整合 + 最终路由表 | P | `src/app/router.tsx`（最终版） | 全部路由注册 |

### 5.4 构建部署验证 [0.25d]

| # | 任务 | 层 | 文件 | 产出 |
|---|------|----|------|------|
| 5.4.1 | `npm run build` 无报错 | — | — | 生产构建通过 |
| 5.4.2 | Nginx SPA 路由回退 | — | `nginx.conf` | `try_files $uri /index.html` |

### 阶段 5 验收

- [ ] 5 个报表页面：计划报告（饼图）、覆盖度（柱状图）、趋势（折线图）、缺陷分布、工作量
- [ ] 趋势报告默认 30 天，可切换天数
- [ ] 命令面板 Cmd+K 可搜索导航 + 快捷操作
- [ ] 渲染异常被 ErrorBoundary 捕获，显示降级 UI
- [ ] 404 页面正确显示
- [ ] 全局错误 Toast 自动弹出
- [ ] 快捷键帮助面板 `?` 键唤起
- [ ] 1024px 以下侧边栏自动折叠
- [ ] `npm run build` 构建成功，无报错
- [ ] Recharts 独立 chunk，不影响首屏加载

---

## 文件清单汇总

### 页面组件（src/pages/）

| 页面 | 文件路径 | 阶段 |
|------|----------|------|
| 登录 | `src/pages/Login/index.tsx` | 1 |
| 项目列表 | `src/pages/Projects/index.tsx` | 1 |
| 项目概览 | `src/pages/Overview/index.tsx` | 1 |
| 系统设置 | `src/pages/Settings/index.tsx` | 1 |
| 项目设置 | `src/pages/Settings/ProjectSettings.tsx` | 1 |
| 模块管理 | `src/pages/Modules/index.tsx` | 2 |
| 创建模块 | `src/pages/Modules/CreateModule.tsx` | 2 |
| 用例列表 | `src/pages/TestCases/index.tsx` | 2 |
| 创建用例 | `src/pages/TestCases/CreateTestCase.tsx` | 2 |
| 编辑用例 | `src/pages/TestCases/EditTestCase.tsx` | 2 |
| 用例详情 | `src/pages/TestCases/TestCaseDetail.tsx` | 2 |
| 集合列表 | `src/pages/TestCases/Collections.tsx` | 2 |
| 计划列表 | `src/pages/TestPlans/index.tsx` | 3 |
| 创建计划 | `src/pages/TestPlans/CreatePlan.tsx` | 3 |
| 计划详情 | `src/pages/TestPlans/PlanDetail.tsx` | 3 |
| 执行页面 | `src/pages/TestPlans/ExecutePlan.tsx` | 3 |
| 执行列表 | `src/pages/Executions/index.tsx` | 3 |
| 执行详情 | `src/pages/Executions/ExecutionDetail.tsx` | 3 |
| 文件列表 | `src/pages/Files/index.tsx` | 4 |
| 文件详情 | `src/pages/Files/FileDetail.tsx` | 4 |
| AI 生成 | `src/pages/AIGenerate/index.tsx` | 4 |
| AI 分析 | `src/pages/AIAnalysis/index.tsx` | 4 |
| 计划报告 | `src/pages/Reports/PlanReport.tsx` | 5 |
| 覆盖度 | `src/pages/Reports/Coverage.tsx` | 5 |
| 趋势 | `src/pages/Reports/Trend.tsx` | 5 |
| 缺陷分布 | `src/pages/Reports/BugDistribution.tsx` | 5 |
| 工作量 | `src/pages/Reports/Workload.tsx` | 5 |
| 404 | `src/pages/NotFound.tsx` | 5 |

### 自定义 Hooks（src/hooks/）

| Hook | 文件 | 阶段 |
|------|------|------|
| useAuthInit | `src/hooks/useAuthInit.ts` | 1 |
| useProject | `src/hooks/useProject.ts` | 1 |
| useConfirm | `src/hooks/useConfirm.ts` | 1 |
| usePagination | `src/hooks/usePagination.ts` | 2 |
| useTestCases | `src/hooks/useTestCases.ts` | 2 |
| useTestCaseFilters | `src/hooks/useTestCaseFilters.ts` | 2 |
| useModules | `src/hooks/useModules.ts` | 2 |
| useTags | `src/hooks/useTags.ts` | 2 |
| useCollections | `src/hooks/useCollections.ts` | 2 |
| useUnsavedChanges | `src/hooks/useUnsavedChanges.ts` | 2 |
| usePlans | `src/hooks/usePlans.ts` | 3 |
| useExecution | `src/hooks/useExecution.ts` | 3 |
| useKeyboard | `src/hooks/useKeyboard.ts` | 3 |
| useFiles | `src/hooks/useFiles.ts` | 4 |
| useFileIndexStatus | `src/hooks/useFileIndexStatus.ts` | 4 |
| useAITask | `src/hooks/useAITask.ts` | 4 |
| useReports | `src/hooks/useReports.ts` | 5 |

---

## 工时汇总

| 阶段 | 子任务合计 | 含缓冲日历天 |
|------|-----------|-------------|
| 阶段 1：基础骨架 | 4.5d | 1 周 |
| 阶段 2：核心用例管理 | 10d | 2 周 |
| 阶段 3：计划与执行 | 6.25d | 1.5 周 |
| 阶段 4：文件与 AI | 5.25d | 1 周 |
| 阶段 5：报表与打磨 | 6.75d | 1.5 周 |
| **合计** | **32.75d** | **~7 周** |

---

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Shadcn/ui 组件定制复杂度 | 样式开发延期 | 优先使用默认样式，MVP 不做深度定制 |
| @dnd-kit 拖拽在嵌套树形结构中的兼容性 | 模块拖拽体验差 | 提前做 Spike 验证树形拖拽方案（阶段 2.2） |
| 分块上传后端接口未实现 | >10MB 文件无法上传 | 当前使用单次上传兜底，分块上传做前后端开关 |
| AI 轮询频率过高导致请求量过大 | 后端压力增加 | completed/failed 后立即停止轮询，设置最大轮询次数 |
| Recharts 包体积大 | 首屏加载变慢 | 路由级懒加载 + Vite manualChunks 分割 |
| 前后端接口字段不一致 | 联调效率低 | 每阶段交付后对照 Swagger 同步接口定义 |
| React Router v6 嵌套路由参数获取 | 开发困惑 | 统一使用 `useParams` + 类型断言工具 |
| TanStack Query v5 缓存失效策略 | 数据不一致 | 遵循 queryKeys 层级设计，mutation 精确 invalidation |
| 前后端阶段交付不对齐 | 前端验收阻塞 | 前端每个阶段明确标注 API 依赖，后端 plan 同步对齐。前端可先用 MSW mock 联调，后端就绪后切换真实 API |

---

**文档版本**：v1.1（review 修复）
**最后更新**：2026-05-18
