---
description: 启动 TDD 红灯协议：精准定位任务书中的 Phase，生成对应的失败测试。
argument-hint: [Phase_Title_From_Tasks_MD]
allowed-tools: [Read, Grep, Bash]
---

# 执行协议：The Phase-Specific Red-Light Protocol

当收到 `/tdd [Phase 名称]` (如 `/tdd Phase 2: 领域模型 (P0)`) 时：

1. **精准定位 (Targeting)**：
   - **扫描任务书**：读取 `@./specs/001-backend-functionality/tasks.md`。
   - **正则匹配**：定位以 `## [Phase 名称]` 开头的 Markdown 章节。
   - **提取 AC**：抓取该标题下方直到下一个 `##` 标题前的所有任务点、验收标准 (Acceptance Criteria) 和 DDL 要求。

2. **计划关联 (Plan Mapping)**：
   - 同步读取 `@./specs/001-backend-functionality/plan.md`。
   - 交叉比对，确认该 Phase 对应的模块路径（如 `internal/domain` 或 `internal/repository`）。

3. **红灯阶段 (Red Phase) - [生成失败测试]**：
   - **原则**：只动 `*_test.go`，禁止修改业务逻辑。
   - **代码生成**：根据提取出的领域对象或接口定义，生成 **Table-Driven Tests**。
   - **Go 1.24 规范**：使用 `t.Parallel()`，并确保测试用例覆盖 Happy Path 及任务书中提到的边界条件。
   - **验证**：运行 `go test ./...` 并捕获预期的失败（如：`undefined` 错误或 `assertion failed`）。

4. **汇报与中断 (Checkpoint)**：
   - 展示针对该 Phase 提取的 1-2 个核心验收点。
   - 展示生成的测试代码 Skeleton。
   - **询问**：“针对 [Phase 名称] 的红灯测试已就绪。是否开始功能代码实现以‘转绿’？”