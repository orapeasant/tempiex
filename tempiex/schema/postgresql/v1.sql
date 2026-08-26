CREATE TABLE IF NOT EXISTS namespaces (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  retention_days INTEGER NOT NULL DEFAULT 7,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_executions (
  namespace_id TEXT NOT NULL,
  workflow_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  workflow_type TEXT NOT NULL,
  task_queue TEXT NOT NULL,
  status INTEGER NOT NULL DEFAULT 1,
  input BYTEA,
  result BYTEA,
  failure BYTEA,
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  closed_at TIMESTAMPTZ,
  PRIMARY KEY (namespace_id, workflow_id, run_id)
);

CREATE TABLE IF NOT EXISTS history_events (
  namespace_id TEXT NOT NULL,
  workflow_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  event_id BIGINT NOT NULL,
  event_type INTEGER NOT NULL,
  event_data BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (namespace_id, workflow_id, run_id, event_id)
);

CREATE TABLE IF NOT EXISTS workflow_tasks (
  id BIGSERIAL PRIMARY KEY,
  namespace_id TEXT NOT NULL,
  workflow_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  task_queue TEXT NOT NULL,
  scheduled_event_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS activity_tasks (
  id BIGSERIAL PRIMARY KEY,
  namespace_id TEXT NOT NULL,
  workflow_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  task_queue TEXT NOT NULL,
  activity_id TEXT NOT NULL,
  activity_type TEXT NOT NULL,
  scheduled_event_id BIGINT NOT NULL,
  input BYTEA,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
