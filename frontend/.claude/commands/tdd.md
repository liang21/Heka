---
description: 启动 TDD 红灯协议：精准定位任务书中的 Phase/Task，生成对应的失败测试
model: sonnet
argument-hint: [Phase_Title 或 Task_ID]
allowed-tools: [Read, Grep, Bash, Write, Edit]
---

# 执行协议：The Phase-Specific Red-Light Protocol (Frontend)

当收到 `/tdd [任务 ID]`（如 `/tdd T001`）或 `/tdd [Phase 名称]`（如 `/tdd Phase 1`）时：

**基于现有计划执行 TDD 流程：**
- 📋 计划文件：`specs/001-frontend-functionality/tasks.md`
- 📐 设计规范：`specs/001-frontend-functionality/spec.md`
- 📝 实现计划：`specs/001-frontend-functionality/plan.md`
- 🎨 前端规范：`specs/frontend-design-spec.md`

**执行模式：**
1. **单任务模式**：`/tdd T001` - 针对特定任务执行 TDD
2. **阶段模式**：`/tdd Phase 1` - 针对整个阶段的所有任务执行 TDD

**重要说明：**
- ✅ 计划已存在，直接基于 tasks.md 执行
- ✅ 不创建新计划，严格遵循现有任务定义
- ✅ 依赖检查基于任务的 `Depends` 字段
- ✅ 实现细节参考任务的 `Details` 字段

## ⚠️ 前置条件检查

**在执行任何操作前，必须先检查测试基础设施是否存在：**
1. 检查 `tests/` 目录是否存在
2. 检查 `vitest.config.ts` 是否存在
3. 检查 `package.json` 是否包含测试依赖（vitest、@testing-library/react、msw 等）

**如果测试基础设施不存在，必须先提示用户：**
```
⚠️ 检测到项目缺少测试基础设施。需要先安装以下依赖：

npm install -D vitest @vitest/ui @testing-library/react @testing-library/jest-dom @testing-library/user-event msw jsdom

然后创建基础配置文件：
- vitest.config.ts
- tests/msw/server.ts
- tests/setup.ts
```

**只有用户确认测试基础设施就绪后，才继续执行 TDD 流程。**

## 1. 精准定位 (Targeting)

- **扫描任务书**：读取 `@./specs/001-frontend-functionality/tasks.md`（如果文件不存在，报错并退出）。
- **匹配目标**：
  - 若输入为 Phase 名称（如 `Phase 1` 或 `阶段 1`），定位 `## Phase X: ...` 章节，提取该 Phase 下的所有任务。
  - 若输入为任务 ID（如 `T001`），直接定位到该任务。
- **任务分析**：
  - 提取任务的 `File` 字段（目标文件路径）
  - 提取任务的 `Details` 字段（实现细节）
  - 提取任务的 `Depends` 字段（前置依赖）
  - 确认任务类型：Create（新建文件）vs Modify（修改现有文件）

## 2. 计划关联 (Plan Mapping)

**直接读取现有计划文件，不创建新计划：**
- **主文件**: `@./specs/001-frontend-functionality/tasks.md` - 原子化任务列表
- **参考文件**: 
  - `@./specs/001-frontend-functionality/plan.md` - 实现计划
  - `@./specs/001-frontend-functionality/spec.md` - 规格文档
  - `@./specs/frontend-design-spec.md` - 前端设计规范

**交叉验证：**
- 确认目标 Feature 的目录路径（从 tasks.md 的 File 字段提取）
- 确认涉及的数据类型（从 src/types/ 的相关类型文件）
- 确认 API 端点（从 plan.md 的 API 服务层提取）
- 检查依赖的前置任务（Depends 字段）是否已完成（验证对应文件存在）

**计划已存在，直接执行 TDD 流程，无需创建新计划。**

## 3. 红灯阶段 (Red Phase) - 生成失败测试

- **依赖检查**：
  - 验证 `Depends` 字段列出的所有前置任务已完成
  - 检查对应的源文件是否存在（如 `src/types/api.ts`）
  - 如依赖未满足，报错并列出缺失的前置任务

- **测试文件路径**：
  - 基于任务的 `File` 字段生成对应的测试文件
  - 例如：`src/pages/Login/index.tsx` → `src/pages/Login/index.test.tsx`
  - 例如：`src/hooks/useAuth.ts` → `src/hooks/useAuth.test.ts`
  - 例如：`src/services/auth.ts` → `src/services/auth.test.ts`

- **原则**：只创建 `*.test.ts` / `*.test.tsx` 文件，禁止修改或创建业务代码。
- **代码生成规则**：

### 3a. Hook 测试（`*.test.ts`）

```typescript
// 模板结构（src/hooks/useXxx.test.ts）
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { server } from '@/tests/msw/server'

// Helper: 创建测试用 QueryClient（关闭 retry）
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={createTestQueryClient()}>
      {children}
    </QueryClientProvider>
  )
}

describe('useXxx', () => {
  it('should fetch data successfully', async () => {
    // Arrange: MSW handler
    server.use(
      http.get('/api/v1/xxx', () =>
        HttpResponse.json({ data: mockData })
      )
    )
    // Act
    const { result } = renderHook(() => useXxx(), { wrapper })
    // Assert
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(expected)
  })

  it('should handle error response', async () => {
    server.use(
      http.get('/api/v1/xxx', () =>
        HttpResponse.json({ error: 'Unauthorized' }, { status: 401 })
      )
    )
    const { result } = renderHook(() => useXxx(), { wrapper })
    await waitFor(() => expect(result.current.isError).toBe(true))
  })
})
```

### 3b. 组件测试（`*.test.tsx`）

```typescript
// 模板结构（src/pages/Xxx/index.test.tsx）
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { server } from '@/tests/msw/server'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{ui}</BrowserRouter>
    </QueryClientProvider>
  )
}

describe('XxxPage', () => {
  it('should render page heading', () => {
    renderWithProviders(<XxxPage />)
    expect(screen.getByRole('heading', { name: /xxx/i })).toBeInTheDocument()
  })

  it('should handle form submission', async () => {
    const user = userEvent.setup()
    server.use(
      http.post('/api/v1/xxx', () =>
        HttpResponse.json({ success: true })
      )
    )
    renderWithProviders(<XxxPage />)

    await user.type(screen.getByLabelText(/email/i), 'test@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /submit/i }))

    await waitFor(() => {
      expect(screen.getByText(/success/i)).toBeInTheDocument()
    })
  })
})
```

### 3c. Service 测试（`*.test.ts`）

```typescript
// 模板结构（src/services/xxx.test.ts）
import { describe, it, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/tests/msw/server'
import { xxxService } from './xxx'

describe('xxxService', () => {
  beforeEach(() => {
    // 每个 test 前重置 handlers
    server.resetHandlers()
  })

  it('should fetch items', async () => {
    server.use(
      http.get('/api/v1/xxx', () =>
        HttpResponse.json({ data: [{ id: 1, name: 'Test' }] })
      )
    )

    const result = await xxxService.getAll()
    expect(result).toEqual([{ id: 1, name: 'Test' }])
  })

  it('should handle network error', async () => {
    server.use(
      http.get('/api/v1/xxx', () =>
        HttpResponse.error()
      )
    )

    await expect(xxxService.getAll()).rejects.toThrow()
  })
})
```

### 3d. MSW Handler（如需要新建）

```typescript
// tests/msw/handlers/xxx.ts
import { http, HttpResponse } from 'msw'

export const xxxHandlers = [
  http.get('/api/v1/xxx', () =>
    HttpResponse.json({ data: [], total: 0, offset: 0, limit: 10 })
  ),
  http.post('/api/v1/xxx', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({ data: { ...body, id: 1 } }, { status: 201 })
  }),
]
```

- **验证**：运行 `npm test -- <test-file>` 并确认测试失败（编译错误如 `Cannot find module` 或断言失败如 `expected true to be false`）。

## 4. 汇报与中断 (Checkpoint)

- **任务映射**：列出目标任务的 `File` 字段 → 生成的测试文件路径
- **依赖状态**：确认所有 `Depends` 任务已完成
- **测试覆盖**：基于任务的 `Details` 字段，列出测试覆盖的核心功能点（1-3 个）
- **红灯验证**：展示 `npm test -- --run` 的失败输出摘要（确认红灯状态）
- **询问**："针对任务 [TXXX] 的红灯测试已就绪。是否开始功能代码实现以'转绿'？"

## 5. 绿灯阶段（用户确认后）

- **参考实现**：严格按照 tasks.md 中该任务的 `Details` 字段实现业务代码
- **文件创建/修改**：基于任务的 `File` 字段和 `Action` 字段（Create/Modify）
- **持续验证**：每完成一个实现步骤，运行对应测试确认通过
- **完成检查**：
  - 运行 `npm test -- --run <test-file>` 确认该测试通过
  - 检查代码是否符合 `@./specs/frontend-design-spec.md` 的规范
  - 验证 `Details` 字段中列出的所有验收要点

---

# 🔧 测试基础设施参考

**如果需要创建测试基础设施，参考以下配置：**

## vitest.config.ts
```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './tests/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

## tests/setup.ts
```typescript
import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

afterEach(() => {
  cleanup()
})
```

## tests/msw/server.ts
```typescript
import { setupServer } from 'msw/node'
import { HttpResponse, http } from 'msw'

export const server = setupServer(
  // 默认 handlers
  http.get('/api/v1/health', () => {
    return HttpResponse.json({ status: 'ok' })
  })
)

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())
```

---

# 📖 使用示例

## 示例 1：单任务 TDD 流程

```bash
# 1. 启动 TDD 流程
/tdd T001

# 2. AI 将执行：
#    - 读取 specs/001-frontend-functionality/tasks.md
#    - 定位 T001 任务（Vite + React + TypeScript 脚手架）
#    - 检查依赖（Depends: —）
#    - 生成测试文件：package.json.test.ts（验证依赖配置）
#    - 运行测试确认失败（红灯）

# 3. 用户确认后，AI 将：
#    - 按照 T001 的 Details 字段实现
#    - 运行测试确认通过（绿灯）
```

## 示例 2：阶段 TDD 流程

```bash
# 1. 启动整个阶段的 TDD 流程
/tdd Phase 1

# 2. AI 将执行：
#    - 读取 Phase 1 的所有任务（T001-TXXX）
#    - 按依赖顺序生成所有测试文件
#    - 逐任务确认红灯状态
#    - 等待用户确认后开始实现

# 3. 实现阶段：
#    - 按 Depends 顺序逐任务实现
#    - 每完成一个任务确认测试通过
#    - 最终运行 npm test 确认无回归
```

## 示例 3：基于现有计划的测试生成

```bash
# 任务 T008: 创建通用类型
# File: src/types/api.ts
# Depends: —

/tdd T008

# AI 将：
# 1. 读取 T008 的 Details 字段
# 2. 生成 src/types/api.test.ts
# 3. 测试 ApiResponse<T>, PaginatedData<T> 等类型定义
# 4. 确认测试失败（文件不存在）
# 5. 等待用户确认后创建 src/types/api.ts
# 6. 运行测试确认通过
```
