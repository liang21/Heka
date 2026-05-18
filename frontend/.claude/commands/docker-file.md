---
description: 首席架构师审计：构建生产级、零风险、极速缓存的 React 容器镜像
model: sonnet
allowed-tools: [Read, Write, Bash]
---

# 角色设定

你是一位精通 **React 19**、**Vite 8**、**Nginx** 与 **前端容器化**的顶级 DevOps 专家。

# 任务目标

深度分析项目结构（优先查看 `@package.json`, `@vite.config.ts`），为该项目定制一份**工业级**的 `Dockerfile`。

# 核心规范 (Hard Constraints):

## 1. 构建加速 (Extreme Caching)

- **Yarn 缓存挂载**: 使用 `mount=type=cache,target=/usr/local/share/.cache/yarn` 确保利用宿主机依赖缓存。
- **分层 COPY 策略**: 严格遵循 `COPY package.json yarn.lock ./` → `RUN yarn install --frozen-lockfile` → `COPY . .` 顺序。
- **构建产物复用**: 将 `node_modules` 缓存在单独层中，源码变更不触发全量重装。

## 2. 多阶段构建 (Distroless Final)

### Stage 1 (Builder)

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
# 依赖安装（缓存挂载）
COPY package.json yarn.lock ./
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn \
    yarn install --frozen-lockfile
# 复制源码并构建
COPY . .
RUN yarn build
```

### Stage 2 (Final)

- **基础镜像**: `nginx:1.27-alpine`（轻量级，< 10MB）
- **静态文件提取**: 从 `builder` 阶段复制 `dist/` 到 nginx html 目录
- **Nginx 配置**: 内嵌 `nginx.conf`，配置 SPA 路由回退、Gzip 压缩、缓存头

## 3. 运行时安全 (Runtime Security)

- **非 Root 强制**: 使用 `USER nginx`（alpine nginx 镜像已内置 nginx 用户）
- **只读根文件系统**: nginx 运行不依赖写入镜像内部，可配合 `--read-only` 标志使用
- **暴露端口**: 自动识别并执行 `EXPOSE 8080`（或根据 vite.config.server.port 配置）
- **健康检查**: 添加 `HEALTHCHECK --interval=30s --timeout=3s CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1`

## 4. Nginx 配置优化

生成的 `nginx.conf` 应包含：

```nginx
user nginx;
worker_processes auto;
pid /run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;

    server {
        listen 8080;
        server_name _;
        root /usr/share/nginx/html;
        index index.html;

        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # SPA 路由回退
        location / {
            try_files $uri $uri/ /index.html;
        }

        # 安全头
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
    }
}
```

## 5. 环境变量注入

构建时支持注入版本元数据（从 Makefile 继承）：

- `VITE_APP_VERSION`: 应用版本号
- `VITE_APP_COMMIT`: Git commit hash
- `VITE_APP_BUILD_TIME`: 构建时间戳

在 Dockerfile 中通过 `ARG` 和 `ENV` 传递：

```dockerfile
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ENV VITE_APP_VERSION=${VERSION}
ENV VITE_APP_COMMIT=${COMMIT}
ENV VITE_APP_BUILD_TIME=${BUILD_TIME}
```

# 输出要求:

1. **执行分析**: 简述你识别到的入口点、依赖项、构建命令及端口配置。
2. **生成 Dockerfile**: 将优化后的 `Dockerfile` 写入项目根目录。
3. **生成 nginx.conf**: 将 Nginx 配置写入项目根目录（如不存在）。
4. **构建建议**: 提供一条包含 `docker buildx` 的推荐构建命令，以支持多架构打包（linux/amd64, linux/arm64）。

# 示例构建命令:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=$(node -p "require('./package.json').version") \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t aitestos/frontend:latest \
  -f Dockerfile \
  .
```

---

**开始分析项目并执行容器化方案。**
