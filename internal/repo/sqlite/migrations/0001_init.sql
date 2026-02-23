CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('inbox', 'todo', 'doing', 'done', 'deleted')),
    project TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'P2',
    due_at DATETIME NULL,
    done_at DATETIME NULL,
    meta_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS task_tags (
    task_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (task_id, tag_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ai_cache (
    cache_key TEXT PRIMARY KEY,
    response_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project);
CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX IF NOT EXISTS idx_tasks_done_at ON tasks(done_at);
