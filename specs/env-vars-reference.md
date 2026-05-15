# Heka 环境变量参考

> 统一所有环境变量命名、默认值和说明
> 版本：v1.0
> 日期：2025-05-15

---

## 1. 命名规范

- 全大写，下划线分隔：`SECTION_VARIABLE_NAME`
- 使用 `HEKA_` 前缀区分服务环境变量
- 敏感信息（密码、密钥）**必须**通过环境变量注入
- 非敏感配置可通过 `config.yaml` 或环境变量设置

---

## 2. 环境变量清单

### 2.1 数据库

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_DB_HOST` | ✅ | `localhost` | PostgreSQL 主机 |
| `HEKA_DB_PORT` | ✅ | `5432` | PostgreSQL 端口 |
| `HEKA_DB_NAME` | ✅ | `heka` | 数据库名 |
| `HEKA_DB_USER` | ✅ | `heka` | 数据库用户 |
| `HEKA_DB_PASSWORD` | ✅ | - | 数据库密码（敏感） |
| `HEKA_DB_MAX_OPEN_CONNS` | ❌ | `50` | 最大开放连接数 |
| `HEKA_DB_MAX_IDLE_CONNS` | ❌ | `10` | 最大空闲连接数 |
| `HEKA_DB_CONN_MAX_LIFETIME` | ❌ | `1h` | 连接最大存活时间 |

### 2.2 Redis

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_REDIS_HOST` | ✅ | `localhost` | Redis 主机 |
| `HEKA_REDIS_PORT` | ✅ | `6379` | Redis 端口 |
| `HEKA_REDIS_PASSWORD` | ❌ | `""` | Redis 密码（敏感） |
| `HEKA_REDIS_DB` | ❌ | `0` | Redis 数据库编号 |

### 2.3 Milvus

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_MILVUS_HOST` | ✅ | `localhost` | Milvus 主机 |
| `HEKA_MILVUS_PORT` | ✅ | `19530` | Milvus 端口 |
| `HEKA_MILVUS_COLLECTION` | ❌ | `heka_chunks` | Collection 名称 |
| `HEKA_EMBEDDING_DIMENSION` | ❌ | `1536` | 向量维度（随模型变化） |

### 2.4 JWT 认证

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_JWT_SECRET` | ✅ | - | JWT 签名密钥（敏感，>= 32 字符） |
| `HEKA_JWT_TTL` | ❌ | `24h` | Access Token 有效期 |

### 2.5 AI Provider

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_AI_CLAUDE_API_KEY` | ❌ | - | Claude API Key（敏感） |
| `HEKA_AI_OPENAI_API_KEY` | ❌ | - | OpenAI API Key（敏感） |
| `HEKA_AI_GEMINI_API_KEY` | ❌ | - | Gemini API Key（敏感） |
| `HEKA_AI_OLLAMA_HOST` | ❌ | `http://localhost:11434` | Ollama 本地服务地址 |
| `HEKA_AI_POOL_WORKERS` | ❌ | `5` | AI Worker 并发数 |
| `HEKA_AI_POOL_QUEUE_SIZE` | ❌ | `50` | AI 任务队列长度 |
| `HEKA_AI_TIMEOUT_REQUEST` | ❌ | `60s` | AI 请求超时 |
| `HEKA_AI_RETRY_MAX_ATTEMPTS` | ❌ | `3` | AI 调用重试次数 |

### 2.6 文件上传

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_UPLOAD_PATH` | ❌ | `/var/heka/uploads` | 文件存储路径 |
| `HEKA_UPLOAD_MAX_SIZE` | ❌ | `104857600` | 最大文件大小（100MB） |

### 2.7 Figma 集成

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_FIGMA_ACCESS_TOKEN` | ❌ | - | Figma API Token（敏感） |

### 2.8 服务配置

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `HEKA_SERVER_PORT` | ❌ | `8080` | HTTP 服务端口 |
| `HEKA_SERVER_MODE` | ❌ | `release` | 运行模式（debug/release） |
| `HEKA_LOG_LEVEL` | ❌ | `info` | 日志级别（debug/info/warn/error） |

---

## 3. Docker Compose 兼容映射

为兼容 `docker-compose.yml` 中的旧命名，Backend 容器同时支持以下环境变量：

| Docker Compose 变量 | 映射到 |
|---------------------|--------|
| `DATABASE_HOST` | `HEKA_DB_HOST` |
| `DATABASE_PORT` | `HEKA_DB_PORT` |
| `DATABASE_NAME` | `HEKA_DB_NAME` |
| `DATABASE_USER` | `HEKA_DB_USER` |
| `DATABASE_PASSWORD` | `HEKA_DB_PASSWORD` |
| `REDIS_HOST` | `HEKA_REDIS_HOST` |
| `REDIS_PORT` | `HEKA_REDIS_PORT` |
| `MILVUS_HOST` | `HEKA_MILVUS_HOST` |
| `MILVUS_PORT` | `HEKA_MILVUS_PORT` |
| `JWT_SECRET` | `HEKA_JWT_SECRET` |
| `POSTGRES_PASSWORD` | PostgreSQL 容器使用（非 Backend） |
| `CLAUDE_API_KEY` | `HEKA_AI_CLAUDE_API_KEY` |
| `OPENAI_API_KEY` | `HEKA_AI_OPENAI_API_KEY` |

---

## 4. .env 模板

```bash
# ===== Heka 环境变量 =====
# 复制此文件为 .env 并填入实际值

# 数据库
HEKA_DB_HOST=localhost
HEKA_DB_PORT=5432
HEKA_DB_NAME=heka
HEKA_DB_USER=heka
HEKA_DB_PASSWORD=your_secure_password

# Redis
HEKA_REDIS_HOST=localhost
HEKA_REDIS_PORT=6379
HEKA_REDIS_PASSWORD=

# Milvus
HEKA_MILVUS_HOST=localhost
HEKA_MILVUS_PORT=19530

# JWT
HEKA_JWT_SECRET=your_jwt_secret_min_32_chars_

# AI Provider（按需启用）
HEKA_AI_CLAUDE_API_KEY=
HEKA_AI_OPENAI_API_KEY=
HEKA_AI_GEMINI_API_KEY=
HEKA_AI_OLLAMA_HOST=http://localhost:11434

# 文件上传
HEKA_UPLOAD_PATH=/var/heka/uploads
HEKA_UPLOAD_MAX_SIZE=104857600

# Figma（可选）
HEKA_FIGMA_ACCESS_TOKEN=

# 服务
HEKA_SERVER_PORT=8080
HEKA_SERVER_MODE=release
HEKA_LOG_LEVEL=info
```

---

## 5. 其他文档引用说明

以下文档中的环境变量命名应以此文档为准：

- `deployment-architecture.md` 中的 `.env` 示例
- `heka-design-doc.md` 中的配置项说明
- `database-performance-spec.md` 中的连接池配置

---

**文档版本**：v1.0
**最后更新**：2025-05-15
