package persistence

import (
	"context"
	"time"
)

// Namespace represents a namespace record.
type Namespace struct {
	ID            string
	Name          string
	Description   string
	RetentionDays int32
	CreatedAt     time.Time
}

// WorkflowExecution represents a workflow execution record.
type WorkflowExecution struct {
	NamespaceID  string
	WorkflowID   string
	RunID        string
	WorkflowType string
	TaskQueue    string
	Status       int32
	Input        []byte
	Result       []byte
	Failure      []byte
	StartedAt    time.Time
	ClosedAt     *time.Time
}

// HistoryEventRow represents a stored history event.
type HistoryEventRow struct {
	NamespaceID string
	WorkflowID  string
	RunID       string
	EventID     int64
	EventType   int32
	EventData   []byte
	CreatedAt   time.Time
}

// WorkflowTask represents a workflow task.
type WorkflowTask struct {
	ID               int64
	NamespaceID      string
	WorkflowID       string
	RunID            string
	TaskQueue        string
	ScheduledEventID int64
	CreatedAt        time.Time
}

// ActivityTask represents an activity task.
type ActivityTask struct {
	ID               int64
	NamespaceID      string
	WorkflowID       string
	RunID            string
	TaskQueue        string
	ActivityID       string
	ActivityType     string
	ScheduledEventID int64
	Input            []byte
	CreatedAt        time.Time
}

// Store defines the persistence interface.
type Store interface {
	// Schema
	InitSchema(ctx context.Context) error

	// Namespaces
	CreateNamespace(ctx context.Context, ns *Namespace) error
	GetNamespace(ctx context.Context, name string) (*Namespace, error)

	// Workflow executions
	CreateWorkflowExecution(ctx context.Context, wf *WorkflowExecution) error
	GetWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID string) (*WorkflowExecution, error)
	UpdateWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID string, status int32, result, failure []byte, closedAt *time.Time) error
	UpdateWorkflowExecutionStatus(ctx context.Context, namespaceID, workflowID, runID string, status int32) error
	CountWorkflowExecutions(ctx context.Context, namespaceID string) (int64, error)
	ListWorkflowExecutions(ctx context.Context, namespaceID string, pageSize int32, pageToken []byte) ([]*WorkflowExecution, []byte, error)

	// History events
	AppendHistoryEvent(ctx context.Context, event *HistoryEventRow) error
	GetHistoryEvents(ctx context.Context, namespaceID, workflowID, runID string) ([]*HistoryEventRow, error)
	NextEventID(ctx context.Context, namespaceID, workflowID, runID string) (int64, error)

	// Workflow tasks
	CreateWorkflowTask(ctx context.Context, task *WorkflowTask) error
	ClaimWorkflowTask(ctx context.Context, namespaceID, taskQueue string) (*WorkflowTask, error)
	DeleteWorkflowTask(ctx context.Context, id int64) error

	// Activity tasks
	CreateActivityTask(ctx context.Context, task *ActivityTask) error
	ClaimActivityTask(ctx context.Context, namespaceID, taskQueue string) (*ActivityTask, error)
	DeleteActivityTask(ctx context.Context, id int64) error
}
