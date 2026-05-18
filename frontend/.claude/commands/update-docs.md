---
description: 为 frontend 项目生成完整的 README.md 文档
model: sonnet
allowed-tools: [Read, Write, Bash]
---

你是一位优秀的前端技术文档工程师。请为 Aitestos 前端项目生成一份完整、清晰、与当前代码完全同步的 `README.md` 文档。

## 执行步骤

请按以下顺序扫描并分析项目文件：

1. **核心配置文件**
   - 读取 `package.json` 获取：项目名称、技术栈依赖、npm scripts
   - 读取 `vite.config.ts` 了解：构建配置、插件设置
   - 读取 `tsconfig.json` 了解：TypeScript 配置

2. **开发规范**
   - 读取 `CLAUDE.md` 提取：代码风格、命名规范、目录架构、开发工作流

3. **项目结构**
   - 扫描 `src/` 目录生成：目录树结构、关键文件说明

4. **命令与配置**
   - 读取 `Makefile` 提取：所有可用的 make 命令及用途
   - 检查 `.env.example` 或 `.env` 提取：环境变量说明（如果存在）

## README 文档结构

请生成包含以下部分的 `README.md`：

```markdown
# Aitestos Frontend

## 项目简介

- 项目定位和目标
- 核心功能概述

## 技术栈

- 框架与库（从 package.json 提取，版本号精确到小版本）
- 开发工具链
- 代码质量工具

## 目录架构
```

src/
├── app/ # 应用入口
├── features/ # 业务功能模块
├── components/ # 共享组件
└── ...

```
（根据实际扫描结果生成）

## 快速开始

### 前置要求
- Node.js 版本要求
- 包管理器（pnpm/npm/yarn）

### 安装依赖
\`\`\`bash
pnpm install
\`\`\`

### 开发模式
\`\`\`bash
make dev
# 或
pnpm dev
\`\`\`

### 构建部署
\`\`\`bash
make build
# 或
pnpm build
\`\`\`

## 开发指南

### 代码规范
- 引用 [CLAUDE.md](./CLAUDE.md) 获取完整规范
- 关键要点提炼（3-5 条）

### Git 提交规范
\`\`\`
<type>(<scope>): <subject>
\`\`\`
- type: feat, fix, refactor, test, docs, chore
- scope: feature 目录名

### 测试
\`\`\`bash
make test          # 运行测试
make test-watch    # watch 模式
\`\`\`

### 代码检查
\`\`\`bash
make check         # lint + format + type-check
\`\`\`

## Makefile 命令参考
| 命令 | 说明 |
|------|------|
| make dev | 启动开发服务器 |
| make test | 运行单元测试 |
| ... | ... |

## 环境变量
| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| VITE_API_BASE_URL | API 基础路径 | /api |
| ... | ... | ... |

## 常见问题

### Q: 端口被占用怎么办？
A: ...

### Q: 如何切换 API 环境？
A: ...

## 许可证
Copyright © 2026 Aitestos
\`\`\`

## 输出要求

1. **格式规范**：使用标准 Markdown 格式，代码块指定语言
2. **版本精确**：技术栈版本号从 package.json 实际提取
3. **命令可用**：所有命令必须是真实可执行的
4. **结构清晰**：使用适当的标题层级和分隔
5. **链接有效**：内部链接使用相对路径

生成完成后，请将完整的 README.md 内容写入项目根目录。
```
