package history

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tempiex/tempiex/internal/matching"
	"github.com/tempiex/tempiex/internal/persistence"
	enumsv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/enums/v1"
	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	failurev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/failure/v1"
	historyv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/history/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskToken is the JSON-encoded task token sent to workers.
type TaskToken struct {
	NamespaceID      string `json:"namespace_id"`
	WorkflowID       string `json:"workflow_id"`
	RunID            string `json:"run_id"`
	ScheduledEventID int64  `json:"scheduled_event_id"`
	Type             string `json:"type"` // "workflow" or "activity"
	ActivityID       string `json:"activity_id,omitempty"`
}

// Command represents a decoded command from a workflow task response.
type Command struct {
	Type string

	// ScheduleActivityTask
	ActivityID         string
	ActivityType       string
	ActivityTaskQueue  string
	ActivityInput      []byte
	StartToCloseTimeout time.Duration

	// CompleteWorkflow
	Result []byte

	// FailWorkflow
	Failure *Failure
}

// Failure represents a workflow/activity failure.
type Failure struct {
	Message    string
	Type       string
	StackTrace string
}

// StartRequest is the input for starting a workflow.
type StartRequest struct {
	NamespaceID  string
	WorkflowID   string
	RunID        string
	WorkflowType string
	TaskQueue    string
	Input        []byte
}

// Engine manages the workflow lifecycle.
type Engine struct {
	store    persistence.Store
	matching *matching.Engine
}

// NewEngine creates a new history engine.
func NewEngine(store persistence.Store, matching *matching.Engine) *Engine {
	return &Engine{store: store, matching: matching}
}

func (e *Engine) appendEvent(ctx context.Context, namespaceID, workflowID, runID string, event *historyv1.HistoryEvent) (int64, error) {
	nextID, err := e.store.NextEventID(ctx, namespaceID, workflowID, runID)
	if err != nil {
		return 0, err
	}
	event.EventId = nextID
	event.EventTime = timestamppb.Now()

	data, err := proto.Marshal(event)
	if err != nil {
		return 0, err
	}

	return nextID, e.store.AppendHistoryEvent(ctx, &persistence.HistoryEventRow{
		NamespaceID: namespaceID,
		WorkflowID:  workflowID,
		RunID:       runID,
		EventID:     nextID,
		EventType:   int32(event.EventType),
		EventData:   data,
	})
}

// StartWorkflowExecution starts a new workflow execution.
func (e *Engine) StartWorkflowExecution(ctx context.Context, req *StartRequest) (string, error) {
	if err := e.store.CreateWorkflowExecution(ctx, &persistence.WorkflowExecution{
		NamespaceID:  req.NamespaceID,
		WorkflowID:   req.WorkflowID,
		RunID:        req.RunID,
		WorkflowType: req.WorkflowType,
		TaskQueue:    req.TaskQueue,
		Status:       int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_RUNNING),
		Input:        req.Input,
	}); err != nil {
		return "", fmt.Errorf("create workflow execution: %w", err)
	}

	startedID, err := e.appendEvent(ctx, req.NamespaceID, req.WorkflowID, req.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &historyv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &historyv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &commonv1.WorkflowType{Name: req.WorkflowType},
				TaskQueue:    req.TaskQueue,
				WorkflowId:   req.WorkflowID,
				Input:        unmarshalPayloads(req.Input),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("append started event: %w", err)
	}
	_ = startedID

	scheduledID, err := e.appendEvent(ctx, req.NamespaceID, req.WorkflowID, req.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historyv1.WorkflowTaskScheduledEventAttributes{
				TaskQueue: req.TaskQueue,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("append task scheduled event: %w", err)
	}

	if err := e.matching.AddWorkflowTask(ctx, &matching.WorkflowTask{
		NamespaceID:      req.NamespaceID,
		WorkflowID:       req.WorkflowID,
		RunID:            req.RunID,
		TaskQueue:        req.TaskQueue,
		ScheduledEventID: scheduledID,
	}); err != nil {
		return "", fmt.Errorf("add workflow task: %w", err)
	}

	return req.RunID, nil
}

// RespondWorkflowTaskCompleted processes a workflow task completion.
func (e *Engine) RespondWorkflowTaskCompleted(ctx context.Context, token TaskToken, commands []*Command, startedEventID int64) error {
	completedID, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_COMPLETED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskCompletedEventAttributes{
			WorkflowTaskCompletedEventAttributes: &historyv1.WorkflowTaskCompletedEventAttributes{
				ScheduledEventId: token.ScheduledEventID,
				StartedEventId:   startedEventID,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("append task completed event: %w", err)
	}

	for _, cmd := range commands {
		switch cmd.Type {
		case "schedule_activity":
			actScheduledID, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
				EventType: enumsv1.EventType_EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
				Attributes: &historyv1.HistoryEvent_ActivityTaskScheduledEventAttributes{
					ActivityTaskScheduledEventAttributes: &historyv1.ActivityTaskScheduledEventAttributes{
						ActivityId:                   cmd.ActivityID,
						ActivityType:                 &commonv1.ActivityType{Name: cmd.ActivityType},
						TaskQueue:                    cmd.ActivityTaskQueue,
						WorkflowTaskCompletedEventId: completedID,
						StartToCloseTimeout:          durationpb.New(cmd.StartToCloseTimeout),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("append activity scheduled event: %w", err)
			}

			inputBytes, err := marshalPayloads(cmd.ActivityInput)
			if err != nil {
				return err
			}
			if err := e.matching.AddActivityTask(ctx, &matching.ActivityTask{
				NamespaceID:      token.NamespaceID,
				WorkflowID:       token.WorkflowID,
				RunID:            token.RunID,
				TaskQueue:        cmd.ActivityTaskQueue,
				ActivityID:       cmd.ActivityID,
				ActivityType:     cmd.ActivityType,
				ScheduledEventID: actScheduledID,
				Input:            inputBytes,
			}); err != nil {
				return fmt.Errorf("add activity task: %w", err)
			}

		case "complete_workflow":
			resultBytes, err := marshalPayloads(cmd.Result)
			if err != nil {
				return err
			}
			var resultProto *commonv1.Payloads
			if len(cmd.Result) > 0 {
				resultProto = &commonv1.Payloads{}
				if err := proto.Unmarshal(cmd.Result, resultProto); err != nil {
					return fmt.Errorf("unmarshal workflow result: %w", err)
				}
			}
			if _, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
				EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
				Attributes: &historyv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
					WorkflowExecutionCompletedEventAttributes: &historyv1.WorkflowExecutionCompletedEventAttributes{
						WorkflowTaskCompletedEventId: completedID,
						Result:                       resultProto,
					},
				},
			}); err != nil {
				return err
			}
			now := time.Now()
			if err := e.store.UpdateWorkflowExecution(ctx, token.NamespaceID, token.WorkflowID, token.RunID,
				int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_COMPLETED),
				resultBytes, nil, &now); err != nil {
				return err
			}

		case "fail_workflow":
			if _, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
				EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
				Attributes: &historyv1.HistoryEvent_WorkflowExecutionFailedEventAttributes{
					WorkflowExecutionFailedEventAttributes: &historyv1.WorkflowExecutionFailedEventAttributes{
						WorkflowTaskCompletedEventId: completedID,
						Failure: &failurev1.Failure{
							Message:    cmd.Failure.Message,
							Type:       cmd.Failure.Type,
							StackTrace: cmd.Failure.StackTrace,
						},
					},
				},
			}); err != nil {
				return err
			}
			failureBytes, _ := json.Marshal(cmd.Failure)
			now := time.Now()
			if err := e.store.UpdateWorkflowExecution(ctx, token.NamespaceID, token.WorkflowID, token.RunID,
				int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_FAILED),
				nil, failureBytes, &now); err != nil {
				return err
			}
		}
	}
	return nil
}

// RespondActivityTaskCompleted processes an activity task completion.
func (e *Engine) RespondActivityTaskCompleted(ctx context.Context, token TaskToken, resultPayloads []byte) error {
	startedEventID, err := e.getActivityStartedEventID(ctx, token)
	if err != nil {
		return err
	}

	var resultProto *commonv1.Payloads
	if len(resultPayloads) > 0 {
		resultProto = &commonv1.Payloads{}
		if err := proto.Unmarshal(resultPayloads, resultProto); err != nil {
			return fmt.Errorf("unmarshal activity result: %w", err)
		}
	}

	if _, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
		Attributes: &historyv1.HistoryEvent_ActivityTaskCompletedEventAttributes{
			ActivityTaskCompletedEventAttributes: &historyv1.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: token.ScheduledEventID,
				StartedEventId:   startedEventID,
				Result:           resultProto,
			},
		},
	}); err != nil {
		return err
	}

	// Get workflow execution to find task queue
	wf, err := e.store.GetWorkflowExecution(ctx, token.NamespaceID, token.WorkflowID, token.RunID)
	if err != nil || wf == nil {
		return fmt.Errorf("get workflow execution: %w", err)
	}

	scheduledID, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historyv1.WorkflowTaskScheduledEventAttributes{
				TaskQueue: wf.TaskQueue,
			},
		},
	})
	if err != nil {
		return err
	}

	return e.matching.AddWorkflowTask(ctx, &matching.WorkflowTask{
		NamespaceID:      token.NamespaceID,
		WorkflowID:       token.WorkflowID,
		RunID:            token.RunID,
		TaskQueue:        wf.TaskQueue,
		ScheduledEventID: scheduledID,
	})
}

// RespondActivityTaskFailed processes an activity task failure.
func (e *Engine) RespondActivityTaskFailed(ctx context.Context, token TaskToken, failure *Failure) error {
	startedEventID, err := e.getActivityStartedEventID(ctx, token)
	if err != nil {
		return err
	}

	if _, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_ACTIVITY_TASK_FAILED,
		Attributes: &historyv1.HistoryEvent_ActivityTaskFailedEventAttributes{
			ActivityTaskFailedEventAttributes: &historyv1.ActivityTaskFailedEventAttributes{
				ScheduledEventId: token.ScheduledEventID,
				StartedEventId:   startedEventID,
				Failure: &failurev1.Failure{
					Message:    failure.Message,
					Type:       failure.Type,
					StackTrace: failure.StackTrace,
				},
			},
		},
	}); err != nil {
		return err
	}

	wf, err := e.store.GetWorkflowExecution(ctx, token.NamespaceID, token.WorkflowID, token.RunID)
	if err != nil || wf == nil {
		return fmt.Errorf("get workflow execution: %w", err)
	}

	scheduledID, err := e.appendEvent(ctx, token.NamespaceID, token.WorkflowID, token.RunID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historyv1.WorkflowTaskScheduledEventAttributes{
				TaskQueue: wf.TaskQueue,
			},
		},
	})
	if err != nil {
		return err
	}

	return e.matching.AddWorkflowTask(ctx, &matching.WorkflowTask{
		NamespaceID:      token.NamespaceID,
		WorkflowID:       token.WorkflowID,
		RunID:            token.RunID,
		TaskQueue:        wf.TaskQueue,
		ScheduledEventID: scheduledID,
	})
}

// AppendActivityTaskStarted appends ActivityTaskStarted event and returns the started event ID.
func (e *Engine) AppendActivityTaskStarted(ctx context.Context, namespaceID, workflowID, runID string, scheduledEventID int64) (int64, error) {
	return e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &historyv1.HistoryEvent_ActivityTaskStartedEventAttributes{
			ActivityTaskStartedEventAttributes: &historyv1.ActivityTaskStartedEventAttributes{
				ScheduledEventId: scheduledEventID,
			},
		},
	})
}

// AppendWorkflowTaskStarted appends WorkflowTaskStarted event and returns the started event ID.
func (e *Engine) AppendWorkflowTaskStarted(ctx context.Context, namespaceID, workflowID, runID string, scheduledEventID int64) (int64, error) {
	return e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_STARTED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskStartedEventAttributes{
			WorkflowTaskStartedEventAttributes: &historyv1.WorkflowTaskStartedEventAttributes{
				ScheduledEventId: scheduledEventID,
			},
		},
	})
}

func (e *Engine) getActivityStartedEventID(ctx context.Context, token TaskToken) (int64, error) {
	events, err := e.store.GetHistoryEvents(ctx, token.NamespaceID, token.WorkflowID, token.RunID)
	if err != nil {
		return 0, err
	}
	for _, ev := range events {
		if int32(enumsv1.EventType_EVENT_TYPE_ACTIVITY_TASK_STARTED) == ev.EventType {
			evProto := &historyv1.HistoryEvent{}
			if err := proto.Unmarshal(ev.EventData, evProto); err != nil {
				continue
			}
			attrs := evProto.GetActivityTaskStartedEventAttributes()
			if attrs != nil && attrs.ScheduledEventId == token.ScheduledEventID {
				return ev.EventID, nil
			}
		}
	}
	return 0, nil
}

// marshalPayloads is a helper to pass raw bytes as-is (already serialized payloads).
func marshalPayloads(data []byte) ([]byte, error) {
	return data, nil
}

// unmarshalPayloads decodes raw bytes into a Payloads proto, returning nil on empty input.
func unmarshalPayloads(data []byte) *commonv1.Payloads {
	if len(data) == 0 {
		return nil
	}
	p := &commonv1.Payloads{}
	if err := proto.Unmarshal(data, p); err != nil {
		return nil
	}
	return p
}

// SignalWorkflowExecution appends a signal event and schedules a new workflow task.
func (e *Engine) SignalWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID, signalName string, input []byte) error {
	wf, err := e.store.GetWorkflowExecution(ctx, namespaceID, workflowID, runID)
	if err != nil {
		return fmt.Errorf("get workflow: %w", err)
	}
	if wf.Status != int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_RUNNING) {
		return fmt.Errorf("workflow is not running")
	}

	_, err = e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED,
		Attributes: &historyv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
			WorkflowExecutionSignaledEventAttributes: &historyv1.WorkflowExecutionSignaledEventAttributes{
				SignalName: signalName,
				Input:      unmarshalPayloads(input),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("append signal event: %w", err)
	}

	// Schedule a new workflow task so the worker sees the signal
	taskQueue := wf.TaskQueue
	scheduledID, err := e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historyv1.WorkflowTaskScheduledEventAttributes{
				TaskQueue: taskQueue,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("append workflow task scheduled: %w", err)
	}

	if err := e.matching.AddWorkflowTask(ctx, &matching.WorkflowTask{
		NamespaceID:      namespaceID,
		WorkflowID:       workflowID,
		RunID:            runID,
		TaskQueue:        taskQueue,
		ScheduledEventID: scheduledID,
	}); err != nil {
		return fmt.Errorf("add workflow task: %w", err)
	}

	return nil
}

// RequestCancelWorkflowExecution records a cancellation request event.
func (e *Engine) RequestCancelWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID, reason string) error {
	wf, err := e.store.GetWorkflowExecution(ctx, namespaceID, workflowID, runID)
	if err != nil {
		return fmt.Errorf("get workflow: %w", err)
	}
	if wf.Status != int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_RUNNING) {
		return fmt.Errorf("workflow is not running")
	}

	_, err = e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED,
		Attributes: &historyv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes{
			WorkflowExecutionCancelRequestedEventAttributes: &historyv1.WorkflowExecutionCancelRequestedEventAttributes{},
		},
	})
	if err != nil {
		return fmt.Errorf("append cancel requested event: %w", err)
	}

	taskQueue := wf.TaskQueue
	scheduledID, err := e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_SCHEDULED,
		Attributes: &historyv1.HistoryEvent_WorkflowTaskScheduledEventAttributes{
			WorkflowTaskScheduledEventAttributes: &historyv1.WorkflowTaskScheduledEventAttributes{
				TaskQueue: taskQueue,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("append workflow task scheduled: %w", err)
	}

	return e.matching.AddWorkflowTask(ctx, &matching.WorkflowTask{
		NamespaceID:      namespaceID,
		WorkflowID:       workflowID,
		RunID:            runID,
		TaskQueue:        taskQueue,
		ScheduledEventID: scheduledID,
	})
}

// TerminateWorkflowExecution forcefully closes a workflow execution.
func (e *Engine) TerminateWorkflowExecution(ctx context.Context, namespaceID, workflowID, runID, reason string) error {
	wf, err := e.store.GetWorkflowExecution(ctx, namespaceID, workflowID, runID)
	if err != nil {
		return fmt.Errorf("get workflow: %w", err)
	}
	if wf.Status != int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_RUNNING) {
		return nil // already closed
	}

	_, err = e.appendEvent(ctx, namespaceID, workflowID, runID, &historyv1.HistoryEvent{
		EventType: enumsv1.EventType_EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED,
		Attributes: &historyv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes{
			WorkflowExecutionTerminatedEventAttributes: &historyv1.WorkflowExecutionTerminatedEventAttributes{
				Reason: reason,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("append terminated event: %w", err)
	}

	return e.store.UpdateWorkflowExecutionStatus(ctx, namespaceID, workflowID, runID,
		int32(enumsv1.WorkflowExecutionStatus_WORKFLOW_EXECUTION_STATUS_TERMINATED))
}

// CountWorkflowExecutions counts workflows in a namespace matching an optional query.
func (e *Engine) CountWorkflowExecutions(ctx context.Context, namespaceID string) (int64, error) {
	return e.store.CountWorkflowExecutions(ctx, namespaceID)
}

// ListWorkflowExecutions lists workflows in a namespace.
func (e *Engine) ListWorkflowExecutions(ctx context.Context, namespaceID string, pageSize int32, pageToken []byte) ([]*persistence.WorkflowExecution, []byte, error) {
	return e.store.ListWorkflowExecutions(ctx, namespaceID, pageSize, pageToken)
}
