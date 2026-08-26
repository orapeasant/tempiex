package postgresql

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tempiex/tempiex/internal/config"
	"github.com/tempiex/tempiex/internal/persistence"
)

//go:embed schema.sql
var schemaSql string

// Store implements persistence.Store using PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new PostgreSQL store.
func NewStore(cfg *config.Config) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) InitSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schemaSql)
	return err
}

// CreateNamespace inserts a new namespace.
func (s *Store) CreateNamespace(ctx context.Context, ns *persistence.Namespace) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO namespaces (id, name, description, retention_days) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (name) DO NOTHING`,
		ns.ID, ns.Name, ns.Description, ns.RetentionDays)
	return err
}

// GetNamespace retrieves a namespace by name.
func (s *Store) GetNamespace(ctx context.Context, name string) (*persistence.Namespace, error) {
	ns := &persistence.Namespace{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, description, retention_days, created_at FROM namespaces WHERE name = $1`,
		name).Scan(&ns.ID, &ns.Name, &ns.Description, &ns.RetentionDays, &ns.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ns, err
}

// CreateWorkflowExecution inserts a new workflow execution.
func (s *Store) CreateWorkflowExecution(ctx context.Context, wf *persistence.WorkflowExecution) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO workflow_executions
		 (namespace_id, workflow_id, run_id, workflow_type, task_queue, status, input)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		wf.NamespaceID, wf.WorkflowID, wf.RunID, wf.WorkflowType, wf.TaskQueue, wf.Status, wf.Input)
	return err
}

// GetWorkflowExecution retrieves a workflow execution.
func (s *Store) GetWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID string) (*persistence.WorkflowExecution, error) {
	wf := &persistence.WorkflowExecution{}
	err := s.pool.QueryRow(ctx,
		`SELECT namespace_id, workflow_id, run_id, workflow_type, task_queue, status, input, result, failure, started_at, closed_at
		 FROM workflow_executions WHERE namespace_id=$1 AND workflow_id=$2 AND run_id=$3`,
		namespaceID, workflowID, runID).Scan(
		&wf.NamespaceID, &wf.WorkflowID, &wf.RunID, &wf.WorkflowType, &wf.TaskQueue,
		&wf.Status, &wf.Input, &wf.Result, &wf.Failure, &wf.StartedAt, &wf.ClosedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return wf, err
}

// UpdateWorkflowExecution updates a workflow execution status.
func (s *Store) UpdateWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID string, status int32, result, failure []byte, closedAt *time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE workflow_executions SET status=$4, result=$5, failure=$6, closed_at=$7
		 WHERE namespace_id=$1 AND workflow_id=$2 AND run_id=$3`,
		namespaceID, workflowID, runID, status, result, failure, closedAt)
	return err
}

// NextEventID returns the next event ID for a workflow execution.
func (s *Store) NextEventID(ctx context.Context, namespaceID, workflowID, runID string) (int64, error) {
	var maxID *int64
	err := s.pool.QueryRow(ctx,
		`SELECT MAX(event_id) FROM history_events WHERE namespace_id=$1 AND workflow_id=$2 AND run_id=$3`,
		namespaceID, workflowID, runID).Scan(&maxID)
	if err != nil {
		return 0, err
	}
	if maxID == nil {
		return 1, nil
	}
	return *maxID + 1, nil
}

// AppendHistoryEvent appends a history event.
func (s *Store) AppendHistoryEvent(ctx context.Context, event *persistence.HistoryEventRow) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO history_events (namespace_id, workflow_id, run_id, event_id, event_type, event_data)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		event.NamespaceID, event.WorkflowID, event.RunID, event.EventID, event.EventType, event.EventData)
	return err
}

// GetHistoryEvents retrieves all history events for a workflow execution.
func (s *Store) GetHistoryEvents(ctx context.Context, namespaceID, workflowID, runID string) ([]*persistence.HistoryEventRow, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT namespace_id, workflow_id, run_id, event_id, event_type, event_data, created_at
		 FROM history_events WHERE namespace_id=$1 AND workflow_id=$2 AND run_id=$3
		 ORDER BY event_id`,
		namespaceID, workflowID, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*persistence.HistoryEventRow
	for rows.Next() {
		e := &persistence.HistoryEventRow{}
		if err := rows.Scan(&e.NamespaceID, &e.WorkflowID, &e.RunID, &e.EventID, &e.EventType, &e.EventData, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// CreateWorkflowTask creates a workflow task.
func (s *Store) CreateWorkflowTask(ctx context.Context, task *persistence.WorkflowTask) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO workflow_tasks (namespace_id, workflow_id, run_id, task_queue, scheduled_event_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		task.NamespaceID, task.WorkflowID, task.RunID, task.TaskQueue, task.ScheduledEventID)
	return err
}

// ClaimWorkflowTask atomically claims a workflow task from the DB.
func (s *Store) ClaimWorkflowTask(ctx context.Context, namespaceID, taskQueue string) (*persistence.WorkflowTask, error) {
	task := &persistence.WorkflowTask{}
	err := s.pool.QueryRow(ctx,
		`DELETE FROM workflow_tasks WHERE id = (
		   SELECT id FROM workflow_tasks
		   WHERE namespace_id=$1 AND task_queue=$2
		   ORDER BY id LIMIT 1
		   FOR UPDATE SKIP LOCKED
		 ) RETURNING id, namespace_id, workflow_id, run_id, task_queue, scheduled_event_id, created_at`,
		namespaceID, taskQueue).Scan(
		&task.ID, &task.NamespaceID, &task.WorkflowID, &task.RunID,
		&task.TaskQueue, &task.ScheduledEventID, &task.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return task, err
}

// DeleteWorkflowTask deletes a workflow task.
func (s *Store) DeleteWorkflowTask(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM workflow_tasks WHERE id=$1`, id)
	return err
}

// CreateActivityTask creates an activity task.
func (s *Store) CreateActivityTask(ctx context.Context, task *persistence.ActivityTask) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO activity_tasks (namespace_id, workflow_id, run_id, task_queue, activity_id, activity_type, scheduled_event_id, input)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		task.NamespaceID, task.WorkflowID, task.RunID, task.TaskQueue,
		task.ActivityID, task.ActivityType, task.ScheduledEventID, task.Input)
	return err
}

// ClaimActivityTask atomically claims an activity task from the DB.
func (s *Store) ClaimActivityTask(ctx context.Context, namespaceID, taskQueue string) (*persistence.ActivityTask, error) {
	task := &persistence.ActivityTask{}
	err := s.pool.QueryRow(ctx,
		`DELETE FROM activity_tasks WHERE id = (
		   SELECT id FROM activity_tasks
		   WHERE namespace_id=$1 AND task_queue=$2
		   ORDER BY id LIMIT 1
		   FOR UPDATE SKIP LOCKED
		 ) RETURNING id, namespace_id, workflow_id, run_id, task_queue, activity_id, activity_type, scheduled_event_id, input, created_at`,
		namespaceID, taskQueue).Scan(
		&task.ID, &task.NamespaceID, &task.WorkflowID, &task.RunID,
		&task.TaskQueue, &task.ActivityID, &task.ActivityType,
		&task.ScheduledEventID, &task.Input, &task.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return task, err
}

// DeleteActivityTask deletes an activity task.
func (s *Store) DeleteActivityTask(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM activity_tasks WHERE id=$1`, id)
	return err
}

// UpdateWorkflowExecutionStatus updates only the status field.
func (s *Store) UpdateWorkflowExecutionStatus(ctx context.Context, namespaceID, workflowID, runID string, status int32) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE workflow_executions SET status=$1 WHERE namespace_id=$2 AND workflow_id=$3 AND run_id=$4`,
		status, namespaceID, workflowID, runID)
	return err
}

// CountWorkflowExecutions returns the number of workflow executions in a namespace.
func (s *Store) CountWorkflowExecutions(ctx context.Context, namespaceID string) (int64, error) {
	var count int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_executions WHERE namespace_id=$1`, namespaceID).Scan(&count)
	return count, err
}

// ListWorkflowExecutions lists workflow executions with cursor-based pagination.
func (s *Store) ListWorkflowExecutions(ctx context.Context, namespaceID string, pageSize int32, pageToken []byte) ([]*persistence.WorkflowExecution, []byte, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := int64(0)
	if len(pageToken) > 0 {
		_ = json.Unmarshal(pageToken, &offset)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT namespace_id, workflow_id, run_id, workflow_type, task_queue, status, input, result, failure, started_at, closed_at
		 FROM workflow_executions
		 WHERE namespace_id=$1
		 ORDER BY started_at DESC
		 LIMIT $2 OFFSET $3`,
		namespaceID, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*persistence.WorkflowExecution
	for rows.Next() {
		wf := &persistence.WorkflowExecution{}
		if err := rows.Scan(
			&wf.NamespaceID, &wf.WorkflowID, &wf.RunID, &wf.WorkflowType,
			&wf.TaskQueue, &wf.Status, &wf.Input, &wf.Result, &wf.Failure,
			&wf.StartedAt, &wf.ClosedAt,
		); err != nil {
			return nil, nil, err
		}
		results = append(results, wf)
	}

	var nextToken []byte
	if int32(len(results)) == pageSize {
		nextOffset := offset + int64(pageSize)
		nextToken, _ = json.Marshal(nextOffset)
	}

	return results, nextToken, nil
}
