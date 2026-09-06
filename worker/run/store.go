package run

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store persists runs and their events in SQLite.
type Store struct {
	db *sql.DB
}

// Open opens (and if needed creates) the run database at path. Use ":memory:"
// for an ephemeral store.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			namespace TEXT NOT NULL,
			task_queue TEXT NOT NULL,
			workflow_id TEXT NOT NULL,
			workflow_run_id TEXT NOT NULL,
			activity_id TEXT NOT NULL,
			activity_type TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			ended_at INTEGER
		);
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_events_run ON events (run_id, created_at);
	`); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) SaveRun(ctx context.Context, r *Run) error {
	var endedAt any
	if r.EndedAt != nil {
		endedAt = r.EndedAt.UnixMilli()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO runs
		   (id, namespace, task_queue, workflow_id, workflow_run_id, activity_id, activity_type, status, started_at, ended_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Namespace, r.TaskQueue, r.WorkflowID, r.WorkflowRunID,
		r.ActivityID, r.ActivityType, string(r.Status), r.StartedAt.UnixMilli(), endedAt,
	)
	return err
}

const runColumns = `id, namespace, task_queue, workflow_id, workflow_run_id, activity_id, activity_type, status, started_at, ended_at`

func (s *Store) LoadRun(ctx context.Context, id string) (*Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+runColumns+` FROM runs WHERE id = ?`, id)
	return scanRun(row)
}

func (s *Store) ListRuns(ctx context.Context) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+runColumns+` FROM runs ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *r)
	}
	return runs, rows.Err()
}

// PruneCompleted keeps at most keep finished runs (completed, failed, or
// cancelled), deleting the oldest beyond that along with their events. Active
// runs are never pruned. keep <= 0 disables pruning.
func (s *Store) PruneCompleted(ctx context.Context, keep int) error {
	if keep <= 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM events WHERE run_id IN (
		    SELECT id FROM runs WHERE status != ?
		    ORDER BY started_at DESC LIMIT -1 OFFSET ?
		 )`, string(StatusActive), keep)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`DELETE FROM runs WHERE id IN (
		    SELECT id FROM runs WHERE status != ?
		    ORDER BY started_at DESC LIMIT -1 OFFSET ?
		 )`, string(StatusActive), keep)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRun(s scanner) (*Run, error) {
	var r Run
	var status string
	var startedMs int64
	var endedMs sql.NullInt64
	if err := s.Scan(&r.ID, &r.Namespace, &r.TaskQueue, &r.WorkflowID, &r.WorkflowRunID,
		&r.ActivityID, &r.ActivityType, &status, &startedMs, &endedMs); err != nil {
		return nil, err
	}
	r.Status = Status(status)
	r.StartedAt = time.UnixMilli(startedMs).UTC()
	if endedMs.Valid {
		t := time.UnixMilli(endedMs.Int64).UTC()
		r.EndedAt = &t
	}
	return &r, nil
}

func (s *Store) AppendEvent(ctx context.Context, e *Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO events (id, run_id, data, created_at) VALUES (?, ?, ?, ?)`,
		e.ID, e.RunID, string(data), e.Timestamp.UnixMilli(),
	)
	return err
}

func (s *Store) LoadEvents(ctx context.Context, runID string) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT data FROM events WHERE run_id = ? ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var e Event
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
