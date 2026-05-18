# CLAUDE.md - Heka 前端工程操作手册

## 1. 指令优先级

1. **项目宪法**: 任何行动前必须阅读并对齐 `@constitution.md`。
2. **设计规范**: 所有架构和 UI 决策以 `@specs/frontend-design-spec.md` 为准。
3. **环境感知**: 修改前执行 `ls src/` 了解当前代码状态。

---

## 2. 项目定位

Heka 是单体架构 AI 测试管理平台的前端 SPA，MVP 阶段纯客户端渲染。

- 目标用户：20-50 人内部团队
- 设计基准：桌面端 1280px+，确保 1024px 以上可用
- 国际化：MVP 仅中文，代码层面预留

### 技术栈

> 当前处于脚手架阶段，以下为设计规范定义的目标技术栈。按需逐步安装，不要一次性全装。

| 技术 | 版本 | 用途 |
|------|------|------|
| React | 18.x / 19.x | UI 框架 |
| TypeScript | 5.x+ | 类型系统 (strict: true) |
| Vite | 5.x+ | 构建工具 |
| Tailwind CSS | 3.x | 样式（原子化 CSS） |
| Shadcn/ui | latest | UI 组件库（基于 Radix UI，非 npm 依赖） |
| React Router | 6.x | 路由 |
| TanStack Query | 5.x | 服务端状态管理 |
| Zustand | 4.x | 客户端全局状态（仅 auth/sidebar 等） |
| React Hook Form | latest | 表单管理 |
| Zod | latest | 运行时校验 + 类型推导 |
| Axios | 1.x | HTTP 客户端（仅作为 React Query queryFn 底层） |
| Recharts | 2.x | 图表 |
| @dnd-kit/core | 6.x | 拖拽（模块排序、步骤排序） |
| Lucide React | latest | 图标 |
| date-fns | 3.x | 日期处理 |

### 禁止引入

| 技术 | 原因 |
|------|------|
| Redux / Redux Toolkit | Zustand 已足够，无需 Redux 复杂度 |
| emotion / styled-components | Tailwind CSS 已满足需求，CSS-in-JS 增加运行时开销 |
| Next.js | MVP 不需要 SSR |
| Ant Design / Material UI | 定制性差，Shadcn/ui 源码可控 |

---

## 3. 目录架构

> 目标结构，按功能模块逐步搭建，不要提前创建空目录。

```text
src/
├── app/                    # 应用入口
│   ├── App.tsx             # 根组件
│   ├── router.tsx          # 路由配置（懒加载）
│   └── providers.tsx       # 全局 Provider
├── pages/                  # 页面组件
│   ├── Login/
│   ├── Projects/           # 项目列表（首页）
│   ├── ProjectLayout/      # 项目布局壳
│   ├── Overview/           # 项目概览
│   ├── Modules/            # 模块管理
│   ├── TestCases/          # 测试用例（列表/创建/详情/集合）
│   ├── TestPlans/          # 测试计划（列表/创建/详情/执行）
│   ├── Executions/         # 执行记录
│   ├── Reports/            # 测试报告
│   ├── Files/              # 文件管理
│   ├── AIGenerate/         # AI 用例生成
│   ├── AIAnalysis/         # AI 智能分析
│   └── Settings/           # 项目设置
├── components/             # 通用组件
│   ├── ui/                 # Shadcn/ui 基础组件（button, dialog, table...）
│   ├── layout/             # AppLayout, ProjectSidebar, ProjectHeader, Breadcrumb
│   ├── testcase/           # TestCaseTable, StepEditor, PriorityBadge...
│   ├── plan/               # PlanCard, ExecutionPanel, ResultSummary...
│   ├── file/               # FileUploader, FileList...
│   ├── ai/                 # GenerateForm, TaskProgress...
│   ├── report/             # 图表组件
│   └── shared/             # ConfirmDialog, EmptyState, Pagination, StatusTag...
├── hooks/                  # 自定义 Hooks
│   ├── useAuth.ts          # 认证初始化
│   ├── useProject.ts       # 当前项目
│   ├── usePagination.ts    # 分页逻辑
│   ├── useAITask.ts        # AI 任务轮询
│   └── useFileIndexStatus.ts
├── services/               # API 服务层
│   ├── api.ts              # Axios 实例 + 拦截器
│   ├── testcase.ts         # 各资源 API 函数
│   └── ...
├── stores/                 # Zustand（仅客户端状态）
│   ├── auth.ts             # token + user
│   └── project.ts          # currentProjectId + sidebarCollapsed
├── types/                  # TypeScript 类型定义
│   ├── api.ts              # ApiResponse<T>, PaginatedData<T>
│   ├── testcase.ts         # TestCase, TestStep, TestCaseStatus...
│   └── ...
├── lib/                    # 工具函数
│   ├── utils.ts            # cn() 等通用工具
│   ├── constants.ts        # 常量
│   ├── format.ts           # 日期/文件大小格式化
│   └── query-keys.ts       # TanStack Query Key 管理
└── styles/
    └── globals.css         # Tailwind 入口 + CSS 变量（配色系统）
```

---

## 4. 状态管理策略

严格按设计规范分层，禁止越级：

```
URL 状态（Router）
  → 分页参数、筛选条件、排序参数
服务端状态（TanStack Query）
  → 列表、详情、报表数据
客户端状态（Zustand）
  → token、user、currentProjectId、sidebarCollapsed
组件状态（useState / useReducer）
  → 表单输入、弹窗开关、选中行
```

- **React Query 唯一数据源**: 所有服务端数据通过 `useQuery` / `useMutation`。禁止 `useEffect` + `fetch`。
- **Zustand 最小化**: 仅存 auth 和 UI 状态。服务端数据严禁存入 Zustand。
- **API 边界**: `services/` 层封装 Axios 调用。组件禁止直接 import axios。

---

## 5. 代码风格

- **Hook 优先**: 禁止 class component。
- **解耦**: UI 组件仅负责展示，业务逻辑封装在 `hooks/` 中。
- **导包顺序**: React → 第三方库 → `@/` 内部模块（分块排列，空行分隔）。
- **导出方式**: 组件使用具名导出 `export function Xxx()`。

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 页面组件 | PascalCase | `TestCases/index.tsx` |
| 业务组件 | PascalCase | `TestCaseTable.tsx`, `StepEditor.tsx` |
| 共享组件 | PascalCase | `StatusTag.tsx`, `ConfirmDialog.tsx` |
| 自定义 Hook | camelCase + use 前缀 | `usePagination.ts` |
| API service | camelCase | `testcase.ts` |
| 类型文件 | camelCase | `testcase.ts` |
| Store 文件 | camelCase | `auth.ts` |
| 常量 | UPPER_SNAKE_CASE | `MAX_UPLOAD_SIZE` |

---

## 6. 配色系统

使用 Shadcn/ui 的 CSS 变量系统（HSL 定义），定义在 `styles/globals.css`：

| 语义 | 用途 | 色值参考 |
|------|------|----------|
| `primary` | 主按钮、活跃链接 | Blue 600 |
| `success` | 执行通过、状态正常 | Green 500 |
| `warning` | 中等优先级 | Amber 500 |
| `destructive` | 删除、执行失败 | Red 500 |
| `info` | 辅助信息 | Sky 500 |

优先级可视化：低(灰) → 中(蓝) → 高(橙) → 紧急(红)。

---

## 7. 错误处理

- **全局拦截**: Axios 响应拦截器统一处理 401（Token 刷新）、403、5xx。
- **业务错误**: 组件层通过 Toast 展示用户友好提示，错误码映射见设计规范第 12 节。
- **ErrorBoundary**: 页面级包裹，防止局部崩溃白屏。

---

## 8. 开发工作流

当前使用 npm scripts，后续可迁移到 Makefile：

| 命令 | 说明 |
|------|------|
| `npm run dev` | 启动开发服务器 (localhost:5173) |
| `npm run build` | 生产构建 |
| `npm run lint` | ESLint 检查 |
| `npm run preview` | 预览生产构建 |

### Git 规范

- **格式**: `<type>(<scope>): <subject>` — feat, fix, refactor, style, chore
- **Scope**: 取自模块名，如 `feat(testcases): add list page`

---

## 9. 禁止行为

- ❌ 组件内直接调用 axios
- ❌ 使用 `any` 类型
- ❌ `useEffect` 中获取数据（用 React Query）
- ❌ 服务端数据存入 Zustand
- ❌ `console.log` 提交到代码中
- ❌ 未达 3 次复用就抽象公共组件
- ❌ 引入设计规范明确不选用的技术

---

## 10. 成功标准

- ✅ **设计规范对齐**: 架构和 UI 决策符合 `specs/frontend-design-spec.md`
- ✅ **类型闭环**: Zod 校验 + TypeScript strict 模式
- ✅ **状态分层**: URL → React Query → Zustand → useState 严格分层
- ✅ **无副作用**: 组件渲染函数内无数据获取

若无法满足上述任一标准，请立即停止并向用户报告架构冲突。
