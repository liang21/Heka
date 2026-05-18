CREATE TABLE tags (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#000000',
    created_by VARCHAR(36) NOT NULL REFERENCES users(id),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_tags_project ON tags(project_id);
