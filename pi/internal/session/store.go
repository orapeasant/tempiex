package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			run_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
	`); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveSession(ctx context.Context, sess *Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sessions (id, workflow_id, run_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.WorkflowID, sess.RunID, string(sess.Status),
		sess.CreatedAt.UnixMilli(), sess.UpdatedAt.UnixMilli(),
	)
	return err
}

func (s *Store) LoadSession(ctx context.Context, id string) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, workflow_id, run_id, status, created_at, updated_at FROM sessions WHERE id = ?`, id)
	return scanSession(row)
}

func (s *Store) ListSessions(ctx context.Context) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, workflow_id, run_id, status, created_at, updated_at FROM sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *sess)
	}
	return sessions, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSession(s scanner) (*Session, error) {
	var sess Session
	var createdMs, updatedMs int64
	var status string
	if err := s.Scan(&sess.ID, &sess.WorkflowID, &sess.RunID, &status, &createdMs, &updatedMs); err != nil {
		return nil, err
	}
	sess.Status = SessionStatus(status)
	sess.CreatedAt = time.UnixMilli(createdMs).UTC()
	sess.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &sess, nil
}

func (s *Store) AppendEvent(ctx context.Context, e *StoredEvent) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO events (id, session_id, data, created_at) VALUES (?, ?, ?, ?)`,
		e.ID, e.SessionID, string(data), e.Timestamp.UnixMilli(),
	)
	return err
}

func (s *Store) LoadEvents(ctx context.Context, sessionID string) ([]StoredEvent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT data FROM events WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []StoredEvent
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var e StoredEvent
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
