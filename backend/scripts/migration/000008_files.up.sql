CREATE TABLE files (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    path VARCHAR(1000) NOT NULL,
    source_type VARCHAR(20) NOT NULL DEFAULT 'upload',
    source_url VARCHAR(1000),
    content_preview TEXT,
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE,
    index_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (index_status IN ('pending', 'processing', 'completed', 'failed')),
    index_error TEXT,
    indexed_at TIMESTAMPTZ,
    uploaded_by VARCHAR(36) NOT NULL REFERENCES users(id),
    version INT NOT NULL DEFAULT 1,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_files_project ON files(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_status ON files(project_id, index_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_type ON files(project_id, type) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by) WHERE deleted_at IS NULL;

CREATE TABLE file_versions (
    id VARCHAR(36) PRIMARY KEY,
    file_id VARCHAR(36) NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    version INT NOT NULL,
    path VARCHAR(1000) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    uploaded_by VARCHAR(36) NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(file_id, version)
);
