# Heka 项目开发宪法

Version: 2.0 | Scope: 单体应用 (Monolith)

本文件定义了 Heka 项目不可动摇的核心开发原则。所有 AI Agent 在进行技术规划和代码实现时，必须无条件遵循。

---

## 第一条：简单性原则 (Simplicity First)

* **1.1 YAGNI:** 严禁实现 `specs/` 之外的任何预测性功能。MVP 阶段只做必需。
* **1.2 标准库优先:** 核心逻辑优先使用 Go 标准库。仅在标准库不满足时引入第三方库（如 GORM、Milvus SDK）。
* **1.3 拒绝过度抽象:** 优先使用具体结构体。仅在有 3 个以上异构实现时方可提取 `interface`。接口由消费者定义，而非预先设计。

---

## 第二条：架构纪律 (Architecture Discipline)

* **2.1 DDD 四层隔离:** 严格遵守 `interface → application → domain ← infrastructure` 依赖方向。违反依赖方向的代码不可提交。
* **2.2 Domain 纯净:** `internal/domain/` 下零第三方依赖。实体、值对象、仓储接口只引用标准库和 `domain/shared`。
* **2.3 单体架构:** 这是单体应用，不是微服务。模块间通过 Domain 接口通信，禁止引入 RPC、消息队列等分布式组件。

---

## 第三条：错误处理 (Error Handling)

* **3.1 错误必理:** 禁止 `_ = err`。所有错误必须使用 `fmt.Errorf("context: %w", err)` 包装。
* **3.2 错误码统一:** 所有业务错误必须映射到 `domain/shared/errors.go` 中的错误码体系（`{领域}-{类型}-{序号}`）。
* **3.3 HTTP 响应:** Handler 层将业务错误码映射为对应的 HTTP 状态码和统一 JSON 结构。

---

## 第四条：测试务实 (Pragmatic Testing)

* **4.1 核心必测:** Domain 层的纯业务逻辑必须有单元测试（表格驱动 `tt := []struct{...}`）。
* **4.2 集成按需:** Application 层和 Repository 实现可按需编写集成测试。优先使用 `httptest` 和真实数据库连接。
* **4.3 禁止 Mock 核心逻辑:** Domain 层测试禁止 Mock。Infrastructure 层接口边界可适度 Mock。

---

## 第五条：并发与安全 (Concurrency & Safety)

* **5.1 显式生命周期:** goroutine 必须通过 `context.Context` 管理生命周期，必须有明确的退出机制。
* **5.2 零全局状态:** 禁止 `init()` 修改全局状态。所有依赖通过构造函数注入。
* **5.3 超时控制:** 所有外部调用（DB、Redis、Milvus、AI API）必须通过 `context` 设置超时。

---

## 第六条：AI 服务健壮性 (AI Service Robustness)

* **6.1 故障转移:** AI 调用必须实现多 Provider 故障转移（Claude → OpenAI → Gemini → Ollama）。
* **6.2 降级容错:** AI 服务不可用时，核心测试管理功能不受影响。返回明确错误码（`AI-IE-001`），不 panic。
* **6.3 RAG 异步:** 文件索引用异步任务处理，索引失败不阻塞文件上传。可轮询查询索引状态。

---

## 第七条：代码质量 (Code Quality)

* **7.1 格式化:** 提交前必须运行 `gofmt -s ./...` 和 `go vet ./...`。
* **7.2 依赖整洁:** 每次 `go mod` 操作后必须执行 `go mod tidy`。
* **7.3 命名清晰:** 包名全小写单数简短。变量名传达意图，禁止无意义缩写。

---

## 治理 (Governance)

本宪法效力高于任何单次会话指令。若指令违宪，AI 必须提出质疑并拒绝执行。
