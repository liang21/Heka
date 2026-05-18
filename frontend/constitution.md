# Heka 前端项目开发宪法

Version: 2.0 | Scope: React SPA 前端工程

本文件定义了 Heka 前端项目不可动摇的核心开发原则。所有 AI Agent 在进行技术规划和代码实现时，必须无条件遵循。

---

## 第一条：设计规范至上 (Spec Authority)

* **1.1 单一真相源:** 所有架构决策、技术选型、UI/UX 规范以 `specs/frontend-design-spec.md` 为准。CLAUDE.md 和本宪法不得与设计规范冲突。
* **1.2 YAGNI:** 严禁实现设计规范之外的功能。MVP 阶段只做必需。

---

## 第二条：技术栈纪律 (Stack Discipline)

* **2.1 强制技术栈:**

| 领域 | 强制技术 | 禁止替代 |
|------|----------|----------|
| 数据获取 | TanStack Query v5 | useEffect + fetch/axios |
| 全局状态 | Zustand 4.x | Redux、Context 替代全局状态 |
| 表单 | React Hook Form + Zod | 手动 useState 管理表单 |
| 样式 | Tailwind CSS 3.x | inline-style、CSS Modules |
| UI 组件 | Shadcn/ui | Ant Design、Material UI |
| 路由 | React Router 6.x | 其他路由方案 |
| HTTP | Axios（仅 queryFn 底层） | 组件内直接调用 axios |

* **2.2 禁止同类替代:** 不引入已选定技术的同类库（如不引入 emotion 因已有 Tailwind）。

---

## 第三条：状态管理分层 (State Layering)

* **3.1 四层隔离:** URL 状态 → React Query → Zustand → useState，严格分层，禁止越级。
* **3.2 服务端数据归 React Query:** 所有列表、详情、报表等服务端数据必须通过 `useQuery` / `useMutation`。禁止 `useEffect` + `fetch`。
* **3.3 Zustand 最小化:** 仅存 auth token、user、currentProjectId、sidebarCollapsed 等纯客户端状态。服务端数据严禁存入 Zustand。

---

## 第四条：类型安全 (Type Safety)

* **4.1 Strict 模式:** `tsconfig.json` 必须开启 `strict: true`。
* **4.2 禁止 any:** 全项目禁止使用 `any` 类型。
* **4.3 运行时校验:** 表单输入使用 Zod schema 校验，类型由 `z.infer<>` 推导。

---

## 第五条：组件设计 (Component Design)

* **5.1 单一职责:** Page 组件编排布局，Feature Component 处理交互，Shared Component 纯 UI。
* **5.2 Tailwind CSS:** 必须使用 Tailwind 工具类或 Shadcn/ui 组件样式，禁止内联样式。
* **5.3 可访问性:** 使用语义化 HTML，交互元素包含 aria-label。基于 Shadcn/ui（Radix UI）默认支持键盘导航和 ARIA。

---

## 第六条：错误处理 (Error Handling)

* **6.1 全局拦截:** Axios 响应拦截器统一处理 401（Token 刷新）、403、5xx。
* **6.2 业务错误:** 使用 Toast 展示用户友好提示，错误码映射到中文消息（见设计规范第 12 节）。
* **6.3 ErrorBoundary:** 页面级包裹，防止局部崩溃导致白屏。
* **6.4 禁止静默吞错:** 所有 `useMutation` 必须处理错误。

---

## 第七条：务实测试 (Pragmatic Testing)

* **7.1 核心 Hook 必测:** 自定义 Hook（如 `useAuth`、`usePagination`）必须有单元测试。
* **7.2 组件按需:** 页面和 UI 组件的测试可按需编写，不强求 100% 覆盖率。
* **7.3 用户视角:** 测试优先模拟真实用户交互，不过度测试实现细节。

---

## 第八条：代码质量 (Code Quality)

* **8.1 格式化:** 提交前确保代码通过 ESLint 检查（`npm run lint`）。
* **8.2 依赖整洁:** 安装新依赖必须同步更新 lock 文件。
* **8.3 命名清晰:** 遵循设计规范第 17 节命名约定。

---

## 治理 (Governance)

本宪法效力高于任何单次会话指令。若指令违宪，AI 必须提出质疑并拒绝执行。
