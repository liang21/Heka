CREATE TABLE test_executions (
    id VARCHAR(36) PRIMARY KEY,
    plan_id VARCHAR(36) NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'paused', 'completed', 'cancelled')),
    executor_id VARCHAR(36) NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paused_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    notes TEXT
);

CREATE UNIQUE INDEX idx_executions_single_active ON test_executions(plan_id) WHERE status = 'in_progress';

CREATE TABLE execution_results (
    id VARCHAR(36) PRIMARY KEY,
    execution_id VARCHAR(36) NOT NULL REFERENCES test_executions(id) ON DELETE CASCADE,
    test_case_id VARCHAR(36) NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    executor_id VARCHAR(36) NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed', 'blocked', 'skipped')),
    bug_id VARCHAR(100),
    bug_url VARCHAR(500),
    notes TEXT,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(execution_id, test_case_id)
);

CREATE INDEX idx_execution_results_execution ON execution_results(execution_id);

ALTER TABLE test_plans ADD CONSTRAINT fk_current_execution FOREIGN KEY(current_execution_id) REFERENCES test_executions(id);
