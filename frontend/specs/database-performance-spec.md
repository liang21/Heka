# Heka 数据库性能规范

> PostgreSQL + Milvus 双数据库性能规范
> 版本：v1.0
> 日期：2025-05-15

**适用范围**：
- PostgreSQL 15+ （主数据存储）
- Milvus 2.3+ （向量存储）
- 所有建表和查询操作必须遵循本规范

---

## 1. PostgreSQL 性能规范

### 1.1 通用字段约定

所有表**必须**包含以下标准字段：

```sql
-- ===== 通用字段模板 =====
-- 每个表都应该包含以下字段（除非特殊说明）

CREATE TABLE template_table (
    -- 主键：必须使用 UUID
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- 外键关联：如果属于某个项目，必须包含 project_id
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- 审计字段：创建信息（必须）
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 审计字段：更新信息（可选，用于需要追踪变更的表）
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 软删除标记（可选，用于需要软删除的表）
    deleted_at TIMESTAMP,
    
    -- 乐观锁版本号（可选，用于需要并发控制的表）
    version INTEGER NOT NULL DEFAULT 0
);
```

**字段约定说明**：

| 字段 | 类型 | 约束 | 说明 | 必需 |
|------|------|------|------|------|
| `id` | `UUID` | `PRIMARY KEY` | 主键，使用 UUID 避免自增ID暴露 | ✅ |
| `project_id` | `UUID` | `NOT NULL, FK` | 多租户隔离，除用户表外都应包含 | ✅* |
| `created_by` | `UUID` | `NOT NULL, FK` | 创建人，引用 users(id) | ✅ |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | 创建时间，使用数据库默认值 | ✅ |
| `updated_by` | `UUID` | `FK` | 最后更新人 | ❌ |
| `updated_at` | `TIMESTAMP` | `DEFAULT NOW()` | 最后更新时间 | ❌ |
| `deleted_at` | `TIMESTAMP` | - | 软删除标记，NULL 表示未删除 | ❌ |
| `version` | `INTEGER` | `DEFAULT 0` | 乐观锁版本号 | ❌ |

\* 除 `users`、`project_members` 表外的所有业务表

---

### 1.2 索引设计原则

#### 1.2.1 索引命名规范

```sql
-- ===== 索引命名规范 =====
-- 格式：idx_{table}_{column}_{suffix}
-- 示例：

-- 单列索引
idx_test_cases_project
idx_files_type

-- 复合索引（最左前缀原则）
idx_test_cases_project_status
idx_executions_plan_status

-- 唯一索引
uniq_users_email
uniq_modules_project_parent_name

-- 全文搜索索引
idx_test_cases_fulltext

-- GIN 索引（用于数组/JSON）
idx_test_cases_tags
```

#### 1.2.2 必须创建的索引

**规则 1：外键列必须建索引**

```sql
-- ✅ 正确：所有外键列都有索引
CREATE INDEX idx_test_cases_project ON test_cases(project_id);
CREATE INDEX idx_test_cases_module ON test_cases(module_id);
CREATE INDEX idx_test_cases_created_by ON test_cases(created_by);
CREATE INDEX idx_files_project ON files(project_id);
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by);
```

**规则 2：高频查询条件必须建索引**

```sql
-- ✅ 正确：status 是常见过滤条件
CREATE INDEX idx_test_cases_status ON test_cases(status);

-- ✅ 正确：复合索引，支持项目+状态组合查询
CREATE INDEX idx_test_cases_project_status ON test_cases(project_id, status);

-- ✅ 正确：模块树查询需要 parent_id
CREATE INDEX idx_modules_parent ON modules(parent_id);
```

**规则 3：唯一约束必须建索引**

```sql
-- ✅ 正确：邮箱唯一约束
CREATE UNIQUE INDEX uniq_users_email ON users(email);

-- ✅ 正确：模块名称在同一项目和父节点下唯一
CREATE UNIQUE INDEX uniq_modules_project_parent_name 
    ON modules(project_id, parent_id, name);
```

#### 1.2.3 复合索引设计原则

**最左前缀原则**：

```sql
-- ===== 复合索引设计 =====
-- 如果查询经常同时用 A 和 B 过滤，创建 (A, B) 索引

-- ✅ 正确：按查询频率和选择性排序
CREATE INDEX idx_test_cases_project_status_priority 
    ON test_cases(project_id, status, priority);

-- 可以使用此索引的查询：
-- WHERE project_id = ? (使用索引)
-- WHERE project_id = ? AND status = ? (使用索引)
-- WHERE project_id = ? AND status = ? AND priority = ? (使用索引)
-- WHERE status = ? (不使用索引，违反最左前缀)

-- ✅ 正确：时间范围查询放在最后
CREATE INDEX idx_execution_results_execution_time 
    ON execution_results(execution_id, executed_at);

-- 时间范围查询
WHERE execution_id = ? AND executed_at >= ? AND executed_at <= ?
```

**选择性原则**（高选择性列在前）：

```sql
-- ===== 选择性排序 =====
-- 高选择性（唯一值多）→ 低选择性（唯一值少）

-- ✅ 正确：status 选择性高于 priority
CREATE INDEX idx_test_cases_project_status_priority 
    ON test_cases(project_id, status, priority);

-- ❌ 错误：priority 选择性低，不应该放在前面
-- CREATE INDEX idx_test_cases_project_priority_status 
--     ON test_cases(project_id, priority, status);
```

#### 1.2.4 全文搜索索引

```sql
-- ===== 全文搜索索引 =====
-- 用于标题、描述等文本字段的模糊搜索

-- ✅ 正确：中文和英文搜索
CREATE INDEX idx_test_cases_fulltext 
    ON test_cases USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- 使用示例
SELECT * FROM test_cases 
WHERE to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ to_tsquery('english', '登录');

-- ✅ 正确：支持中文搜索（需要 zhparser 扩展）
-- CREATE INDEX idx_test_cases_fulltext_zh 
--     ON test_cases USING GIN(to_tsvector('zhcfg', title || ' ' || COALESCE(description, '')));
```

#### 1.2.5 GIN 索引（数组/JSON）

```sql
-- ===== GIN 索引 =====
-- 用于数组类型、JSONB 类型的包含查询

-- ✅ 正确：标签数组搜索
CREATE INDEX idx_test_cases_tags ON test_cases USING GIN(tags);

-- 使用示例：查找包含特定标签的用例
SELECT * FROM test_cases WHERE tags @> ARRAY['登录', 'P0'];

-- ✅ 正确：JSONB 字段索引（如果有）
CREATE INDEX idx_files_metadata ON files USING GIN(metadata);
```

#### 1.2.6 部分索引（Conditional Index）

```sql
-- ===== 部分索引 =====
-- 只对满足条件的行创建索引，减少索引大小

-- ✅ 正确：只索引活跃的执行记录
CREATE INDEX idx_executions_active 
    ON test_executions(plan_id, status) 
    WHERE status IN ('in_progress', 'paused');

-- ✅ 正确：只索引未删除的数据
CREATE INDEX idx_test_cases_active 
    ON test_cases(project_id, created_at) 
    WHERE deleted_at IS NULL;
```

#### 1.2.7 表达式索引

```sql
-- ===== 表达式索引 =====
-- 对计算结果创建索引，加速函数查询

-- ✅ 正确：大小写不敏感搜索
CREATE INDEX idx_users_email_lower ON users(LOWER(email));

-- 使用示例
SELECT * FROM users WHERE LOWER(email) = LOWER('user@example.com');

-- ✅ 正确：日期范围查询
CREATE INDEX idx_executions_date ON test_executions(DATE(started_at));
```

#### 1.2.8 覆盖索引（Covering Index）

```sql
-- ===== 覆盖索引 =====
-- 包含查询所需的所有列，避免回表查询

-- ✅ 正确：列表查询常用字段
CREATE INDEX idx_test_cases_list 
    ON test_cases(project_id, status, created_at) 
    INCLUDE (title, priority, tags);

-- 使用示例：只需要索引中的列，不需要回表
SELECT title, priority, tags 
FROM test_cases 
WHERE project_id = ? AND status = ? 
ORDER BY created_at DESC;
```

---

### 1.3 大表预判和应对策略

#### 1.3.1 大表识别标准

| 表名 | 预估数据量 | 增长速率 | 大表级别 | 应对策略 |
|------|-----------|---------|---------|----------|
| `users` | < 1000 | 慢 | 小 | 无特殊处理 |
| `projects` | < 100 | 慢 | 小 | 无特殊处理 |
| `modules` | < 5000 | 慢 | 小 | 定期归档 |
| `test_cases` | 10万+ | 快 | **大** | 分区 + 归档 |
| `test_steps` | 50万+ | 快 | **大** | 分区 + 归档 |
| `test_plans` | < 1000 | 中 | 中 | 定期归档 |
| `test_executions` | 5万+ | 快 | **大** | 分区 + 归档 |
| `execution_results` | 100万+ | 快 | **超大** | 分区 + 归档 |
| `files` | < 5000 | 中 | 中 | 文件存储分离 |
| `document_chunks` | 10万+ | 快 | **大** | 分区 + 冷热分离 |
| `vector_embeddings` | 10万+ | 快 | **大** | Milvus 存储 |

#### 1.3.2 分区策略

**按时间分区**（适用于时间序列数据）：

```sql
-- ===== 按月分区：execution_results =====
-- 执行结果表按月分区，提升查询和归档效率

CREATE TABLE execution_results (
    id UUID DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL,
    test_case_id UUID NOT NULL,
    executor_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    bug_id VARCHAR(255),
    bug_url VARCHAR(500),
    notes TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (executed_at);

-- 创建分区（提前 3 个月，按月创建）
CREATE TABLE execution_results_2025_01 PARTITION OF execution_results
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE execution_results_2025_02 PARTITION OF execution_results
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE execution_results_2025_03 PARTITION OF execution_results
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

-- 创建默认分区（接收未来数据）
CREATE TABLE execution_results_default PARTITION OF execution_results
    DEFAULT;

-- 自动创建分区脚本
-- CREATE OR REPLACE FUNCTION create_monthly_partition()
-- RETURNS void AS $$
-- DECLARE
--     partition_name TEXT;
--     start_date TEXT;
--     end_date TEXT;
-- BEGIN
--     partition_name := 'execution_results_' || to_char(CURRENT_DATE + INTERVAL '1 month', 'YYYY_MM');
--     start_date := to_char(CURRENT_DATE + INTERVAL '1 month', 'YYYY-MM') || '-01';
--     end_date := to_char(CURRENT_DATE + INTERVAL '2 months', 'YYYY-MM') || '-01';
--     
--     EXECUTE format('
--         CREATE TABLE IF NOT EXISTS %I PARTITION OF execution_results
--         FOR VALUES FROM (%L) TO (%L)
--     ', partition_name, start_date, end_date);
-- END;
-- $$ LANGUAGE plpgsql;
```

**按项目分区**（适用于多租户数据）：

```sql
-- ===== 按项目 Hash 分区：test_cases =====
-- 适用于项目间数据隔离的场景

CREATE TABLE test_cases_partitioned (
    id UUID DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    module_id UUID,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    priority INTEGER NOT NULL DEFAULT 1,
    tags TEXT[],
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
) PARTITION BY HASH (project_id);

-- 创建 4 个分区
CREATE TABLE test_cases_p0 PARTITION OF test_cases_partitioned FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE test_cases_p1 PARTITION OF test_cases_partitioned FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE test_cases_p2 PARTITION OF test_cases_partitioned FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE test_cases_p3 PARTITION OF test_cases_partitioned FOR VALUES WITH (MODULUS 4, REMAINDER 3);
```

#### 1.3.3 归档策略

**冷热数据分离**：

```sql
-- ===== 归档表结构 =====
-- 与原表结构相同，用于存储历史数据

CREATE TABLE execution_results_archive (
    LIKE execution_results INCLUDING ALL
);

-- 归档索引（精简版）
CREATE INDEX idx_execution_results_archive_execution 
    ON execution_results_archive(execution_id);
CREATE INDEX idx_execution_results_archive_date 
    ON execution_results_archive(executed_at);

-- ===== 归档存储过程 =====
CREATE OR REPLACE FUNCTION archive_execution_results(older_than_months INT DEFAULT 6)
RETURNS INT AS $$
DECLARE
    archived_count INT;
BEGIN
    -- 移动数据到归档表
    WITH archived AS (
        DELETE FROM execution_results
        WHERE executed_at < CURRENT_DATE - (older_than_months || ' months')::INTERVAL
        RETURNING *
    )
    INSERT INTO execution_results_archive 
    SELECT * FROM archived;
    
    GET DIAGNOSTICS archived_count = ROW_COUNT;
    
    -- 返回归档的行数
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- 使用示例：归档 6 个月前的数据
-- SELECT archive_execution_results(6);
```

**自动归档任务**：

```sql
-- ===== pg_cron 定时任务 =====
-- 每月 1 号凌晨 2 点归档 6 个月前的数据

-- CREATE EXTENSION pg_cron;

-- SELECT cron.schedule(
--     'archive-execution-results',
--     '0 2 1 * *',  -- 每月 1 号凌晨 2 点
--     'SELECT archive_execution_results(6)'
-- );
```

#### 1.3.4 分表策略

**垂直分表**（按访问频率拆分）：

```sql
-- ===== 垂直分表：test_cases =====
-- 热数据：经常访问的字段
CREATE TABLE test_cases_hot (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    module_id UUID,
    title VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    priority INTEGER NOT NULL DEFAULT 1,
    tags TEXT[],
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- 冷数据：不常访问的大字段
CREATE TABLE test_cases_cold (
    test_case_id UUID PRIMARY KEY REFERENCES test_cases_hot(id) ON DELETE CASCADE,
    description TEXT,
    steps JSONB,  -- 序列化的步骤数据
    full_metadata JSONB
);

-- 创建视图，提供统一访问接口
CREATE VIEW test_cases AS
SELECT 
    h.*,
    c.description,
    c.steps,
    c.full_metadata
FROM test_cases_hot h
LEFT JOIN test_cases_cold c ON h.id = c.test_case_id;
```

---

### 1.4 分页查询注意事项

#### 1.4.1 基础分页（Offset + Limit）

```sql
-- ===== 基础分页（适用于小数据量） =====
-- 语法简单，但数据量大时性能差

-- ✅ 正确：基础分页
SELECT * FROM test_cases
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;   -- 第 1 页

-- ⚠️ 性能问题：越往后越慢
-- LIMIT 20 OFFSET 10000;  -- 需要扫描 10020 行
```

#### 1.4.2 游标分页（Cursor Pagination）

```sql
-- ===== 游标分页（推荐用于大数据量） =====
-- 使用上一页最后一条记录的 ID 作为游标

-- ✅ 正确：基于 ID 的游标分页
SELECT * FROM test_cases
WHERE project_id = ? 
  AND id > ?  -- 游标：上一页最后一条的 ID
ORDER BY id ASC
LIMIT 20;

-- ✅ 正确：基于时间戳的游标分页
SELECT * FROM test_cases
WHERE project_id = ? 
  AND created_at < ?  -- 游标：上一页最后一条的时间
ORDER BY created_at DESC
LIMIT 20;

-- 应用层需要返回最后一条记录的游标
-- {
--   "data": [...],
--   "next_cursor": "2025-05-15T10:30:00Z",  或最后一条的 ID
--   "has_more": true
-- }
```

#### 1.4.3 键集分页（Keyset Pagination）

```sql
-- ===== 键集分页（性能最优） =====
-- 使用复合索引，支持多字段排序

-- 假设有索引：idx_test_cases_project_status_priority
CREATE INDEX idx_test_cases_project_status_priority 
    ON test_cases(project_id, status, priority, created_at);

-- ✅ 正确：键集分页
SELECT * FROM test_cases
WHERE project_id = ? 
  AND status = ? 
  AND (
      priority > ?  -- 上一页最后一条的 priority
      OR (priority = ? AND created_at > ?)  -- 同 priority 时比较时间
  )
ORDER BY priority DESC, created_at DESC
LIMIT 20;

-- 这种方式可以利用索引，性能稳定
```

#### 1.4.4 分页查询优化原则

**原则 1：避免大 Offset**

```sql
-- ❌ 错误：大 Offset 性能差
SELECT * FROM test_cases
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;  -- 需要扫描 10020 行

-- ✅ 正确：使用游标分页
SELECT * FROM test_cases
WHERE created_at < '2025-05-01T00:00:00Z'
ORDER BY created_at DESC
LIMIT 20;
```

**原则 2：只查询需要的列**

```sql
-- ❌ 错误：SELECT *
SELECT * FROM test_cases LIMIT 20;

-- ✅ 正确：只查询列表页需要的字段
SELECT 
    id, 
    title, 
    status, 
    priority, 
    tags, 
    created_at
FROM test_cases 
LIMIT 20;

-- ✅ 更优：使用覆盖索引
CREATE INDEX idx_test_cases_list 
    ON test_cases(project_id, status, created_at) 
    INCLUDE (title, priority, tags);
```

**原则 3：预估总数优化**

```sql
-- ===== 预估总数（适用于大数据量） =====
-- 精确 COUNT(*) 在大表上很慢

-- ❌ 错误：精确 COUNT
SELECT COUNT(*) FROM test_cases WHERE project_id = ?;

-- ✅ 正确：使用预估
-- 方式 1：使用 EXPLAIN
-- SELECT reltuples::BIGINT AS estimate
-- FROM pg_class
-- WHERE relname = 'test_cases';

-- 方式 2：统计表 + 采样
CREATE TABLE table_stats (
    table_name TEXT PRIMARY KEY,
    row_count BIGINT,
    last_updated TIMESTAMP
);

-- 定时更新统计信息
-- INSERT INTO table_stats (table_name, row_count, last_updated)
-- SELECT 'test_cases', COUNT(*), NOW()
-- FROM test_cases TABLESAMPLE SYSTEM(1);  -- 1% 采样
```

#### 1.4.5 分页查询最佳实践

```sql
-- ===== 分页查询模板 =====

-- 1. 列表页（使用游标分页）
-- 第一页
SELECT 
    id, title, status, priority, tags, created_at
FROM test_cases
WHERE project_id = 'project-uuid'
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 20;

-- 下一页（使用上一页最后一条的 created_at）
SELECT 
    id, title, status, priority, tags, created_at
FROM test_cases
WHERE project_id = 'project-uuid'
  AND deleted_at IS NULL
  AND created_at < '2025-05-15T10:00:00Z'  -- 上一页最后一条的时间
ORDER BY created_at DESC
LIMIT 20;

-- 2. 带过滤条件的分页（使用键集分页）
-- 需要对应的复合索引：idx_test_cases_project_status_priority
SELECT 
    id, title, status, priority, tags, created_at
FROM test_cases
WHERE project_id = 'project-uuid'
  AND status = 'ready'
  AND (
      priority > 2
      OR (priority = 2 AND created_at < '2025-05-15T10:00:00Z')
  )
ORDER BY priority DESC, created_at DESC
LIMIT 20;

-- 3. 总数查询（使用缓存）
-- 从 Redis 缓存获取，异步更新
-- cache_key = "project:{id}:testcases:count"
```

---

### 1.5 查询优化清单

#### 1.5.1 必须避免的反模式

```sql
-- ===== 反模式 1：SELECT * =====
-- ❌ 错误
SELECT * FROM test_cases WHERE project_id = ?;

-- ✅ 正确：只查询需要的列
SELECT id, title, status FROM test_cases WHERE project_id = ?;

-- ===== 反模式 2：在索引列上使用函数 =====
-- ❌ 错误：无法使用索引
SELECT * FROM test_cases WHERE LOWER(title) = 'test';

-- ✅ 正确：创建表达式索引
CREATE INDEX idx_test_cases_title_lower ON test_cases(LOWER(title));
-- 或使用 ILIKE（如果支持）
SELECT * FROM test_cases WHERE title ILIKE 'test';

-- ===== 反模式 3：隐式类型转换 =====
-- ❌ 错误：类型不匹配导致无法使用索引
SELECT * FROM test_cases WHERE project_id = 123;  -- project_id 是 UUID

-- ✅ 正确：使用正确的类型
SELECT * FROM test_cases WHERE project_id = '123e4567-e89b-12d3-a456-426614174000';

-- ===== 反模式 4：OR 条件 =====
-- ❌ 错误：OR 条件可能导致全表扫描
SELECT * FROM test_cases 
WHERE status = 'ready' OR priority = 3;

-- ✅ 正确：使用 UNION ALL（如果有对应索引）
SELECT * FROM test_cases WHERE status = 'ready'
UNION ALL
SELECT * FROM test_cases WHERE priority = 3 AND status != 'ready';

-- 或使用 IN
SELECT * FROM test_cases WHERE status IN ('ready', 'other');

-- ===== 反模式 5：NOT IN/NOT EXISTS =====
-- ❌ 错误：NOT IN 可能很慢
SELECT * FROM test_cases 
WHERE id NOT IN (SELECT test_case_id FROM plan_test_cases);

-- ✅ 正确：使用 LEFT JOIN
SELECT tc.* 
FROM test_cases tc
LEFT JOIN plan_test_cases ptc ON tc.id = ptc.test_case_id
WHERE ptc.test_case_id IS NULL;
```

#### 1.5.2 查询性能分析

```sql
-- ===== 分析查询执行计划 =====
EXPLAIN ANALYZE
SELECT * FROM test_cases
WHERE project_id = 'xxx' AND status = 'ready'
ORDER BY created_at DESC
LIMIT 20;

-- 关注指标：
-- 1. Index Scan：使用了索引
-- 2. Seq Scan：全表扫描（需要优化）
-- 3. execution time：执行时间
-- 4. rows：实际扫描行数

-- ===== 查找缺失的索引 =====
-- 查看高频慢查询
SELECT 
    query,
    calls,
    total_time,
    mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- 查看未使用的索引
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey';
```

#### 1.5.3 连接查询优化

```sql
-- ===== JOIN 优化 =====

-- ✅ 正确：小表驱动大表
SELECT tc.*, p.name as project_name
FROM test_cases tc
INNER JOIN projects p ON tc.project_id = p.id
WHERE tc.status = 'ready'
LIMIT 20;

-- ✅ 正确：使用 LATERAL JOIN（关联子查询优化）
SELECT 
    tc.id,
    tc.title,
    latest_result.status
FROM test_cases tc
LEFT JOIN LATERAL (
    SELECT status
    FROM execution_results er
    WHERE er.test_case_id = tc.id
    ORDER BY executed_at DESC
    LIMIT 1
) latest_result ON true
WHERE tc.project_id = 'xxx'
LIMIT 20;
```

---

### 1.6 并发控制

#### 1.6.1 乐观锁（Optimistic Locking）

```sql
-- ===== 乐观锁实现 =====
-- 使用 version 字段检测并发修改

-- ✅ 正确：乐观锁更新
UPDATE test_cases 
SET 
    title = 'New Title',
    description = 'New Description',
    updated_by = 'user-uuid',
    updated_at = CURRENT_TIMESTAMP,
    version = version + 1
WHERE id = 'testcase-uuid'
  AND version = 5;  -- 当前版本号

-- 检查是否更新成功
-- 如果 affected_rows = 0，说明版本冲突

-- 应用层处理：
-- if affected_rows == 0:
--     return Error("版本冲突，数据已被其他用户修改，请刷新后重试")
```

#### 1.6.2 悲观锁（Pessimistic Locking）

```sql
-- ===== 悲观锁实现 =====
-- 使用 FOR UPDATE 行级锁

-- ✅ 正确：悲观锁查询
BEGIN;

SELECT * FROM test_cases
WHERE id = 'testcase-uuid'
FOR UPDATE;  -- 锁定该行，其他事务需要等待

-- 执行业务逻辑
UPDATE test_cases SET ... WHERE id = 'xxx';

COMMIT;  -- 释放锁
```

#### 1.6.3 死锁预防

```sql
-- ===== 死锁预防原则 =====

-- 1. 按固定顺序访问表
-- ✅ 正确：先访问 test_cases，再访问 test_plans
BEGIN;
SELECT * FROM test_cases WHERE id = 'xxx' FOR UPDATE;
SELECT * FROM test_plans WHERE id = 'yyy' FOR UPDATE;
COMMIT;

-- 2. 尽量缩短事务时间
-- ❌ 错误：长事务
BEGIN;
SELECT * FROM test_cases WHERE id = 'xxx' FOR UPDATE;
-- 执行耗时的业务逻辑（如调用外部 API）
UPDATE test_cases SET ...;
COMMIT;

-- ✅ 正确：先查询，再快速更新
-- 在应用层获取数据
SELECT * FROM test_cases WHERE id = 'xxx';

-- 执行业务逻辑
-- ...

-- 快速更新
BEGIN;
UPDATE test_cases SET ... WHERE id = 'xxx' AND version = ?;
COMMIT;

-- 3. 设置锁超时
SET lock_timeout = '5s';  -- 5 秒超时
```

---

### 1.7 数据库配置优化

#### 1.7.1 PostgreSQL 配置

```ini
# ===== postgresql.conf 优化配置 =====
# 适用于 4核 8GB 单机部署

# 连接配置
max_connections = 100          # 最大连接数
shared_buffers = 2GB          # 共享缓冲区（内存的 25%）
effective_cache_size = 6GB     # 有效缓存大小（内存的 75%）
work_mem = 16MB                # 每个查询的工作内存
maintenance_work_mem = 512MB   # 维护操作内存

# WAL 配置
wal_buffers = 16MB
min_wal_size = 1GB
max_wal_size = 4GB
wal_compression = on           # WAL 压缩

# 查询优化
random_page_cost = 1.1         # SSD 使用 1.1，HDD 使用 4.0
effective_io_concurrency = 200 # 并发 IO 数量
max_worker_processes = 4       # 最大后台进程数
max_parallel_workers_per_gather = 2
max_parallel_workers = 4

# 日志配置
log_min_duration_statement = 1000  # 记录超过 1 秒的查询
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_connections = on
log_disconnections = on
log_duration = off
log_lock_waits = on

# 自动清理
autovacuum = on
autovacuum_max_workers = 2
autovacuum_naptime = 1min
```

#### 1.7.2 连接池配置

```go
// ===== Go 连接池配置 =====
// 使用 database/sql

db, err := sql.Open("postgres", dsn)

// 最大开放连接数（<= max_connections）
db.SetMaxOpenConns(50)

// 最大空闲连接数
db.SetMaxIdleConns(10)

// 连接最大存活时间
db.SetConnMaxLifetime(1 * time.Hour)

// 连接最大空闲时间
db.SetConnMaxIdleTime(10 * time.Minute)

// 连接复用
db.SetConnMaxLifetime(1 * time.Hour)
```

---

## 2. Milvus 性能规范

### 2.1 Collection 设计原则

#### 2.1.1 Collection 命名规范

```python
# ===== Collection 命名规范 =====
# 格式：{业务实体}_{数据类型}
# 示例：

document_chunks      # 文档块向量
qa_pairs             # 问答对向量
test_case_embeddings # 测试用例向量
```

#### 2.1.2 字段设计

```python
# ===== Collection 字段模板 =====
from pymilvus import Collection, FieldSchema, CollectionSchema, DataType

# 文档块 Collection（核心 RAG 场景）
document_chunks_schema = CollectionSchema([
    # 主键字段
    FieldSchema(name="id", dtype=DataType.VARCHAR, max_length=36, is_primary=True, auto_id=False),
    
    # 业务字段
    FieldSchema(name="chunk_id", dtype=DataType.VARCHAR, max_length=36),      # PostgreSQL chunk ID
    FieldSchema(name="file_id", dtype=DataType.VARCHAR, max_length=36),       # 所属文件
    FieldSchema(name="project_id", dtype=DataType.VARCHAR, max_length=36),    # 所属项目
    FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=65535),    # 文本内容
    FieldSchema(name="index", dtype=DataType.INT64),                          # 块序号
    
    # 向量字段
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=1536),     # OpenAI Embedding
    
    # 过滤字段（用于过滤查询）
    FieldSchema(name="tokens", dtype=DataType.INT64),                         # Token 数量
    FieldSchema(name="created_at", dtype=DataType.INT64),                     # 创建时间戳
    
    # 全文搜索字段（可选，需要 Milvus 2.4+）
    # FieldSchema(name="fulltext", dtype=DataType.VARCHAR, max_length=65535, enable_analyzer=True),
])

# 创建 Collection
collection = Collection(
    name="document_chunks",
    schema=document_chunks_schema,
    description="文档块向量集合，用于 RAG 检索"
)
```

#### 2.1.3 索引设计

```python
# ===== 向量索引配置 =====
from pymilvus import Collection

# 方案 1：IVF_FLAT（精度最高，内存占用大）
index_params = {
    "index_type": "IVF_FLAT",
    "metric_type": "L2",      # 或 "IP" (内积)
    "params": {
        "nlist": 128          # 聚类中心数，推荐 sqrt(数据量)
    }
}

# 方案 2：IVF_PQ（精度和内存的平衡）
index_params = {
    "index_type": "IVF_PQ",
    "metric_type": "L2",
    "params": {
        "nlist": 128,
        "m": 8                # PQ 压缩因子，必须是 dim 的约数
    }
}

# 方案 3：HNSW（速度快，内存占用大）
index_params = {
    "index_type": "HNSW",
    "metric_type": "L2",
    "params": {
        "M": 16,              # 每个节点的最大连接数
        "efConstruction": 200 # 构建索引时的搜索范围
    }
}

# 创建索引
collection.create_index(
    field_name="embedding",
    index_params=index_params,
    index_name="vector_index"
)

# ===== 标量索引配置 =====
# 为过滤字段创建索引（Milvus 2.4+ 支持）
collection.create_index(
    field_name="project_id",
    index_name="project_index"
)
collection.create_index(
    field_name="file_id",
    index_name="file_index"
)
```

#### 2.1.4 索引选择指南

| 数据量 | 索引类型 | 参数配置 | 适用场景 |
|--------|---------|----------|----------|
| < 10万 | IVF_FLAT | nlist=64 | 小规模，高精度要求 |
| 10万-100万 | IVF_PQ | nlist=128, m=8 | 中等规模，平衡性能 |
| 100万-1000万 | IVF_PQ | nlist=256, m=16 | 大规模，内存受限 |
| > 1000万 | HNSW | M=16, efConstruction=200 | 超大规模，低延迟 |

```python
# ===== 动态选择索引参数 =====
def get_index_params(num_vectors: int, dim: int) -> dict:
    """根据数据量选择合适的索引参数"""
    
    if num_vectors < 100000:
        # 小规模：IVF_FLAT
        return {
            "index_type": "IVF_FLAT",
            "metric_type": "L2",
            "params": {"nlist": max(64, int(num_vectors ** 0.5))}
        }
    elif num_vectors < 1000000:
        # 中规模：IVF_PQ
        m_candidates = [d for d in [8, 16, 32] if dim % d == 0]
        return {
            "index_type": "IVF_PQ",
            "metric_type": "L2",
            "params": {
                "nlist": 128,
                "m": m_candidates[0] if m_candidates else 8
            }
        }
    else:
        # 大规模：HNSW
        return {
            "index_type": "HNSW",
            "metric_type": "L2",
            "params": {
                "M": 16,
                "efConstruction": 200
            }
        }
```

---

### 2.2 插入优化

#### 2.2.1 批量插入

```python
# ===== 批量插入优化 =====
from pymilvus import Collection

# ✅ 正确：批量插入（推荐）
def insert_chunks_batch(collection: Collection, chunks: list, batch_size=1000):
    """批量插入文档块"""
    
    total = len(chunks)
    for i in range(0, total, batch_size):
        batch = chunks[i:i + batch_size]
        
        # 准备数据
        data = [
            [c["id"] for c in batch],
            [c["chunk_id"] for c in batch],
            [c["file_id"] for c in batch],
            [c["project_id"] for c in batch],
            [c["content"] for c in batch],
            [c["index"] for c in batch],
            [c["embedding"] for c in batch],
            [c["tokens"] for c in batch],
            [c["created_at"] for c in batch],
        ]
        
        # 插入
        collection.insert(data)
        
        # 每 10000 条刷新一次
        if (i + batch_size) % 10000 == 0:
            collection.flush()
    
    # 最终刷新
    collection.flush()

# ❌ 错误：逐条插入（性能差）
for chunk in chunks:
    collection.insert([chunk["id"]], [chunk["embedding"]], ...)
```

#### 2.2.2 异步插入

```python
# ===== 异步插入 =====
import asyncio
from pymilvus import Collection

async def async_insert_chunks(collection: Collection, chunks: list):
    """异步批量插入"""
    
    loop = asyncio.get_event_loop()
    
    # 分批插入
    tasks = []
    batch_size = 1000
    
    for i in range(0, len(chunks), batch_size):
        batch = chunks[i:i + batch_size]
        
        # 在线程池中执行插入
        task = loop.run_in_executor(
            None,
            lambda b=batch: collection.insert(prepare_data(b))
        )
        tasks.append(task)
    
    # 等待所有任务完成
    await asyncio.gather(*tasks)
    collection.flush()
```

---

### 2.3 搜索优化

#### 2.3.1 基础搜索

```python
# ===== 向量搜索 =====
from pymilvus import Collection

def search_similar_chunks(
    collection: Collection,
    query_embedding: list[float],
    project_id: str,
    top_k: int = 5,
    filters: str = None
) -> list:
    """搜索相似文档块"""
    
    # 构建过滤表达式
    if filters:
        expr = f"project_id == '{project_id}' and {filters}"
    else:
        expr = f"project_id == '{project_id}'"
    
    # 搜索参数
    search_params = {
        "metric_type": "L2",
        "params": {
            "nprobe": 16  # 搜索的聚类中心数，影响召回率和速度
        }
    }
    
    # 执行搜索
    results = collection.search(
        data=[query_embedding],
        anns_field="embedding",
        param=search_params,
        limit=top_k,
        expr=expr,
        output_fields=["content", "file_id", "tokens"]
    )
    
    return results[0]

# ===== 使用示例 =====
results = search_similar_chunks(
    collection=collection,
    query_embedding=[0.1, 0.2, ...],  # 1536 维向量
    project_id="project-uuid",
    top_k=5,
    filters="tokens >= 100"  # 可选过滤条件
)

for result in results:
    print(f"距离: {result.distance}")
    print(f"内容: {result.entity.get('content')}")
```

#### 2.3.2 混合搜索（向量 + 全文）

```python
# ===== 混合搜索 =====
def hybrid_search(
    collection: Collection,
    query_text: str,
    query_embedding: list[float],
    project_id: str,
    top_k: int = 5,
    alpha: float = 0.7  # 向量权重
) -> list:
    """混合搜索：向量 + 全文（需要 Milvus 2.4+）"""
    
    # 1. 向量搜索
    vector_results = search_similar_chunks(
        collection=collection,
        query_embedding=query_embedding,
        project_id=project_id,
        top_k=top_k * 2  # 多取一些用于重排序
    )
    
    # 2. 全文搜索
    # （需要 Milvus 2.4+ 的全文搜索功能）
    # text_results = collection.text_search(
    #     query_text=query_text,
    #     expr=f"project_id == '{project_id}'",
    #     limit=top_k * 2
    # )
    
    # 3. 融合排序（RRF 算法）
    # scores = rrf_fusion(vector_results, text_results, alpha=alpha)
    
    return vector_results[:top_k]

def rrf_fusion(vector_results, text_results, alpha=0.7, k=60):
    """RRF (Reciprocal Rank Fusion) 融合算法"""
    
    scores = {}
    
    # 向量结果
    for rank, result in enumerate(vector_results):
        doc_id = result.id
        scores[doc_id] = scores.get(doc_id, 0) + alpha / (k + rank + 1)
    
    # 全文结果
    for rank, result in enumerate(text_results):
        doc_id = result.id
        scores[doc_id] = scores.get(doc_id, 0) + (1 - alpha) / (k + rank + 1)
    
    # 按分数排序
    ranked = sorted(scores.items(), key=lambda x: x[1], reverse=True)
    return ranked
```

#### 2.3.3 搜索参数调优

```python
# ===== 搜索参数调优 =====

# 参数 1：nprobe（搜索的聚类中心数）
# 影响：召回率和速度的平衡
search_params = {
    "metric_type": "L2",
    "params": {
        "nprobe": 16  # 推荐：nlist 的 1/10 ~ 1/2
    }
}

# nprobe 选择指南：
# - 追求速度：nprobe = nlist // 20
# - 追求召回：nprobe = nlist // 2

# 参数 2：ef（HNSW 搜索范围）
# 影响：HNSW 索引的召回率
search_params_hnsw = {
    "metric_type": "L2",
    "params": {
        "ef": 64  # 推荐：ef >= top_k * 2
    }
}

# 参数 3：top_k（返回结果数）
# 建议：不要设置过大，影响性能
# - 列表页：top_k = 5 ~ 10
# - 详情页：top_k = 10 ~ 20
# - 批量操作：top_k = 20 ~ 50
```

---

### 2.4 Collection 维护

#### 2.4.1 数据清理

```python
# ===== 删除过期数据 =====
def delete_old_chunks(collection: Collection, days: int = 180):
    """删除指定天数之前的数据"""
    
    import time
    timestamp = int(time.time()) - days * 86400
    
    # 删除
    collection.delete(
        expr=f"created_at < {timestamp}"
    )
    
    # 删除后必须调用 compact
    collection.compact()

# ===== 清理已删除文件的向量 =====
def clean_orphan_vectors(collection: Collection, file_ids: list[str]):
    """清理孤立文件（文件已删除但向量还在）"""
    
    # 批量删除（Milvus 支持 IN 表达式）
    file_ids_str = ", ".join([f"'{fid}'" for fid in file_ids])
    collection.delete(
        expr=f"file_id in [{file_ids_str}]"
    )
    
    collection.compact()
```

#### 2.4.2 数据压缩

```python
# ===== 数据压缩 =====
def compact_collection(collection: Collection):
    """压缩 Collection，释放空间"""
    
    # 1. 调用 compact
    collection.compact()
    
    # 2. 等待 compact 完成
    from pymilvus import utility
    utility.wait_for_compaction_completed(collection.name)
```

#### 2.4.3 性能监控

```python
# ===== 性能监控 =====
from pymilvus import utility

def get_collection_stats(collection_name: str) -> dict:
    """获取 Collection 统计信息"""
    
    from pymilvus import Collection
    collection = Collection(collection_name)
    
    stats = {
        "name": collection_name,
        "num_entities": collection.num_entities,
        "index_info": collection.index().to_dict(),
    }
    
    return stats

# 使用示例
stats = get_collection_stats("document_chunks")
print(f"文档数量: {stats['num_entities']}")
print(f"索引信息: {stats['index_info']}")
```

---

### 2.5 Milvus 配置优化

#### 2.5.1 内存配置

```yaml
# ===== milvus.yaml 配置 =====
# 适用于 4 核 8GB 单机部署

# 缓存配置
cache:
  size: 2GB               # 缓存大小
  slotSize: 16MB          # 每个槽位大小

# 查询配置
queryNode:
  gracefulTime: 1000ms    # 优雅停止时间
  searchReceiveBufSize: 1024  # 搜索接收缓冲区大小

# 插入配置
dataNode:
  # 插入缓冲区大小
  bufferSize: 1GB
  
  # 刷新间隔
  flushInsertInterval: 2s

# 索引配置
indexNode:
  # 索引构建线程数
  maxConcurrency: 4
```

#### 2.5.2 连接池配置

```python
# ===== 连接池配置 =====
from pymilvus import connections

# 连接到 Milvus
connections.connect(
    alias="default",
    host="localhost",
    port="19530",
    # 连接池配置
    pool_size=10,            # 连接池大小
    timeout=30,              # 超时时间（秒）
    retry_on_rate_limit=True  # 限流时自动重试
)
```

---

## 3. 跨数据库事务处理

### 3.1 最终一致性方案

```go
// ===== PostgreSQL + Milvus 事务处理 =====
// 由于 Milvus 不支持 ACID 事务，采用最终一致性

// 1. 先写入 PostgreSQL（事务保证）
func (s *Service) CreateDocument(ctx context.Context, doc *Document) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1.1 写入文档元数据
    if err := s.docRepo.Create(ctx, tx, doc); err != nil {
        return err
    }

    // 1.2 创建待处理记录
    task := &IndexTask{
        DocumentID: doc.ID,
        Status:     "pending",
    }
    if err := s.taskRepo.Create(ctx, tx, task); err != nil {
        return err
    }

    // 1.3 提交事务
    if err := tx.Commit(); err != nil {
        return err
    }

    // 2. 异步处理 Milvus 索引
    go s.processIndexTask(doc.ID)

    return nil
}

// 2. 后台 Worker 处理索引
func (s *Service) processIndexTask(docID string) {
    // 2.1 从 PostgreSQL 读取文档
    doc, err := s.docRepo.FindByID(context.Background(), docID)
    if err != nil {
        s.taskRepo.UpdateStatus(docID, "failed", err.Error())
        return
    }

    // 2.2 分块
    chunks := s.chunker.Split(doc.Content)

    // 2.3 向量化
    embeddings := s.aiClient.Embed(ctx, chunks)

    // 2.4 写入 Milvus
    if err := s.milvusClient.Insert(docID, chunks, embeddings); err != nil {
        s.taskRepo.UpdateStatus(docID, "failed", err.Error())
        return
    }

    // 2.5 更新任务状态
    s.taskRepo.UpdateStatus(docID, "completed", "")
}
```

### 3.2 补偿机制

```go
// ===== 补偿机制 =====
// 定时检查未完成的任务，自动重试

func (s *Service) StartCompensationWorker() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        s.retryFailedTasks()
    }
}

func (s *Service) retryFailedTasks() {
    // 1. 查询失败或超时的任务
    tasks, err := s.taskRepo.FindFailed(context.Background(), time.Hour*24)
    if err != nil {
        return
    }

    // 2. 重新处理
    for _, task := range tasks {
        go s.processIndexTask(task.DocumentID)
    }
}
```

---

## 4. 性能监控和告警

### 4.1 PostgreSQL 监控指标

```sql
-- ===== 性能监控查询 =====

-- 1. 表大小监控
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;

-- 2. 索引使用率监控
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- 3. 慢查询监控
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 1000  -- 超过 1 秒
ORDER BY mean_time DESC
LIMIT 20;

-- 4. 表膨胀监控
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch
FROM pg_stat_user_tables
ORDER BY seq_tup_read DESC;

-- 5. 连接数监控
SELECT 
    count(*) AS total_connections,
    count(*) FILTER (WHERE state = 'active') AS active,
    count(*) FILTER (WHERE state = 'idle') AS idle,
    count(*) FILTER (WHERE state = 'idle in transaction') AS idle_in_transaction
FROM pg_stat_activity;
```

### 4.2 Milvus 监控指标

```python
# ===== Milvus 监控 =====
from pymilvus import utility

def get_milvus_stats() -> dict:
    """获取 Milvus 统计信息"""
    
    stats = {
        "collections": [],
        "memory_usage": {},
    }
    
    # 获取所有 Collection
    collections = utility.list_collections()
    
    for coll_name in collections:
        coll = Collection(coll_name)
        
        coll_stats = {
            "name": coll_name,
            "num_entities": coll.num_entities,
            "index": coll.index().to_dict() if coll.index() else None,
        }
        
        stats["collections"].append(coll_stats)
    
    return stats

# 使用示例
stats = get_milvus_stats()
for coll in stats["collections"]:
    print(f"Collection: {coll['name']}")
    print(f"  文档数量: {coll['num_entities']}")
```

---

## 5. AI 建表执行清单

### 5.1 PostgreSQL 建表检查清单

在创建新表时，必须检查以下项目：

```sql
-- ===== 建表前检查清单 =====

-- [ ] 1. 是否包含通用字段？
--      ✅ id (UUID, PRIMARY KEY)
--      ✅ created_at (TIMESTAMP, DEFAULT NOW())
--      ✅ created_by (UUID, FK → users)
--      ✅ project_id (UUID, FK → projects) -- 如适用

-- [ ] 2. 主键是否使用 UUID？
--      ✅ id UUID PRIMARY KEY DEFAULT gen_random_uuid()

-- [ ] 3. 外键是否有索引？
--      ✅ CREATE INDEX idx_{table}_{column} ON {table}({column});

-- [ ] 4. 是否需要时间戳索引？（用于时间范围查询）
--      ✅ CREATE INDEX idx_{table}_created_at ON {table}(created_at);

-- [ ] 5. 是否需要软删除？
--      ✅ deleted_at TIMESTAMP
--      ✅ CREATE INDEX idx_{table}_deleted_at ON {table}(deleted_at) 
--           WHERE deleted_at IS NULL;

-- [ ] 6. 数组字段是否有 GIN 索引？
--      ✅ CREATE INDEX idx_{table}_array_field ON {table} USING GIN(array_field);

-- [ ] 7. 是否需要唯一约束？
--      ✅ CREATE UNIQUE INDEX uniq_{table}_{field} ON {table}({field});

-- [ ] 8. 是否需要复合索引？（基于查询模式）
--      ✅ CREATE INDEX idx_{table}_a_b ON {table}(a, b);

-- [ ] 9. 是否需要全文字段？（用于搜索）
--      ✅ CREATE INDEX idx_{table}_fulltext ON {table} 
--           USING GIN(to_tsvector('english', field));

-- [ ] 10. 大表是否需要分区？（预估 > 100万行）
--       ✅ PARTITION BY RANGE (created_at);
```

### 5.2 Milvus 建表检查清单

```python
# ===== Milvus 建表前检查清单 =====

# [ ] 1. 命名是否符合规范？
#      ✅ {业务实体}_{数据类型}
#      ✅ 示例：document_chunks, qa_pairs

# [ ] 2. 是否包含主键字段？
#      ✅ id (VARCHAR, PRIMARY KEY)

# [ ] 3. 是否包含业务字段？
#      ✅ chunk_id (VARCHAR, 关联 PostgreSQL)
#      ✅ file_id/project_id (VARCHAR, 过滤查询)
#      ✅ content (VARCHAR, 原始内容)
#      ✅ created_at (INT64, 时间戳)

# [ ] 4. 向量字段维度是否正确？
#      ✅ dim=1536 (OpenAI text-embedding-ada-002)
#      ✅ dim=768 (sentence-transformers)
#      ✅ dim=1024 (Cohere embed-v3)

# [ ] 5. 是否创建了合适的索引？
#      ✅ < 10万: IVF_FLAT (nlist=64)
#      ✅ 10万-100万: IVF_PQ (nlist=128, m=8)
#      ✅ > 100万: HNSW (M=16)

# [ ] 6. metric_type 是否正确？
#      ✅ L2 (欧氏距离，适用于归一化向量)
#      ✅ IP (内积，适用于未归一化向量)
#      ✅ COSINE (余弦相似度，Milvus 会自动归一化)

# [ ] 7. 是否为过滤字段创建了索引？
#      ✅ collection.create_index(field_name="project_id")

# [ ] 8. 是否设置了搜索参数？
#      ✅ nprobe = nlist // 10 (IVF 系列)
#      ✅ ef = top_k * 2 (HNSW)
```

---

## 6. 快速参考

### 6.1 常用索引模板

```sql
-- ===== 快速索引模板 =====

-- 单列索引
CREATE INDEX idx_{table}_{column} ON {table}({column});

-- 复合索引（按选择性排序）
CREATE INDEX idx_{table}_a_b_c ON {table}(a, b, c);

-- 唯一索引
CREATE UNIQUE INDEX uniq_{table}_{column} ON {table}({column});

-- 部分索引（只索引活跃数据）
CREATE INDEX idx_{table}_active ON {table}(project_id, created_at) 
    WHERE deleted_at IS NULL;

-- GIN 索引（数组）
CREATE INDEX idx_{table}_array ON {table} USING GIN(array_field);

-- 全文索引
CREATE INDEX idx_{table}_fulltext ON {table} 
    USING GIN(to_tsvector('english', text_field));

-- 表达式索引
CREATE INDEX idx_{table}_expr ON {table}(LOWER(column));

-- 覆盖索引
CREATE INDEX idx_{table}_covering ON {table}(a, b) INCLUDE (c, d);
```

### 6.2 常用查询模板

```sql
-- ===== 快速查询模板 =====

-- 分页查询（游标）
SELECT * FROM {table}
WHERE project_id = ? AND id > ?
ORDER BY id ASC
LIMIT ?;

-- 列表查询（覆盖索引）
SELECT id, title, status FROM {table}
WHERE project_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT ?;

-- 统计查询（使用缓存）
SELECT COUNT(*) FROM {table} WHERE project_id = ?;
-- 替代方案：从 Redis 获取

-- 关联查询（小表驱动大表）
SELECT t1.*, t2.name
FROM {large_table} t1
INNER JOIN {small_table} t2 ON t1.project_id = t2.id
WHERE t1.status = ?
LIMIT ?;
```

---

## 7. 软删除与硬删除决策表

### 7.1 删除策略

| 表名 | 策略 | 原因 | 归档周期 |
|------|------|------|----------|
| `users` | 软删除 | 用户数据关联广泛，删除影响大 | 永久保留 |
| `projects` | 软删除 | 项目数据可能需要恢复 | 1 年后归档 |
| `modules` | 硬删除 | 解除用例关联即可删除 | - |
| `test_cases` | 软删除 | 可能需要恢复，且关联执行历史 | 6 个月后归档 |
| `test_steps` | 硬删除（级联） | 跟随用例删除 | - |
| `test_plans` | 软删除 | 计划数据需要历史追溯 | 1 年后归档 |
| `test_executions` | 软删除 | 执行历史需要保留 | 6 个月后归档 |
| `execution_results` | 分区归档 | 数据量大，按月归档 | 6 个月 |
| `files` | 软删除 + 异步清理 | 先标记删除，后台清理文件 | 30 天后清理文件 |
| `document_chunks` | 硬删除（级联） | 跟随文件删除 | - |
| `tags` | 硬删除 | 只有关联关系 | - |
| `project_members` | 硬删除 | 成员移除无需保留 | - |

### 7.2 软删除实现

```sql
-- 软删除查询必须过滤 deleted_at
-- ✅ 正确
SELECT * FROM test_cases WHERE project_id = ? AND deleted_at IS NULL;

-- 部分索引自动排除已删除数据
CREATE INDEX idx_test_cases_active 
    ON test_cases(project_id, created_at) 
    WHERE deleted_at IS NULL;

-- 软删除操作
UPDATE test_cases SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?;
```

### 7.3 文件异步清理

```go
// 文件软删除后，后台 Worker 清理物理文件
func (w *CleanupWorker) ProcessDeletedFiles(ctx context.Context) {
    // 查找 30 天前软删除的文件
    files, _ := w.fileRepo.FindDeletedOlderThan(ctx, 30*24*time.Hour)
    
    for _, f := range files {
        // 删除物理文件
        w.storage.Delete(ctx, f.Path)
        // 硬删除数据库记录
        w.fileRepo.HardDelete(ctx, f.ID)
    }
}
```

---

## 8. 数据迁移策略

### 8.1 迁移工具选型

使用 [golang-migrate](https://github.com/golang-migrate/migrate)：

```bash
# 安装
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 8.2 迁移文件组织

```
scripts/migration/
├── 000001_init_schema.up.sql        # 初始建表
├── 000001_init_schema.down.sql      # 回滚
├── 000002_add_file_index.up.sql     # 新增字段
├── 000002_add_file_index.down.sql
├── 000003_add_collections.up.sql
├── 000003_add_collections.down.sql
└── ...
```

### 8.3 迁移文件模板

```sql
-- scripts/migration/000001_init_schema.up.sql
-- 初始数据库 Schema

-- 用户
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_users_email_lower ON users(LOWER(email));
```

```sql
-- scripts/migration/000001_init_schema.down.sql
DROP TABLE IF EXISTS users CASCADE;
```

### 8.4 迁移执行

```bash
# 执行所有待执行的迁移
migrate -path scripts/migration -database "postgres://heka:pass@localhost:5432/heka?sslmode=disable" up

# 回滚最近一次迁移
migrate -path scripts/migration -database "postgres://heka:pass@localhost:5432/heka?sslmode=disable" down 1

# 查看当前版本
migrate -path scripts/migration -database "postgres://heka:pass@localhost:5432/heka?sslmode=disable" version
```

### 8.5 迁移规范

- 每个迁移必须包含 `up` 和 `down`
- 迁移文件一旦合并不再修改（只增新的迁移）
- 禁止在迁移中使用 `DELETE` 或 `UPDATE` 大量数据（用单独脚本）
- 索引创建使用 `CREATE INDEX CONCURRENTLY`（避免锁表）

---

## 9. 备份恢复方案

### 9.1 PostgreSQL 备份

```bash
#!/bin/bash
# scripts/backup/pg_backup.sh
# 每日凌晨 2:00 执行（crontab: 0 2 * * *）

BACKUP_DIR="/var/backups/heka/postgres"
DATE=$(date +%Y%m%d)
RETAIN_DAYS=30

# 全量备份
docker exec heka-postgres pg_dump -U heka -Fc heka > "$BACKUP_DIR/heka_$DATE.dump"

# 压缩
gzip "$BACKUP_DIR/heka_$DATE.dump"

# 清理过期备份
find "$BACKUP_DIR" -name "*.dump.gz" -mtime +$RETAIN_DAYS -delete

echo "[$(date)] PostgreSQL backup completed: heka_$DATE.dump.gz"
```

### 9.2 PostgreSQL 恢复

```bash
# 恢复全量备份
gunzip heka_20250515.dump.gz
docker exec -i heka-postgres pg_restore -U heka -d heka -c < heka_20250515.dump

# 恢复到新数据库（不影响生产）
docker exec -i heka-postgres createdb -U heka heka_restore
docker exec -i heka-postgres pg_restore -U heka -d heka_restore < heka_20250515.dump
```

### 9.3 Redis 备份

```bash
# Redis 使用 AOF 持久化（已配置 appendonly yes）
# 备份 AOF 文件
cp /var/lib/docker/volumes/heka_redis-data/_data/appendonly.aof \
   /var/backups/heka/redis/appendonly_$DATE.aof
```

### 9.4 Milvus 备份

```bash
# 使用 milvus-backup 工具
# 安装：https://github.com/zilliztech/milvus-backup

# 备份
./milvus-backup create -n backup_$DATE

# 恢复
./milvus-backup restore -n backup_$DATE
```

### 9.5 文件存储备份

```bash
# 增量备份上传文件
rsync -av --delete /var/heka/uploads/ /var/backups/heka/uploads/
```

### 9.6 恢复演练

- **频率**：每季度一次
- **流程**：
  1. 在测试环境恢复最新备份
  2. 验证数据完整性（表数量、关键记录数）
  3. 验证应用可正常启动和运行
  4. 记录恢复耗时
- **记录**：演练结果记录到运维文档

---

## 10. Embedding 维度配置化

### 10.1 按模型配置维度

```go
// internal/shared/config/ai.go
package config

// Embedding 模型维度映射
var EmbeddingDimensions = map[string]int{
    "text-embedding-ada-002":     1536,  // OpenAI
    "text-embedding-3-small":     1536,  // OpenAI
    "text-embedding-3-large":     3072,  // OpenAI
    "embedding-3":                1024,  // Cohere
    "all-MiniLM-L6-v2":          384,   // sentence-transformers
    "multilingual-e5-large":     1024,  // multilingual
}

func GetEmbeddingDimension(model string) int {
    if dim, ok := EmbeddingDimensions[model]; ok {
        return dim
    }
    return 1536 // 默认
}
```

### 10.2 Milvus Collection 动态创建

```go
// 创建 Collection 时使用配置化的维度
func (r *vectorRepository) EnsureCollection(ctx context.Context, dim int) error {
    // 检查 Collection 是否存在
    if exists, _ := r.client.HasCollection(ctx, collectionName); exists {
        return nil
    }
    
    // 动态创建
    schema := &entity.Schema{
        CollectionName: collectionName,
        Fields: []*entity.Field{
            {Name: "id", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "36"}, PrimaryKey: true},
            {Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": strconv.Itoa(dim)}},
            // ... 其他字段
        },
    }
    
    return r.client.CreateCollection(ctx, schema, 1)
}
```

---

**文档版本**：v1.0
**最后更新**：2025-05-15

**使用说明**：
1. 所有建表操作必须通过本规范检查
2. 所有索引创建必须遵循命名规范
3. 所有查询必须经过性能分析（EXPLAIN ANALYZE）
4. 定期（每月）检查慢查询和未使用索引
