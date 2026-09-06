// Package run tracks activity executions performed by this worker process.
//
// A Run is one activity task the worker dequeued and executed, correlated to
// the Tempiex workflow that scheduled it (namespace, task queue, workflow id,
// workflow run id, activity id). Runs and their events are published on a Bus
// for live consumers (the SSE endpoint) and persisted in a Store so a consumer
// that connects mid-run can replay what it missed.
package run

import "time"

// Status is the lifecycle state of a Run.
type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Run is one activity execution tracked by this worker.
//
// ID is worker-local. WorkflowRunID is the Tempiex run id of the workflow that
// scheduled the activity — the two are distinct and both are needed to
// correlate a worker run back to a workflow execution.
type Run struct {
	ID            string     `json:"id"`
	Namespace     string     `json:"namespace"`
	TaskQueue     string     `json:"taskQueue"`
	WorkflowID    string     `json:"workflowId"`
	WorkflowRunID string     `json:"workflowRunId"`
	ActivityID    string     `json:"activityId"`
	ActivityType  string     `json:"activityType"`
	Status        Status     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
}

// Filter narrows a Run listing. Zero-valued fields are ignored.
type Filter struct {
	Status     Status
	WorkflowID string
	TaskQueue  string
}

func (f Filter) matches(r Run) bool {
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.WorkflowID != "" && r.WorkflowID != f.WorkflowID {
		return false
	}
	if f.TaskQueue != "" && r.TaskQueue != f.TaskQueue {
		return false
	}
	return true
}

// EventType identifies a point in an activity execution's lifecycle.
type EventType string

const (
	EventExecutionStart     EventType = "execution_start"
	EventExecutionUpdate    EventType = "execution_update"
	EventExecutionEnd       EventType = "execution_end"
	EventExecutionError     EventType = "execution_error"
	EventExecutionCancelled EventType = "execution_cancelled"
)

// Event is one structured observation about a Run. Payload is activity-specific
// and is JSON-serialized on the wire; Error is non-empty only on error events.
type Event struct {
	ID            string    `json:"id"`
	Type          EventType `json:"type"`
	RunID         string    `json:"runId"`
	WorkflowID    string    `json:"workflowId,omitempty"`
	WorkflowRunID string    `json:"workflowRunId,omitempty"`
	ActivityID    string    `json:"activityId,omitempty"`
	ActivityType  string    `json:"activityType,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Payload       any       `json:"payload,omitempty"`
	Error         string    `json:"error,omitempty"`
}
