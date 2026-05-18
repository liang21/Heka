CREATE TABLE test_plans (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'cancelled')),
    current_execution_id VARCHAR(36), -- logical FK to test_executions, added later
    started_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_by VARCHAR(36) NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_test_plans_project ON test_plans(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_plans_status ON test_plans(project_id, status) WHERE deleted_at IS NULL;

CREATE TABLE plan_test_cases (
    plan_id VARCHAR(36) NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
    test_case_id VARCHAR(36) NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    assigned_to VARCHAR(36) REFERENCES users(id),
    order_index INT NOT NULL DEFAULT 0,
    PRIMARY KEY (plan_id, test_case_id)
);
