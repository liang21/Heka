---
description: 启动 TDD 红灯协议：根据 tasks.md 任务 ID 生成对应的失败测试。
argument-hint: [Phase_Number 或 Task_ID 或 “list”]
allowed-tools: [Read, Grep, Bash]
model: opus
---

# 执行协议：The Phase-Specific Red-Light Protocol

## 触发方式

```bash
/tdd [Phase_Number]          # 如: /tdd 1 (Phase 1: Foundation)
/tdd [Task_ID]               # 如: /tdd T006 (ID 类型)
/tdd list                    # 列出所有可用的 Phase 和 Task
```

## 执行流程

### 0. 列出可用任务 (List Mode)

当参数为 `list` 时：
1. 读取 `specs/001-backend-functionality/tasks.md`
2. 解析所有 Phase 和任务 ID
3. 输出格式化的任务清单，包含：
   - Phase 编号与标题
   - 任务 ID、文件路径、依赖关系
   - 并行标记 `[P]`

### 1. 精准定位 (Targeting)

根据参数类型执行：

**Phase 模式** (如 `/tdd 1`):
- 匹配 `## Phase [N]:` 标题
- 提取该 Phase 下所有任务
- 按依赖关系排序，优先处理无依赖的 `[P]` 任务

**Task ID 模式** (如 `/tdd T006`):
- 正则匹配 `#### T(\d+):` 提取任务
- 解析：File、Depends、Action、Details 字段
- 验证依赖任务是否已完成（检查文件是否存在）

### 2. 上下文解析 (Context Resolution)

1. **读取项目宪法**: `@constitutions.md` 第四条（测试务实）
2. **读取项目规范**: `@CLAUDE.md` DDD 架构规范
3. **确定测试策略**:
   - Domain 层：纯单元测试，禁止 Mock（宪法 4.3）
   - Repository 层：集成测试，使用 testcontainers
   - Application 层：Mock Repository
   - Interface 层：httptest

### 3. 路径映射表 (Path Mapping)

| Phase | 测试文件路径 | 说明 |
|-------|-------------|------|
| Phase 1 (Foundation) | 跳过 | 仅定义类型/接口，无需测试 |
| Phase 2 (Infrastructure) | `internal/infrastructure/persistence/postgres/*_test.go` | TDD: Repository 实现 |
| Phase 3 (Application) | `internal/application/*/service_test.go` | Mock Repository |
| Phase 3 (Interface) | `internal/interface/http/handler/*_test.go` | httptest |

### 4. 红灯阶段 (Red Phase) - 生成失败测试

**原则**:
- ✅ 只创建/修改 `*_test.go` 文件
- ❌ 禁止修改业务逻辑代码
- ✅ 测试必须体现验收标准 (AC)

**测试模板** (根据任务类型选择):

```go
// Domain 实体测试
func TestEntityName(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name    string
        input   InputType
        want    ExpectedType
        wantErr bool
    }{
        // Happy Path
        {
            name: “should create valid entity”,
            // ...
        },
        // Boundary Conditions
        {
            name:    “should fail on invalid input”,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test implementation
        })
    }
}
```

**Go 1.24 规范**:
- 使用 `t.Parallel()` 启用并行测试
- 使用 `tt := tt` 闭包捕获
- 使用 `testify/assert` 断言
- 遵循 `gofmt -s` 格式

### 5. 验证与汇报 (Verification & Checkpoint)

1. **运行测试**: `go test ./... -v`
2. **捕获预期失败**:
   - `undefined: XXX` - 接口/类型未定义 ✅ RED
   - `no such file` - 实现文件不存在 ✅ RED
   - `assertion failed` - 逻辑未实现 ✅ RED
3. **输出报告**:
   ```markdown
   ## 🔴 红灯测试已生成

   **任务**: T006 | ID 类型定义
   **文件**: `internal/domain/shared/types.go` (不存在)

   ### 测试覆盖
   - ✅ NewID() 生成有效 UUID
   - ✅ ParseID() 校验格式
   - ✅ IsEmpty() 边界检查

   ### 测试文件
   `internal/domain/shared/types_test.go` (已创建)

   ### 运行结果
   ```
  undefined: NewID
  ```
   ✅ 符合预期红灯状态

   **下一步**: 是否开始实现以转绿？
   ```

## 错误处理

| 场景 | 处理 |
|------|------|
| tasks.md 不存在 | 提示运行 `/makefile` 生成项目骨架 |
| Phase 不存在 | 列出可用 Phase，提示使用 `/tdd list` |
| Task ID 不存在 | 模糊搜索，提供建议 |
| 依赖未满足 | 列出依赖任务 ID，拒绝生成 |
| 目录不存在 | 自动创建目录结构 |

## 与项目宪法的对齐

参照 `constitutions.md` 第四条：
- **4.1 核心必测**: Domain 层纯业务逻辑必须有单元测试（表格驱动）
- **4.2 集成按需**: Repository 使用真实数据库连接
- **4.3 禁止 Mock 核心逻辑**: Domain 层测试禁止 Mock