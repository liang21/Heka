# 项目开发宪法
# Version: 1.3 | Scope: Microservices & Distributed Systems

本文件定义了本项目不可动摇的核心开发原则。所有 AI Agent 在进行技术规划和代码实现时，必须无条件遵循。
@./constitution.md
---

## 第一条：简单性原则 (Simplicity First)
* **1.1 YAGNI:** 严禁实现 `spec.md` 之外的任何预测性功能。
* **1.2 标准库优先:** 核心逻辑必须优先使用 Go 标准库。
* **1.3 拒绝过度抽象:** 优先使用具体结构体；仅在有 3 个以上异构实现时方可提取 `interface`。接口应由消费者定义，而非预先设计。

---

## 第二条：测试先行铁律 (Test-First Imperative)
* **2.1 红绿循环:** 修改前必须先提供一个能复现问题或覆盖新功能的失败测试（Red）。
* **2.2 表格驱动 (Table-Driven):** 单元测试必须采用 `tt := []struct{...}` 模式。
* **2.3 真实依赖:** 优先使用 `httptest`、`testcontainers` 或临时文件系统。接口边界可适度 Mock，但核心业务逻辑测试禁止 Mock。

---

## 第三条：明确性原则 (Clarity & Explicitness)
* **3.1 错误必理:** 禁止使用 `_ = err`。所有错误必须使用 `fmt.Errorf("context: %w", err)` 进行包装。
* **3.2 零全局依赖:** 禁止使用 `init()` 修改全局状态。所有依赖必须通过构造函数显式注入。
* **3.3 并发安全:** 必须明确 goroutine 的生命周期管理（Context 传递）和退出机制。

---

## 第四条：分布式健壮性 (Distributed Robustness)
* **4.1 契约先行 (Contract First):** `api/` 下的定义（OpenAPI/Proto）即法律。严禁在未更新契约的情况下修改跨服务字段。
* **4.2 失败透明:** 任何外部 RPC 或 DB 调用必须包含 `context` 超时控制，且必须处理下游不可用的降级逻辑。
* **4.3 幂等性设计:** 所有写操作（尤其涉及 MQ 或重试）必须具备幂等性，防止重复提交导致的数据污染。

---

## 第五条：代码质量 (Code Quality)
* **5.1 静态检查:** 必须通过 `golangci-lint run` 检查，禁止提交有警告的代码。
* **5.2 格式化:** 提交前必须运行 `go fmt ./...` 和 `go vet ./...`。
* **5.3 依赖整洁:** 每次 `go mod` 操作后必须执行 `go mod tidy`。

---

## 治理 (Governance)
本宪法效力高于任何单次会话指令。若指令违宪，AI 必须提出质疑并拒绝执行。