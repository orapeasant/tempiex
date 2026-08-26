package session

import "time"

type SessionStatus string

const (
	StatusActive    SessionStatus = "active"
	StatusCompleted SessionStatus = "completed"
	StatusFailed    SessionStatus = "failed"
	StatusCancelled SessionStatus = "cancelled"
)

type Session struct {
	ID         string        `json:"id"`
	WorkflowID string        `json:"workflowId"`
	RunID      string        `json:"runId"`
	Status     SessionStatus `json:"status"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}

type StoredEvent struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Type      string    `json:"type"`
	ToolName  string    `json:"toolName"`
	Timestamp time.Time `json:"timestamp"`
	Payload   []byte    `json:"payload,omitempty"`
	Error     string    `json:"error,omitempty"`
}
