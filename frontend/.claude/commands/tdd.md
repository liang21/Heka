---
description: 启动 TDD 红灯协议：精准定位任务书中的 Phase/Task，生成对应的失败测试
model: sonnet
argument-hint: [Phase_Title 或 Task_ID]
allowed-tools: [Read, Grep, Bash, Write, Edit]
---

# 执行协议：The Phase-Specific Red-Light Protocol (Frontend)

当收到 `/tdd [Phase 名称或任务 ID]`（如 `/tdd Phase 1: 认证模块` 或 `/tdd T19`）时：

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
  - 若输入为 Phase 名称（如 `Phase 1`），定位 `## Phase X: ...` 章节，提取该 Phase 下所有测试任务（编号为奇数的 T 任务，如 T13, T15, T17...）。
  - 若输入为任务 ID（如 `T19`），直接定位到该任务，检查是否为测试任务（描述中含"测试"）。若为实现任务，自动回退到其前置测试任务（T19-1=T18 的下一测试任务）。
- **提取 AC**：抓取每个测试任务的 AC（验收标准）——即"X 个测试用例通过"之前列出的具体测试点。

## 2. 计划关联 (Plan Mapping)

- 同步读取 `@./specs/001-frontend-functionality/plan.md` 和 `@./specs/001-frontend-functionality/spec.md`。
- 交叉比对：
  - 确认目标 Feature 的目录路径（如 `src/features/auth/`）。
  - 确认涉及的数据类型（从 `src/types/api.ts` 和 `src/types/enums.ts`）。
  - 确认 API 端点（从 `specs/openapi.yaml`）。
- 确认依赖的前置任务是否已完成（检查对应文件是否存在）。

## 3. 红灯阶段 (Red Phase) - 生成失败测试

- **原则**：只创建 `*.test.ts` / `*.test.tsx` 文件，禁止修改或创建业务代码。
- **代码生成规则**：

### 3a. Hook 测试（`*.test.ts`）

```typescript
// 模板结构
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
  it('should ...', async () => {
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
})
```

### 3b. 组件测试（`*.test.tsx`）

```typescript
// 模板结构
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
  it('should render ...', () => {
    renderWithProviders(<XxxPage />)
    expect(screen.getByRole('heading', { name: /xxx/i })).toBeInTheDocument()
  })

  it('should handle user interaction', async () => {
    const user = userEvent.setup()
    renderWithProviders(<XxxPage />)
    await user.click(screen.getByRole('button', { name: /submit/i }))
    // assertion...
  })
})
```

### 3c. MSW Handler（如需要新建）

```typescript
// tests/msw/handlers/xxx.ts
import { http, HttpResponse } from 'msw'

export const xxxHandlers = [
  http.get('/api/v1/xxx', () =>
    HttpResponse.json({ data: [], total: 0, offset: 0, limit: 10 })
  ),
]
```

- **验证**：运行 `npm test -- <test-file>` 并确认测试失败（编译错误如 `Cannot find module` 或断言失败如 `expected true to be false`）。

## 4. 汇报与中断 (Checkpoint)

- 列出该 Phase/Task 下所有已生成的测试文件路径。
- 展示测试覆盖的核心验收点（1-3 个关键 AC）。
- 展示 `npm test -- --run` 的失败输出摘要（确认红灯状态）。
- **询问**："针对 [Phase/Task] 的红灯测试已就绪。是否开始功能代码实现以'转绿'？"

## 5. 绿灯阶段（用户确认后）

- 按 tasks.md 中实现任务的顺序，逐个创建/修改业务代码。
- 每完成一个实现任务，运行对应测试确认通过。
- 全部转绿后运行 `npm test` 确认无回归。

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
