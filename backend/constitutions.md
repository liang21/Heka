# CLAUDE.md - aitestos 专家级操作手册

## 1. 指令优先级
1. **执行宪法**: 任何行动前必须阅读并对齐 `@constitution.md`。
2. **环境感知**: 修改前执行 `ls -R` 及 `go mod graph`。

## 2. Go 1.26 专家规范
* **工具链**: 强制使用 `go1.26+`。
    * 集合处理优先使用 `iter` 标准库。
    * 高并发场景利用 `sync` 包优化及新 `atomic` 类型。
* **代码风格**: 
    * 严格执行 `gofmt -s`。
    * 导包顺序：标准库、第三方库、内部包（分块排列）。

## 3. 微服务工程约束
* **架构隔离**: 
    * `internal/service/`: 纯净业务逻辑，不含传输协议。
    * `internal/transport/`: 负责 HTTP/GRPC 编解码与路由。
* **可观测性 (Observability)**:
    * **Tracing**: 入口函数必须透传 `context.Context` 以延续 TraceID。
    * **Logging**: 使用 `rs/zerolog` 结构化日志。禁止在循环内打日志，日志必须包含 `service_name` 和 `trace_id`。
    * **Metrics**: 关键路径必须预留 Prometheus 埋点（如：`issue_convert_total`）。
* **生命周期**: `main.go` 必须实现对 `SIGTERM` 的捕获，确保 Graceful Shutdown（优雅停机）。

## 4. 开发工作流
* **构建/调试**: 必须通过 `Makefile`。
    * `make test`: 运行全量测试。
    * `make web`: 启动服务。
    * `make dev`: 联动 OrbStack/docker-compose 启动依赖环境。
* **配置管理**: 严禁硬编码。修改配置结构必须同步更新 `config.example.yaml`。
* **Git 规范**: 严格执行 Conventional Commits (`type(scope): subject`)。

## 5. 目录架构与约束
- **原则**: 严禁在 `internal` 之外引用 `internal` 目录；严禁 `service` 直接依赖 `transport`。
- **/api**: 契约定义。修改任何接口前，先更新此处的 YAML。
- **/cmd**: 入口。只负责初始化环境并调用 `internal/app`。
- **/internal/domain**: 定义核心对象和接口。保持纯净，禁止引入第三方库。
- **/internal/ierrors**: 所有业务错误必须映射至此处的 Code。

## 6. 目录架构
```text
├── api/                # 接口契约定义
├── cmd/                # 程序入口 (含优雅停机逻辑)
├── internal/
│   ├── app/            # 依赖注入与生命周期编排 (Wire/Manual DI)
│   ├── config/         # 环境变量与 YAML 解析
│   ├── domain/         # 领域模型
│   ├── service/        # 核心业务
│   ├── transport/      # HTTP/GRPC 处理
│   ├── repository/         # 数据持久化 (DB/Cache 交互)
│   └── ierrors/        # 统一错误码处理方案
├── pkg/                    # 6. 公共工具 (慎用，仅限可被其他项目引用的逻辑)
├── scripts/                # 7. 脚本文件 (如数据库迁移指令)
├── tests/                  # 8. 跨包集成测试 (Integration Tests)
├── Makefile                # 9. 标准化指令入口
├── CLAUDE.md               # 10. Agent 运行手册
├── constitution.md         # 11. 项目宪法
├── go.mod
└── go.sum
```