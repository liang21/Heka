---
description: 深度分析 Git 变更，自动生成符合项目宪法与前端正工程规范的提交信息并执行提交
model: sonnet
allowed-tools: [Bash]
---

# 核心逻辑流程：

1. **获取变更上下文**：
   - 执行 `git diff --staged` 获取代码变更。
   - 如果 `staged` 为空，尝试执行 `git diff` 并提醒用户"尚未暂存变更"。

2. **多维度语义分析**（基于 React/TypeScript 专家视角）：
   - **Type 识别**：
     - 若包含 `*.test.ts` 或 `*.test.tsx` 的新增或修改：优先归类为 `test`。
     - 若包含 `src/features/*/hooks/` 下的逻辑重构：归类为 `refactor`。
     - 若修改了 `Makefile`、`CLAUDE.md`、`vitest.config.ts` 或 `package.json`（依赖变更）：归类为 `chore` 或 `build`。
     - 若仅修改 `*.css`、`theme.css`、`tailwind` 相关：归类为 `style`。
   - **Scope 提取**：自动提取受影响的 feature 模块名（例如：`src/features/auth/` -> `auth`）。
     - 跨模块变更时 scope 取 `shared` 或 `core`。
     - 修改 `components/business/` 时 scope 取 `components`。
     - 修改 `components/layout/` 时 scope 取 `layout`。
     - 修改 `src/types/` 时 scope 取 `types`。
     - 修改 `src/lib/` 时 scope 取 `lib`。
   - **Breaking Change 检测**：
     - 检查 TypeScript 接口（interface / type）的公共签名是否发生不兼容变更（删除字段、改类型、改必选）。
     - 检查 React 组件 Props 接口是否删除属性或改必选。
     - 检查枚举类型是否删除成员。
     - 若有，必须在 Footer 标记 `BREAKING CHANGE`。

3. **生成规范 Message**：
   - 严格遵循 `CLAUDE.md` 规范：`<type>(<scope>): <subject>`。
   - **Subject 要求**：使用祈使句，首字母小写，不加句号。例如：`feat(auth): add login page with form validation`。
   - **Body 要求**（可选）：若变更复杂，简述"为什么"这么改，而非"改了什么"。
   - **Footer**（可选）：
     - `BREAKING CHANGE: <description>`（如适用）
     - `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`

4. **交互确认与提交**：
   - 向用户展示生成的 Message 预览。
   - 确认后执行 `git commit -m "[Generated Message]"`。
