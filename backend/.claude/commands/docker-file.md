---
description: 首席架构师审计：构建生产级、零风险、极速缓存的 Go 1.24 容器镜像。
model: claude-3-5-sonnet
allowed-tools: [Read, Write, Bash]
---

# 角色设定
你是一位精通 **Go 1.24**、**Distroless 架构**与 **Linux 安全加固**的顶级 DevOps 专家。

# 任务目标
深度分析项目结构（优先查看 `@go.mod`, `@cmd/`），为该项目定制一份**工业级**的 `Dockerfile`。

# 核心规范 (Hard Constraints):

1. **构建加速 (Extreme Caching)**:
   - 使用 `mount=type=cache` 挂载 `/go/pkg/mod` 和 `/root/.cache/go-build`，确保利用宿主机构建缓存。
   - 严格遵循 `COPY go.mod go.sum ./` -> `RUN go mod download` 的顺序。

2. **多阶段构建 (Distroless Final)**:
   - **Stage 1 (Builder)**: 使用 `golang:1.24-bookworm` (或 Alpine)。
   - **编译参数**: 必须包含 `-installsuffix cgo -ldflags="-s -w -X main.version=$(git describe --tags) -X main.buildTime=$(date +%FT%T%z)"`。
   - **Stage 2 (Final)**: 使用 `gcr.io/distroless/static-debian12` (非 root 版本)。

3. **运行时安全 (Runtime Security)**:
   - **非 Root 强制**: 使用 `USER nonroot:nonroot`。
   - **只读文件系统友好**: 确保应用不尝试写入镜像内部目录。
   - **暴露端口**: 自动识别 `main.go` 中的端口定义并执行 `EXPOSE`。

4. **静态分析与优化**:
   - 检查是否需要 `tzdata` 或 `ca-certificates`（Distroless static 已包含，但在自定义镜像中需确认）。
   - 移除所有构建残留，镜像体积目标：< 20MB (对于标准 Go 二进制)。

# 输出要求:
1. **执行分析**: 简述你识别到的入口点、依赖项及特殊构建需求。
2. **生成并覆盖**: 将优化后的 `Dockerfile` 写入项目根目录。
3. **构建建议**: 提供一条包含 `docker buildx` 的推荐构建命令，以支持多架构打包。

---
**开始分析项目并执行容器化方案。**