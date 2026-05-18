CREATE TABLE modules (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    parent_id VARCHAR(36) REFERENCES modules(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    created_by VARCHAR(36) NOT NULL REFERENCES users(id),
    UNIQUE(project_id, parent_id, name)
);

CREATE INDEX idx_modules_project ON modules(project_id);
CREATE INDEX idx_modules_parent ON modules(parent_id);
