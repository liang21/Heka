CREATE TABLE test_cases (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    module_id VARCHAR(36) REFERENCES modules(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'ready', 'archived')),
    priority INT NOT NULL DEFAULT 1 CHECK (priority BETWEEN 0 AND 3),
    tags TEXT[] DEFAULT '{}',
    created_by VARCHAR(36) NOT NULL REFERENCES users(id),
    updated_by VARCHAR(36) REFERENCES users(id),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_test_cases_project ON test_cases(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_module ON test_cases(module_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_status ON test_cases(project_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_priority ON test_cases(project_id, priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_tags ON test_cases USING GIN(tags) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_fulltext ON test_cases USING GIN(to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(description,''))) WHERE deleted_at IS NULL;
CREATE INDEX idx_test_cases_active ON test_cases(project_id, status, priority, created_at DESC) WHERE deleted_at IS NULL AND status != 'archived';

CREATE TABLE test_steps (
    id VARCHAR(36) PRIMARY KEY,
    test_case_id VARCHAR(36) NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    number INT NOT NULL,
    action TEXT NOT NULL,
    expected TEXT NOT NULL,
    UNIQUE(test_case_id, number)
);

CREATE INDEX idx_test_steps_case ON test_steps(test_case_id);
