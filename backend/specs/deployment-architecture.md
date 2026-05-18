# Heka 部署架构设计

> 当前阶段（MVP）单机部署架构
> 版本：v1.0
> 日期：2025-05-15

---

## 1. 架构概览

### 1.1 部署模式

**单机部署**，所有服务运行在同一台服务器上。

**资源要求**：
- CPU：4 核心以上
- 内存：8GB 以上
- 磁盘：100GB 以上 SSD

### 1.1.1 内存预算分配（8GB 总内存）

| 服务 | 预算上限 | 说明 |
|------|---------|------|
| PostgreSQL | 2.0 GB | shared_buffers=2GB, effective_cache_size=6GB（含 OS 缓存） |
| Milvus | 2.0 GB | cache.size=2GB, 包含 etcd/MinIO 开销 |
| Redis | 0.5 GB | appendonly + 缓存数据 + 任务队列 |
| Go Backend | 1.0 GB | 应用内存 + AI Worker（5 Worker × ~100MB） |
| Nginx + etcd + MinIO | 0.5 GB | 反向代理 + Milvus 依赖 |
| **OS + 预留** | **2.0 GB** | 操作系统 + 文件系统缓存 + 缓冲 |

**AI Worker 约束**：
- MVP 阶段：5 Worker（并发 AI 调用上限 5）
- 队列长度：50（防止内存溢出）
- 单次 AI 调用内存峰值：~100MB（含 Prompt 构建 + 响应解析）

> **注意**：以上为 8GB 基础配置。如果实际 AI Worker 需求超过 5，建议将内存升级到 16GB。

### 1.2 架构图

```
                        ┌─────────────────────────────────────────┐
                        │             用户浏览器                 │
                        └─────────────────────────────────────────┘
                                           │
                                           ▼
                        ┌─────────────────────────────────────────┐
                        │    Nginx (可选) :80 / :443              │
                        │    反向代理 + SSL 终止 + 静态文件       │
                        └─────────────────────────────────────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    ▼                      ▼                      ▼
        ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
        │  Frontend :3000   │  │  Backend :8080    │  │   文件上传路径    │
        │  React 静态文件    │  │  Go API 服务      │  │   /var/heka/uploads│
        └───────────────────┘  └───────────────────┘  └───────────────────┘
                                         │
                    ┌──────────────────────┼──────────────────────┐
                    ▼                      ▼                      ▼
        ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
        │  PostgreSQL :5432 │  │    Redis :6379    │  │   Milvus :19530   │
        │   主数据存储      │  │    缓存 + 队列    │  │   向量存储        │
        └───────────────────┘  └───────────────────┘  └───────────────────┘
                                                              │
                                        ┌─────────────────────┼─────────────────────┐
                                        ▼                     ▼                     ▼
                            ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
                            │     etcd :2379    │  │    Minio :9000    │  │   AI Provider     │
                            │   Milvus 元数据   │  │   Milvus 存储     │  │  Claude/OpenAI/   │
                            └───────────────────┘  └───────────────────┘  │   Gemini/Ollama   │
                                                                     └───────────────────┘
```

---

## 2. 组件清单

### 2.1 核心组件

| 组件 | 技术栈 | 端口 | 职责 | 必需 |
|------|--------|------|------|------|
| **Frontend** | React 18 + TS + Tailwind | 3000 | 用户界面，静态文件服务 | ✅ |
| **Backend** | Go 1.21+ + GORM | 8080 | API 服务，业务逻辑 | ✅ |
| **PostgreSQL** | PostgreSQL 15+ | 5432 | 主数据存储 | ✅ |
| **Redis** | Redis 7+ | 6379 | 缓存 + 任务队列 | ✅ |
| **Milvus** | Milvus 2.3+ | 19530 | 向量存储（RAG） | ✅ |

### 2.2 Milvus 依赖组件

| 组件 | 端口 | 职责 | 必需 |
|------|------|------|------|
| **etcd** | 2379 | Milvus 元数据存储 | ✅ |
| **MinIO** | 9000 | Milvus 对象存储 | ✅ |

### 2.3 外部依赖

| 组件 | 职责 | 必需 |
|------|------|------|
| **AI Provider** | LLM API 调用（Claude/OpenAI/Gemini/Ollama） | ✅ |
| **Nginx** (可选) | 反向代理、SSL 终止、静态文件服务 | ❌ |

---

## 3. 请求流转路径

### 3.1 用户访问前端

```
用户浏览器
    │
    ▼
Nginx (可选) :80 / :443
    │ 静态资源请求
    ▼
Frontend :3000 (React 静态文件)
    │ 返回 HTML/JS/CSS
    ▼
用户浏览器渲染
```

### 3.2 API 请求

```
用户浏览器
    │
    ▼
Frontend :3000
    │ API 调用 (/api/*)
    ▼
Nginx (可选)
    │ 反向代理
    ▼
Backend :8080
    │ 验证 JWT Token
    ▼
Handler (Interface 层)
    │ 调用应用服务
    ▼
Service (Application 层)
    │ 业务逻辑编排
    ▼
┌──────────┬──────────┬──────────┐
▼          ▼          ▼          ▼
PostgreSQL  Redis    Milvus   AI Provider
(主数据)   (缓存)   (向量)   (LLM)
```

### 3.3 AI 用例生成流程

```
用户点击"AI 生成"
    │
    ▼
Frontend → Backend API
    │
    ▼
AI Application Service
    │ 1. 获取文件
    ▼
File Repository → PostgreSQL
    │ 2. 检查向量索引
    ▼
RAG Repository → Milvus
    │ 3. 检索相关内容
    ▼
Milvus Vector Search
    │ 4. 构建 Prompt
    ▼
AI Client (Multi-Provider)
    │ 5. 调用 LLM (带重试、熔断)
    ▼
AI Provider (Claude/OpenAI/Gemini)
    │ 6. 返回生成结果
    ▼
Backend 解析并创建用例
    │ 7. 批量写入
    ▼
PostgreSQL
    │ 8. 返回结果
    ▼
Frontend 显示
```

### 3.4 文件上传流程

```
用户选择文件
    │
    ▼
Frontend → Backend API
    │ multipart/form-data
    ▼
Backend :8080
    │ 1. 验证文件类型/大小
    ▼
File Application Service
    │ 2. 保存到本地存储
    ▼
LocalStorage → /var/heka/uploads
    │ 3. 写入文件记录
    ▼
PostgreSQL (files 表)
    │ 4. 异步解析/分块
    ▼
后台 Worker (Redis Queue)
    │ 5. 文件解析
    ▼
File Parser (PDF/Word/Excel/Image)
    │ 6. 文本分块
    ▼
Chunker (语义重叠分块)
    │ 7. 生成向量
    ▼
Embedding Service → AI Provider
    │ 8. 存储向量
    ▼
Milvus (document_chunks 集合)
    │ 9. 更新索引状态
    ▼
PostgreSQL (files 表: is_indexed=true)
    │ 10. 返回成功
    ▼
Frontend 显示
```

---

## 4. 组件职责详解

### 4.1 Frontend (React 应用)

**职责**：
- 用户界面渲染
- 客户端路由
- API 调用封装
- 状态管理（Zustand）
- 数据缓存（React Query）

**关键页面**：
- 登录页 `/login`
- 项目列表 `/`
- 测试用例 `/project/:id/testcases`
- 测试计划 `/project/:id/plans`
- 文件管理 `/project/:id/files`
- AI 生成 `/project/:id/ai-generate`

**部署方式**：
```bash
# 构建
npm run build

# 输出静态文件到 dist/ 目录
# 由 Nginx 或 Backend 直接服务
```

---

### 4.2 Backend (Go API 服务)

**职责**：
- HTTP API 服务
- JWT 认证
- 业务逻辑编排
- 数据访问协调
- AI 调用管理
- 文件处理

**分层结构**：
```
Interface 层 (HTTP Handler + Middleware)
    ↓
Application 层 (Service 编排)
    ↓
Domain 层 (Entity + Repository 接口)
    ↓
Infrastructure 层 (PostgreSQL/Redis/Milvus 实现)
```

**关键 API 端点**：
- `POST /api/auth/login` - 用户登录
- `GET /api/testcases` - 查询用例
- `POST /api/testcases` - 创建用例
- `POST /api/ai/generate-testcases` - AI 生成
- `POST /api/files/upload` - 文件上传

**配置管理**：
```yaml
server:
  port: 8080
database:
  host: postgres
  port: 5432
redis:
  host: redis
  port: 6379
milvus:
  host: milvus
  port: 19530
ai:
  providers:
    claude:
      api_key: ${CLAUDE_API_KEY}
      enabled: true
```

---

### 4.3 PostgreSQL

**职责**：
- 用户数据
- 项目/模块/用例/计划/执行记录
- 文件元数据
- 文档块元数据
- 向量嵌入记录

**关键表**：
```sql
-- 核心业务表
users
projects
modules
test_cases
test_steps
test_plans
test_executions
execution_results

-- 文件和 RAG
files
file_versions
document_chunks
vector_embeddings
```

**数据持久化**：
```yaml
volumes:
  - postgres-data:/var/lib/postgresql/data
```

**连接池配置**：
```go
max_open_conns: 100
max_idle_conns: 10
conn_max_lifetime: 1h
```

---

### 4.4 Redis

**职责**：
1. **缓存**：
   - 用户信息 (1h)
   - 项目信息 (30min)
   - 用例列表 (5min)
   - 模块树 (10min)
   - 统计数据 (10min)

2. **任务队列**：
   - 文件解析任务
   - 向量索引任务
   - AI 生成任务

**缓存策略**：
```go
// Cache-Aside 模式
func (s *Service) GetTestCase(ctx context.Context, id ID) (*TestCase, error) {
    // 1. 先查缓存
    cached, err := s.cache.Get(ctx, fmt.Sprintf("testcase:%s", id))
    if err == nil {
        return cached, nil
    }

    // 2. 缓存未命中，查数据库
    tc, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. 写入缓存
    s.cache.Set(ctx, fmt.Sprintf("testcase:%s", id), tc, 10*time.Minute)
    return tc, nil
}
```

**数据结构**：
```
# 缓存
user:{id} → Hash
project:{id} → Hash
project:{id}:modules → List
testcase:{id} → Hash
ai:task:{id} → Hash

# 队列
file:parse:queue → List
rag:index:queue → List
ai:generate:queue → List
```

---

### 4.5 Milvus (向量数据库)

**职责**：
- 存储文档向量嵌入
- 语义相似度搜索
- RAG 检索

**Collection 设计**：
```python
collection_name = "document_chunks"

fields = [
    {"name": "id", "type": "string", "primary_key": True},
    {"name": "chunk_id", "type": "string"},
    {"name": "file_id", "type": "string"},
    {"name": "content", "type": "string"},
    {"name": "embedding", "type": "float_vector", "dim": 1536},
    {"name": "index", "type": "int64"},
    {"name": "created_at", "type": "int64"}
]

index = {
    "field_name": "embedding",
    "index_type": "IVF_FLAT",
    "metric_type": "L2",
    "params": {"nlist": 128}
}
```

**检索流程**：
```go
// 1. 用户查询 → 向量化
queryEmbedding := aiClient.Embed(ctx, userQuery)

// 2. Milvus 相似度搜索
results, err := milvusClient.Search(
    ctx,
    collectionName,
    [][]float32{queryEmbedding},
    "embedding",
    []string{"content", "file_id"},
    topK=5,
)

// 3. 返回相关文档块
for _, result := range results {
    chunks = append(chunks, Chunk{
        Content: result.Fields["content"].(string),
        Score:   result.Score,
    })
}
```

---

### 4.6 Milvus 依赖组件

#### etcd
- **端口**：2379
- **职责**：Milvus 元数据存储
- **数据**：Collection schema、索引定义

#### MinIO
- **端口**：9000
- **职责**：Milvus 对象存储
- **数据**：向量数据文件、索引文件

---

### 4.7 AI Provider

**职责**：
- LLM API 调用
- 文本生成（测试用例）
- 代码分析（变更影响）
- Embedding 生成（向量化）

**多 Provider 故障转移**：
```
请求
  ↓
Claude (Priority 1)
  ↓ (失败)
OpenAI (Priority 2)
  ↓ (失败)
Gemini (Priority 3)
  ↓ (失败)
Ollama (Priority 4, 本地)
```

**容错机制**：
- 熔断器（连续失败自动熔断）
- 指数退避重试（1s → 2s → 4s）
- 超时控制（60s）
- Worker Pool（并发控制）

---

## 5. 端口分配

| 组件 | 内部端口 | 外部暴露 | 用途 |
|------|---------|---------|------|
| Nginx | 80, 443 | ✅ | HTTP/HTTPS |
| Frontend | 3000 | ❌ | React 开发服务器/静态文件 |
| Backend | 8080 | ❌ | Go API 服务 |
| PostgreSQL | 5432 | ❌ | 数据库 |
| Redis | 6379 | ❌ | 缓存/队列 |
| Milvus | 19530 | ❌ | 向量数据库 |
| etcd | 2379 | ❌ | Milvus 元数据 |
| MinIO | 9000 | 9000 (控制台) | Milvus 存储 |

**网络隔离**：
- 所有服务在同一 Docker Network 中
- 只有 Nginx 暴露到外网
- 服务间通过内部 DNS 通信

---

## 6. 数据流向

### 6.1 读流程

```
用户请求
    ↓
Nginx (可选)
    ↓
Backend
    ↓
Redis (缓存层)
    ↓ [未命中]
PostgreSQL (数据层)
    ↓
返回数据
```

### 6.2 写流程

```
用户请求
    ↓
Backend
    ↓
PostgreSQL (写主数据)
    ↓
Redis (删除缓存)
    ↓
返回结果
```

### 6.3 AI 处理流程

```
用户请求
    ↓
Backend
    ↓
Redis (任务队列)
    ↓
后台 Worker
    ↓
PostgreSQL (读文件)
    ↓
Milvus (向量检索)
    ↓
AI Provider (LLM 调用)
    ↓
PostgreSQL (写结果)
    ↓
Redis (更新缓存)
    ↓
返回结果
```

---

## 7. 部署配置

### 7.1 Docker Compose 配置

```yaml
version: '3.8'

services:
  # PostgreSQL
  postgres:
    image: postgres:15
    container_name: heka-postgres
    environment:
      POSTGRES_DB: heka
      POSTGRES_USER: heka
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - heka-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U heka"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis
  redis:
    image: redis:7-alpine
    container_name: heka-redis
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    networks:
      - heka-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Milvus 依赖
  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    container_name: heka-etcd
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
    volumes:
      - etcd-data:/etcd
    networks:
      - heka-network
    restart: unless-stopped

  minio:
    image: minio/minio:RELEASE.2023-03-20T20-16-18Z
    container_name: heka-minio
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    volumes:
      - minio-data:/minio_data
    command: minio server /minio_data
    networks:
      - heka-network
    restart: unless-stopped

  # Milvus
  milvus:
    image: milvusdb/milvus:v2.3.0
    container_name: heka-milvus
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    volumes:
      - milvus-data:/var/lib/milvus
    ports:
      - "19530:19530"
    depends_on:
      - etcd
      - minio
    networks:
      - heka-network
    restart: unless-stopped

  # Backend
  heka-backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: heka-backend
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_NAME=heka
      - DATABASE_USER=heka
      - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - MILVUS_HOST=milvus
      - MILVUS_PORT=19530
      - JWT_SECRET=${JWT_SECRET}
      - AI_CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - AI_OPENAI_API_KEY=${OPENAI_API_KEY}
    volumes:
      - heka-uploads:/app/uploads
      - heka-logs:/app/logs
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - milvus
    networks:
      - heka-network
    restart: unless-stopped

  # Frontend
  heka-frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: heka-frontend
    ports:
      - "3000:80"
    depends_on:
      - heka-backend
    networks:
      - heka-network
    restart: unless-stopped

  # Nginx (可选)
  nginx:
    image: nginx:alpine
    container_name: heka-nginx
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - heka-backend
      - heka-frontend
    networks:
      - heka-network
    restart: unless-stopped

volumes:
  postgres-data:
  redis-data:
  milvus-data:
  etcd-data:
  minio-data:
  heka-uploads:
  heka-logs:

networks:
  heka-network:
    driver: bridge
```

### 7.2 环境变量

```bash
# .env
POSTGRES_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret

# AI Provider API Keys
CLAUDE_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GEMINI_API_KEY=xxx

# 文件上传
MAX_UPLOAD_SIZE=104857600  # 100MB
UPLOAD_PATH=/app/uploads
```

### 7.3 生产级 Nginx 配置

```nginx
# /etc/nginx/conf.d/heka.conf
upstream heka_backend {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name heka.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name heka.example.com;

    # SSL 配置
    ssl_certificate     /etc/letsencrypt/live/heka.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/heka.example.com/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 前端静态资源
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2)$ {
            proxy_pass http://127.0.0.1:3000;
            expires 7d;
            add_header Cache-Control "public, immutable";
        }
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://heka_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE 支持（AI 进度推送）
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding on;

        # WebSocket 支持
        proxy_set_header Upgrade $http_upgrade;

        # 超时配置（AI 调用可能较长）
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;

        # 文件上传大小限制
        client_max_body_size 100M;

        # 限流
        limit_req zone=api burst=20 nodelay;
    }

    # 健康检查端点（不限流）
    location /api/v1/health {
        proxy_pass http://heka_backend;
        access_log off;
    }
}

# 限流区域定义（在 http 块中）
# limit_req_zone $binary_remote_addr zone=api:10m rate=30r/s;
```

### 7.4 SSL 证书自动化

```bash
# 安装 certbot
apt-get install certbot python3-certbot-nginx

# 获取证书
certbot --nginx -d heka.example.com

# 自动续期（已由 certbot timer 管理）
certbot renew --dry-run
```

---

## 8. 健康检查

### 8.1 健康检查端点

```
GET /api/health
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "status": "healthy",
    "services": {
      "database": "ok",
      "redis": "ok",
      "milvus": "ok"
    },
    "version": "1.0.0"
  }
}
```

### 8.2 监控指标

```
GET /api/monitoring/ai
```

**响应**：
```json
{
  "code": 0,
  "data": {
    "providers": {
      "claude": {
        "state": "closed",
        "requests": 150,
        "successes": 145,
        "failures": 5,
        "success_rate": 0.967
      }
    },
    "pool": {
      "queue_length": 2,
      "max_workers": 10,
      "active_tasks": 8
    }
  }
}
```

---

## 9. 启动顺序

1. **基础设施**：
   ```
   etcd → minio → milvus
   postgres
   redis
   ```

2. **应用服务**：
   ```
   backend → frontend
   ```

3. **反向代理**（可选）：
   ```
   nginx
   ```

**Docker Compose 自动处理依赖顺序**：
```yaml
depends_on:
  - postgres
  - redis
  - milvus
```

---

## 10. 备份策略

### 10.1 数据备份

**PostgreSQL**：
```bash
# 每日备份
docker exec heka-postgres pg_dump -U heka heka > backup_$(date +%Y%m%d).sql

# 恢复
docker exec -i heka-postgres psql -U heka heka < backup_20250515.sql
```

**Redis**：
```bash
# RDB 自动备份（appendonly yes）
# 文件位置：redis-data:/data/dump.rdb
```

**Milvus**：
```bash
# 备份集合数据
# 需要使用 Milvus Backup 工具
```

### 10.2 文件备份

```bash
# 上传文件
tar -czf uploads_backup_$(date +%Y%m%d).tar.gz /var/heka/uploads
```

---

## 11. 监控告警

### 11.1 Prometheus 指标采集

```yaml
# 在 docker-compose.yml 中增加 Prometheus 服务
prometheus:
  image: prom/prometheus:latest
  container_name: heka-prometheus
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus-data:/prometheus
  ports:
    - "9090:9090"
  networks:
    - heka-network
```

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'heka-backend'
    static_configs:
      - targets: ['heka-backend:8080']
    metrics_path: /metrics
```

### 11.2 关键告警规则

| 指标 | 阈值 | 级别 | 说明 |
|------|------|------|------|
| PostgreSQL 连接数 | > 80 | Warning | 连接池即将耗尽 |
| API P95 延迟 | > 2s | Warning | 接口响应变慢 |
| API P99 延迟 | > 5s | Critical | 严重影响用户体验 |
| AI 调用失败率 | > 20% | Critical | AI 服务异常 |
| 磁盘使用率 | > 80% | Warning | 需要清理或扩容 |
| 内存使用率 | > 85% | Warning | 可能 OOM |

### 11.3 日志聚合

MVP 阶段使用 Docker 日志驱动：

```bash
# 查看所有服务日志
docker compose logs -f --tail=100

# 查看特定服务
docker logs heka-backend -f --since=1h

# 导出日志
docker compose logs --no-color > heka-logs-$(date +%Y%m%d).txt
```

生产环境可集成 Loki + Grafana 进行日志聚合和查询。

---

## 12. 性能优化建议

### 12.1 数据库优化

```sql
-- 索引优化
CREATE INDEX CONCURRENTLY idx_test_cases_project_status 
    ON test_cases(project_id, status);

CREATE INDEX CONCURRENTLY idx_executions_plan_status 
    ON test_executions(plan_id, status);

-- 全文搜索
CREATE INDEX idx_test_cases_fulltext 
    ON test_cases USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));
```

### 12.2 缓存优化

```go
// 预热缓存
func (s *Service) WarmupCache(ctx context.Context) error {
    // 预加载热门项目
    projects, _ := s.projectRepo.FindPopular(ctx, 10)
    for _, p := range projects {
        s.cache.Set(ctx, fmt.Sprintf("project:%s", p.ID), p, 30*time.Minute)
    }
    return nil
}
```

### 12.3 连接池配置

```go
// PostgreSQL
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(1 * time.Hour)

// Redis
pool := &redis.Pool{
    MaxIdle:     10,
    MaxActive:   100,
    IdleTimeout: 5 * time.Minute,
}
```

---

## 13. 故障排查

### 13.1 常见问题

**问题 1：Milvus 连接失败**
```bash
# 检查 Milvus 状态
docker logs heka-milvus

# 检查 etcd 和 minio
docker logs heka-etcd
docker logs heka-minio
```

**问题 2：AI 调用超时**
```bash
# 检查网络连接
curl -I https://api.anthropic.com

# 查看日志
docker logs heka-backend | grep "AI"
```

**问题 3：文件上传失败**
```bash
# 检查磁盘空间
df -h

# 检查上传目录权限
ls -la /var/heka/uploads
```

### 13.2 日志查看

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务
docker logs heka-backend -f --tail 100
```

---

**文档版本**：v1.0
**最后更新**：2025-05-15
