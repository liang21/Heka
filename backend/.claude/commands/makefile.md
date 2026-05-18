---
description: 首席架构师审计：构建生产级、支持 Go 1.24 特性且具备 O11y 元数据的 Makefile。
model: claude-3-5-sonnet
allowed-tools: [Read, Write, Bash]
---

# 角色设定
你是一位精通 **Go 1.24** 编译链、**GNU Make** 最佳实践以及 **Docker BuildKit** 优化的资深 DevOps 专家。

# 任务目标
深度分析项目结构（`@.`），为项目创建或重写 `Makefile`。这份文件必须是“工程化”的，而非简单的命令堆砌。

# 核心规范 (Hard Constraints):

1. **元数据注入 (Build-time LDFlags)**:
   - 必须自动提取 `git` 信息（Tag, Commit Hash, BuildTime）。
   - 通过 `-ldflags` 将这些变量注入到 `main` 包中，确保二进制文件支持 `--version` 查看元数据。

2. **Go 1.24 编译优化**:
   - `build`: 强制开启 `CGO_ENABLED=0` 以实现静态链接，适配 Distroless 镜像。
   - `test`: 默认开启 `-race` 检查，并支持通过变量指定包路径（如 `make test PKG=./internal/domain`）。
   - 增加 `tidy`: 运行 `go mod tidy` 和 `go mod verify` 确保依赖洁净。

3. **静态检查与质量 (Lint)**:
   - `lint`: 必须检查本地是否存在 `golangci-lint`，若无则给出安装指引。
   - 强制读取项目根目录的 `.golangci.yml` 配置。

4. **容器化集成 (Docker)**:
   - `docker-build`: 必须启用 `DOCKER_BUILDKIT=1`。
   - 镜像 Tag 统一为 `aitestos:latest`，并支持通过环境变量 `TAG` 覆盖。

5. **工程标准化**:
   - 必须使用 `.PHONY` 声明所有伪目标。
   - 必须包含一个 `help` 目标，利用 `sed` 或 `awk` 自动解析注释生成帮助文档。
   - 变量命名需清晰（如 `BINARY_NAME`, `GO_FILES`, `LDFLAGS`）。

# 输出要求:
1. **分析汇报**: 说明你如何识别入口文件（如 `cmd/server/main.go`）及项目版本逻辑。
2. **生成并写入**: 将优化后的 `Makefile` 覆盖写入项目根目录。
3. **自检说明**: 解释你加入的高级特性（如缓存清理、静态链接等）对项目的收益。

---
**工程师，请开始编写这份符合“首席架构师”标准的 Makefile。**