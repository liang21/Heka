---
description: 启动 TDD 红灯协议：根据 tasks.md 任务 ID 生成对应的失败测试。
argument-hint: [Phase_Number 或 Task_ID 或 “list”]
allowed-tools: [Read, Grep, Bash]
model: opus
---

# 执行协议：The Phase-Specific Red-Light Protocol

## 项目文件关联

本协议严格遵循以下项目规范文件：
- **主规格**: `specs/001-backend-functionality/spec.md` — 后端功能规格 v1.2
- **任务列表**: `specs/001-backend-functionality/tasks.md` — 175 个原子化任务
- **项目计划**: `specs/001-backend-functionality/plan.md` — 高层实施计划 v1.1
- **项目宪法**: `constitutions.md` — 不可动摇的开发原则 v2.0
- **操作手册**: `CLAUDE.md` — DDD 架构与编码规范

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

**检查现有计划**: 执行前先检查是否存在对应 Phase 的活跃计划文件（`.claude/plans/*.md`）。如存在，直接使用该计划，不创建新计划。

根据参数类型执行：

**Phase 模式** (如 `/tdd 1`):
- 匹配 `tasks.md` 中 `## Phase [N]:` 标题
- 提取该 Phase 下所有任务
- 按 `Depends` 字段排序，优先处理无依赖的 `[P]` 任务
- 引用 `spec.md` 相关章节获取验收标准

**Task ID 模式** (如 `/tdd T006`):
- 正则匹配 `#### T(\d+):` 提取任务
- 解析：File、Depends、Action、Details 字段
- 验证依赖任务是否已完成（检查文件是否存在）
- 关联 `spec.md` 中的对应验收标准

### 2. 上下文解析 (Context Resolution)

**读取顺序**:
1. **项目宪法**: `constitutions.md` 第四条（测试务实）
2. **项目规范**: `CLAUDE.md` DDD 架构规范
3. **功能规格**: `specs/001-backend-functionality/spec.md` 相关章节
4. **任务详情**: `specs/001-backend-functionality/tasks.md` 对应任务

**测试策略**:
- Domain 层：纯单元测试，禁止 Mock（宪法 4.3）
- Repository 层：集成测试，使用 testcontainers
- Application 层：Mock Repository
- Interface 层：httptest

### 3. 路径映射表 (Path Mapping)

根据 `tasks.md` 的 Phase 划分：

| Phase | tasks.md 范围 | 测试文件路径 | TDD 类型 |
|-------|--------------|-------------|----------|
| Phase 1 (Foundation) | T001-T038 | 跳过* | 仅类型/接口定义 |
| Phase 2 (Infrastructure) | T039-T099 | `internal/infrastructure/**/*_test.go` | 集成测试 (testcontainers) |
| Phase 3 (Application) | T100-T125 | `internal/application/*/service_test.go` | Mock Repository |
| Phase 3 (Interface) | T126-T148 | `internal/interface/http/handler/*_test.go` | httptest |
| Phase 4 (Integration) | T149-T158 | `tests/e2e/*_test.go`, `tests/migration/*_test.go` | 端到端测试 |

> *Phase 1 例外: T012 (ValidateCaseTransition) 和 T014 (ValidatePlanTransition) 包含 Domain 层业务逻辑，应按宪法 4.1 编写单元测试。

### 4. 计划文件集成 (Plan Integration)

**执行前检查**:
```
.claude/plans/phase-N-*.md 存在？
  ├─ 是 → 读取现有计划，验证与 tasks.md 对齐
  └─ 否 → 继续执行（不自动创建计划）
```

**计划文件命名约定**:
- Phase 范围: `.claude/plans/phase-{N}-{purpose}.md`
- 单任务: `.claude/plans/task-{TXXX}-{purpose}.md`

**与 tasks.md 的关联**:
- 每个测试必须关联 `tasks.md` 中的具体任务 ID
- 验收标准 (AC) 引用 `spec.md` 对应章节
- 依赖关系严格遵循 `tasks.md` 的 `Depends` 字段

### 5. 红灯阶段 (Red Phase) - 生成失败测试

**原则**:
- ✅ 只创建/修改 `*_test.go` 文件
- ❌ 禁止修改业务逻辑代码
- ✅ 测试用例必须引用 `spec.md` 中的验收标准 (AC)
- ✅ 测试文件头部注释关联任务 ID: `// tasks.md: T006`

**验收标准映射**:
从 `spec.md` 提取 AC 的示例：
```go
// tasks.md: T006 | spec.md: §9.1 ID 类型定义
func TestNewID(t *testing.T) {
    t.Parallel()
    // AC: ID 必须是有效的 UUID v4 格式
    id := NewID()
    _, err := uuid.Parse(id.String())
    assert.NoError(t, err)
    assert.Equal(t, 4, uuid.Parse(id.String()).Version())
}
```

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