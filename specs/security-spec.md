# Heka 安全规范

> 认证鉴权、数据安全、输入校验、AI 安全
> 版本：v1.0
> 日期：2025-05-15

---

## 1. 认证鉴权

### 1.1 JWT 认证机制

```go
// Token 结构
type Claims struct {
    UserID    shared.ID `json:"user_id"`
    Email     string    `json:"email"`
    IssuedAt  int64     `json:"iat"`
    ExpiresAt int64     `json:"exp"`
}

// Token 配置
const (
    AccessTokenTTL  = 24 * time.Hour
    RefreshTokenTTL = 7 * 24 * time.Hour
)
```

**规则**：
- Access Token 有效期 24h，Refresh Token 有效期 7 天
- 密码使用 bcrypt 加密（cost=12）
- Token 使用 HS256 签名，密钥从环境变量读取
- 登录失败不区分"用户不存在"和"密码错误"（防用户枚举）

### 1.2 RBAC 权限模型（预留）

MVP 阶段所有项目成员权限相同。预留角色扩展能力：

```go
// 预留角色定义（MVP 只使用 member）
type Role string

const (
    RoleOwner  Role = "owner"   // 项目创建者
    RoleAdmin  Role = "admin"   // 项目管理员
    RoleMember Role = "member"  // 普通成员（MVP 默认）
)

// 权限矩阵
// | 操作                | owner | admin | member |
// |---------------------|-------|-------|--------|
// | 删除项目             | ✅    | ❌    | ❌     |
// | 管理成员             | ✅    | ✅    | ❌     |
// | 创建/编辑/删除用例   | ✅    | ✅    | ✅     |
// | 执行测试             | ✅    | ✅    | ✅     |
// | AI 生成用例          | ✅    | ✅    | ✅     |
```

### 1.3 项目隔离

```go
// Middleware: 确保用户只能访问已加入项目的数据
func ProjectAccess(projectRepo project.Repository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, _ := sharedctx.UserIDFromContext(r.Context())
            projectID := shared.ID(r.URL.Query().Get("project_id"))

            if projectID != "" {
                isMember, err := projectRepo.IsMember(r.Context(), projectID, userID)
                if err != nil || !isMember {
                    response.Error(w, http.StatusForbidden, "AUTH-AU-003")
                    return
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 2. 数据安全

### 2.1 敏感数据存储

| 数据类型 | 存储方式 | 说明 |
|----------|----------|------|
| 用户密码 | bcrypt (cost=12) | 不可逆哈希 |
| JWT Secret | 环境变量 | 不落盘 |
| AI API Key | 环境变量 | 不落盘 |
| 数据库密码 | 环境变量 | 不落盘 |
| 用户邮箱 | 明文（需唯一约束） | 不脱敏 |

### 2.2 数据传输加密

- 生产环境强制 HTTPS
- 数据库连接使用 SSL（生产环境）
- Redis 密码认证
- Milvus 内网通信（不暴露外网）

### 2.3 SQL 注入防护

```go
// ✅ 参数化查询
db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", userID)

// ❌ 字符串拼接（严禁）
db.QueryRowContext(ctx, fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID))
```

**规则**：
- 所有 SQL 必须使用参数化查询（`$1`, `$2`, ...）
- GORM 默认参数化，但 Raw SQL 需手动确保
- 禁止拼接用户输入到 SQL 语句

---

## 3. 输入校验

### 3.1 服务端校验（必须）

```go
// 使用 validator 库进行结构体校验
type CreateTestCaseRequest struct {
    ProjectID   shared.ID  `validate:"required,uuid"`
    Title       string     `validate:"required,min=1,max=500"`
    Description string     `validate:"max=10000"`
    Steps       []StepDTO  `validate:"required,min=1,max=50"`
    Priority    int        `validate:"min=0,max=3"`
    Tags        []string   `validate:"max=10,dive,max=50"`
}
```

### 3.2 文件上传安全

```go
// 文件类型白名单
var allowedMIMETypes = map[string]bool{
    "application/pdf":                             true,
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
    "image/png":                                   true,
    "image/jpeg":                                  true,
}

// 文件大小限制
const maxUploadSize = 100 * 1024 * 1024 // 100MB

func validateUpload(file *multipart.FileHeader) error {
    if file.Size > maxUploadSize {
        return fmt.Errorf("FILE-VL-002: file size exceeds limit")
    }

    // 读取文件头判断真实类型（不依赖扩展名）
    f, _ := file.Open()
    defer f.Close()

    buffer := make([]byte, 512)
    f.Read(buffer)
    mimeType := http.DetectContentType(buffer)

    if !allowedMIMETypes[mimeType] {
        return fmt.Errorf("FILE-VL-001: unsupported file type")
    }
    return nil
}
```

### 3.3 UUID 格式校验

```go
// 所有 ID 参数必须验证 UUID 格式
func validateUUID(s string) error {
    _, err := uuid.Parse(s)
    if err != nil {
        return fmt.Errorf("SYS-VL-001: invalid UUID format")
    }
    return nil
}
```

---

## 4. AI 安全

### 4.1 Prompt Injection 防护

```go
// internal/infrastructure/ai/sanitize.go
package ai

// SanitizeInput 清洗用户输入，防止 Prompt Injection
func SanitizeInput(input string) string {
    // 1. 移除常见的 Prompt Injection 模式
    patterns := []string{
        "ignore previous instructions",
        "ignore all previous",
        "disregard all",
        "forget everything",
        "system:",
        "assistant:",
    }

    sanitized := input
    for _, pattern := range patterns {
        sanitized = strings.ReplaceAll(strings.ToLower(sanitized), pattern, "[filtered]")
    }

    return sanitized
}

// ValidateAIOutput 校验 AI 输出合法性
func ValidateAIOutput(output string) error {
    // 检查输出是否包含异常内容
    if strings.Contains(output, "```system") {
        return fmt.Errorf("AI-IE-002: AI output validation failed")
    }
    return nil
}
```

**规则**：
- 用户输入作为 Prompt 内容时，必须经过 `SanitizeInput` 清洗
- AI 输出必须经过 `ValidateAIOutput` 校验
- 用户文档内容（RAG 检索结果）作为上下文时，用明确的分隔符包裹

### 4.2 Prompt 结构化

```go
// 使用结构化 Prompt 模板，将用户输入和系统指令明确分离
const testGenerationPrompt = `You are a test case generation assistant.

Based on the following requirement document content, generate test cases.

<document_content>
%s
</document_content>

Requirements:
- Generate %d test cases
- Include negative test cases: %v
- Priority level: %s

Output format: JSON array of test cases.
`
```

---

## 5. CORS 和 CSRF

### 5.1 CORS 配置

```go
// 生产环境限制允许的域名
var allowedOrigins = []string{
    "https://heka.example.com",
    "http://localhost:3000", // 开发环境
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        if isAllowedOrigin(origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Max-Age", "86400")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### 5.2 CSRF 防护

- API 使用 Bearer Token 认证（非 Cookie），天然防 CSRF
- 确保 Cookie 设置 `SameSite=Strict`（如使用 Cookie）
- 关键操作（删除项目、移除成员）需要二次确认

---

## 6. 速率限制

```go
// 基于用户 ID 的速率限制
type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.Mutex
}

type visitor struct {
    tokens    float64
    lastCheck time.Time
}

// API 限流配置
var rateLimits = map[string]RateConfig{
    "api_global":     {Rate: 30, Burst: 50},        // 全局 30 req/s
    "ai_generate":    {Rate: 2, Burst: 5},          // AI 生成 2 req/min
    "file_upload":    {Rate: 5, Burst: 10},         // 文件上传 5 req/min
    "auth_login":     {Rate: 5, Burst: 10},         // 登录 5 req/min
}
```

---

## 7. 安全检查清单

部署前必须检查：

- [ ] 所有密码使用环境变量，未硬编码
- [ ] JWT Secret 足够长（>= 32 字符随机字符串）
- [ ] 生产环境启用 HTTPS
- [ ] 数据库不暴露外网端口
- [ ] Redis 设置密码
- [ ] MinIO 修改默认密码
- [ ] 文件上传限制类型和大小
- [ ] API 实现速率限制
- [ ] CORS 白名单配置正确
- [ ] AI 输入经过清洗
- [ ] 所有 SQL 参数化查询
- [ ] UUID 格式校验

---

**文档版本**：v1.0
**最后更新**：2025-05-15
