CREATE TABLE IF NOT EXISTS devices (
    device_id   TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
    rel_path     TEXT PRIMARY KEY,
    content_hash TEXT NOT NULL,
    size         INTEGER NOT NULL,
    mod_time     TEXT NOT NULL,
    state        INTEGER NOT NULL DEFAULT 0,
    device_id    TEXT NOT NULL,
    last_sync_at TEXT NOT NULL,
    FOREIGN KEY (device_id) REFERENCES devices(device_id)
);

CREATE TABLE IF NOT EXISTS sync_log (
    sync_id        TEXT PRIMARY KEY,
    device_id      TEXT NOT NULL,
    created_at     TEXT NOT NULL,
    parent_sync_id TEXT,
    file_count     INTEGER NOT NULL,
    direction      TEXT NOT NULL,
    FOREIGN KEY (device_id) REFERENCES devices(device_id)
);

CREATE TABLE IF NOT EXISTS conflicts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    rel_path      TEXT NOT NULL,
    local_hash    TEXT NOT NULL,
    remote_hash   TEXT NOT NULL,
    local_device  TEXT NOT NULL,
    remote_device TEXT NOT NULL,
    detected_at   TEXT NOT NULL,
    resolved      INTEGER NOT NULL DEFAULT 0,
    resolution    TEXT
);

CREATE TABLE IF NOT EXISTS tombstones (
    rel_path   TEXT NOT NULL,
    deleted_by TEXT NOT NULL,
    deleted_at TEXT NOT NULL,
    sync_id    TEXT,
    PRIMARY KEY (rel_path, deleted_by)
);

CREATE INDEX IF NOT EXISTS idx_files_state ON files(state);
CREATE INDEX IF NOT EXISTS idx_conflicts_unresolved ON conflicts(resolved) WHERE resolved = 0;
CREATE INDEX IF NOT EXISTS idx_tombstones_path ON tombstones(rel_path);
