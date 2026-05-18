---
description: 前端专家级工程管理控制台（适配 React 19+、Vite、Docker BuildKit）
model: sonnet
allowed-tools: [Read, Write, Bash]
---

# ==============================================================================

# 前端专家级工程管理控制台

# 适配 React 19+, Vite, pnpm, Docker BuildKit

# ==============================================================================

.DEFAULT_GOAL := help

# --- 变量定义 ---

APP_NAME := aitest-frontend
VERSION := $(shell node -p "require('./package.json').version")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
IMAGE_TAG ?= latest

# 环境变量注入

VITE_METADATA := '{"version":"$(VERSION)","commit":"$(COMMIT_HASH)","buildTime":"$(BUILD_TIME)"}'

# --- 核心指令 ---

.PHONY: setup
setup: ## 初始化开发环境并安装依赖
@echo "🚀 Initializing environment..."
@command -v pnpm >/dev/null 2>&1 || { echo "❌ pnpm is required. Install via 'npm i -g pnpm'"; exit 1; }
pnpm install --frozen-lockfile

.PHONY: dev
dev: ## 启动 Vite 开发服务器
pnpm dev

.PHONY: tidy
tidy: ## 清理无用依赖并校验锁文件 (对应 Go Tidy)
pnpm prune
pnpm store prune

.PHONY: type-check
type-check: ## 静态类型检查 (宪法第三条)
pnpm exec tsc --noEmit

.PHONY: lint
lint: ## 执行 ESLint 与 Stylelint 检查 (符合 .golangci.yml 逻辑)
@echo "🔍 Linting code base..."
pnpm lint

.PHONY: test
test: ## 运行单元测试与集成测试 (Vitest/Testing Library)
@echo "🧪 Running tests with coverage..."
pnpm test -- --coverage $(if $(PKG),--dir $(PKG),)

.PHONY: test-e2e
test-e2e: ## 运行 Playwright E2E 测试
@echo "🎭 Running End-to-End tests..."
pnpm exec playwright test

.PHONY: build
build: tidy type-check ## 生成静态生产版本，注入元数据
@echo "🏗️ Building production bundle [$(VERSION)]..."
VITE_APP_METADATA=$(VITE_METADATA) pnpm build

.PHONY: docker-build
docker-build: ## 使用 BuildKit 构建高性能镜像
@echo "🐳 Building Docker image: $(APP_NAME):$(IMAGE_TAG)..."
DOCKER_BUILDKIT=1 docker build \
 --build-arg BUILD_METADATA=$(VITE_METADATA) \
		-t $(APP_NAME):$(IMAGE_TAG) .

.PHONY: clean
clean: ## 清理构建产物与缓存
rm -rf dist
rm -rf coverage
rm -rf .next
find . -name "node_modules" -type d -prune -exec rm -rf '{}' +

.PHONY: help
help: ## 显示此帮助信息
@grep -E '^[a-zA-Z_-]+:._?## ._$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.\*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
