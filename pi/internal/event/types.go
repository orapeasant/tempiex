package event

import "time"

type EventType string

const (
	EventExecutionStart  EventType = "execution_start"
	EventExecutionUpdate EventType = "execution_update"
	EventExecutionEnd    EventType = "execution_end"
	EventExecutionError  EventType = "execution_error"
)

type Event struct {
	ID         string
	Type       EventType
	SessionID  string
	RunID      string
	WorkflowID string
	ToolName   string
	ToolCallID string
	Timestamp  time.Time
	Payload    any
	Error      string
}
