---
description: 深度分析 Git 变更，自动生成符合项目宪法与 Go 工程规范的提交信息并执行提交。
allowed-tools: [bash]
---
# 核心逻辑流程：

1. **获取变更上下文**：
   - 执行 `git diff --staged` 获取代码变更。
   - 如果 `staged` 为空，尝试执行 `git diff` 并提醒用户“尚未暂存变更”。

2. **多维度语义分析**（基于 Go 专家视角）：
   - **Type 识别**：
     - 若包含 `_test.go` 的新增或修改：优先归类为 `test`。
     - 若包含 `internal/` 或 `cmd/` 下逻辑重构：归类为 `refactor`。
     - 若修改了 `Makefile`、`CLAUDE.md` 或 `go.mod`：归类为 `chore` 或 `build`。
   - **Scope 提取**：自动提取受影响的包名（例如：`internal/parser` -> `parser`）。
   - **Breaking Change 检测**：检查接口（interface）签名或公共 DDL 是否发生不兼容变更，若有，必须在 Footer 标记 `BREAKING CHANGE`。

3. **生成规范 Message**：
   - 严格遵循 `CLAUDE.md` 规范：`<type>(<scope>): <subject>`。
   - **Subject 要求**：使用祈使句，首字母小写，不加句号。例如：`feat(figma): add parser for vector nodes`。
   - **Body 要求**（可选）：若变更复杂，简述“为什么”这么改，而非“改了什么”。

4. **交互确认与提交**：
   - 向用户展示生成的 Message 预览。
   - 确认后执行 `git commit -m "[Generated Message]"`。