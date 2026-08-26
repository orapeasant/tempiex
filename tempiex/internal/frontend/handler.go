package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	enumsv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/enums/v1"
	failurev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/failure/v1"
	historyv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/history/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	"github.com/tempiex/tempiex/internal/history"
	"github.com/tempiex/tempiex/internal/matching"
	"github.com/tempiex/tempiex/internal/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const pollTimeout = 30 * time.Second

// Handler implements the WorkflowService gRPC interface.
type Handler struct {
	workflowservicev1.UnimplementedWorkflowServiceServer
	history  *history.Engine
	matching *matching.Engine
	store    persistence.Store
}

// NewHandler creates a new frontend handler.
func NewHandler(hist *history.Engine, match *matching.Engine, store persistence.Store) *Handler {
	return &Handler{
		history:  hist,
		matching: match,
		store:    store,
	}
}

// resolveNamespace looks up the namespace by name, auto-creating if needed.
func (h *Handler) resolveNamespace(ctx context.Context, name string) (string, error) {
	if name == "" {
		name = "default"
	}
	ns, err := h.store.GetNamespace(ctx, name)
	if err != nil {
		return "", fmt.Errorf("get namespace: %w", err)
	}
	if ns != nil {
		return ns.ID, nil
	}
	// Auto-create namespace
	id := uuid.NewString()
	if err := h.store.CreateNamespace(ctx, &persistence.Namespace{
		ID:            id,
		Name:          name,
		Description:   "auto-created",
		RetentionDays: 7,
	}); err != nil {
		return "", fmt.Errorf("create namespace: %w", err)
	}
	// Re-read to handle race (ON CONFLICT DO NOTHING)
	ns, err = h.store.GetNamespace(ctx, name)
	if err != nil || ns == nil {
		return id, err
	}
	return ns.ID, nil
}

func encodeTaskToken(token history.TaskToken) ([]byte, error) {
	return json.Marshal(token)
}

func decodeTaskToken(data []byte) (history.TaskToken, error) {
	var token history.TaskToken
	err := json.Unmarshal(data, &token)
	return token, err
}

// RegisterNamespace registers a new namespace.
func (h *Handler) RegisterNamespace(ctx context.Context, req *workflowservicev1.RegisterNamespaceRequest) (*workflowservicev1.RegisterNamespaceResponse, error) {
	id := uuid.NewString()
	retentionDays := req.RetentionDays
	if retentionDays == 0 {
		retentionDays = 7
	}
	if err := h.store.CreateNamespace(ctx, &persistence.Namespace{
		ID:            id,
		Name:          req.Name,
		Description:   req.Description,
		RetentionDays: retentionDays,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "create namespace: %v", err)
	}
	return &workflowservicev1.RegisterNamespaceResponse{}, nil
}

// DescribeNamespace describes a namespace.
func (h *Handler) DescribeNamespace(ctx context.Context, req *workflowservicev1.DescribeNamespaceRequest) (*workflowservicev1.DescribeNamespaceResponse, error) {
	ns, err := h.store.GetNamespace(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get namespace: %v", err)
	}
	if ns == nil {
		return nil, status.Errorf(codes.NotFound, "namespace %q not found", req.Name)
	}
	return &workflowservicev1.DescribeNamespaceResponse{
		Id:            ns.ID,
		Name:          ns.Name,
		Description:   ns.Description,
		RetentionDays: ns.RetentionDays,
	}, nil
}

// StartWorkflowExecution starts a workflow execution.
func (h *Handler) StartWorkflowExecution(ctx context.Context, req *workflowservicev1.StartWorkflowExecutionRequest) (*workflowservicev1.StartWorkflowExecutionResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolve namespace: %v", err)
	}

	runID := uuid.NewString()

	var inputBytes []byte
	if req.Input != nil {
		inputBytes, err = proto.Marshal(req.Input)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "marshal input: %v", err)
		}
	}

	wfTypeName := ""
	if req.WorkflowType != nil {
		wfTypeName = req.WorkflowType.Name
	}

	if _, err := h.history.StartWorkflowExecution(ctx, &history.StartRequest{
		NamespaceID:  nsID,
		WorkflowID:   req.WorkflowId,
		RunID:        runID,
		WorkflowType: wfTypeName,
		TaskQueue:    req.TaskQueue,
		Input:        inputBytes,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "start workflow: %v", err)
	}

	log.Info().Str("workflow_id", req.WorkflowId).Str("run_id", runID).Msg("workflow started")
	return &workflowservicev1.StartWorkflowExecutionResponse{RunId: runID}, nil
}

// DescribeWorkflowExecution describes a workflow execution.
func (h *Handler) DescribeWorkflowExecution(ctx context.Context, req *workflowservicev1.DescribeWorkflowExecutionRequest) (*workflowservicev1.DescribeWorkflowExecutionResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolve namespace: %v", err)
	}

	exec := req.Execution
	if exec == nil {
		return nil, status.Error(codes.InvalidArgument, "execution required")
	}

	wf, err := h.store.GetWorkflowExecution(ctx, nsID, exec.WorkflowId, exec.RunId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get workflow: %v", err)
	}
	if wf == nil {
		return nil, status.Errorf(codes.NotFound, "workflow not found")
	}

	resp := &workflowservicev1.DescribeWorkflowExecutionResponse{
		Execution:    &commonv1.WorkflowExecution{WorkflowId: wf.WorkflowID, RunId: wf.RunID},
		Status:       enumsv1.WorkflowExecutionStatus(wf.Status),
		WorkflowType: &commonv1.WorkflowType{Name: wf.WorkflowType},
		TaskQueue:    wf.TaskQueue,
		StartTime:    timestamppb.New(wf.StartedAt),
	}
	if wf.ClosedAt != nil {
		resp.CloseTime = timestamppb.New(*wf.ClosedAt)
	}
	return resp, nil
}

// GetWorkflowExecutionHistory returns the history of a workflow execution.
func (h *Handler) GetWorkflowExecutionHistory(ctx context.Context, req *workflowservicev1.GetWorkflowExecutionHistoryRequest) (*workflowservicev1.GetWorkflowExecutionHistoryResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolve namespace: %v", err)
	}

	exec := req.Execution
	if exec == nil {
		return nil, status.Error(codes.InvalidArgument, "execution required")
	}

	hist, err := h.loadHistory(ctx, nsID, exec.WorkflowId, exec.RunId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load history: %v", err)
	}

	return &workflowservicev1.GetWorkflowExecutionHistoryResponse{History: hist}, nil
}

func (h *Handler) loadHistory(ctx context.Context, nsID, workflowID, runID string) (*historyv1.History, error) {
	rows, err := h.store.GetHistoryEvents(ctx, nsID, workflowID, runID)
	if err != nil {
		return nil, err
	}
	hist := &historyv1.History{}
	for _, row := range rows {
		ev := &historyv1.HistoryEvent{}
		if err := proto.Unmarshal(row.EventData, ev); err != nil {
			return nil, err
		}
		hist.Events = append(hist.Events, ev)
	}
	return hist, nil
}

// PollWorkflowTaskQueue polls for a workflow task.
func (h *Handler) PollWorkflowTaskQueue(ctx context.Context, req *workflowservicev1.PollWorkflowTaskQueueRequest) (*workflowservicev1.PollWorkflowTaskQueueResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolve namespace: %v", err)
	}

	task, err := h.matching.PollWorkflowTask(ctx, nsID, req.TaskQueue, pollTimeout)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "poll workflow task: %v", err)
	}
	if task == nil {
		// Timeout - return empty response
		return &workflowservicev1.PollWorkflowTaskQueueResponse{}, nil
	}

	// Append WorkflowTaskStarted event
	startedEventID, err := h.history.AppendWorkflowTaskStarted(ctx, nsID, task.WorkflowID, task.RunID, task.ScheduledEventID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "append task started: %v", err)
	}

	// Load full history
	hist, err := h.loadHistory(ctx, nsID, task.WorkflowID, task.RunID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load history: %v", err)
	}

	wf, err := h.store.GetWorkflowExecution(ctx, nsID, task.WorkflowID, task.RunID)
	if err != nil || wf == nil {
		return nil, status.Errorf(codes.Internal, "get workflow execution: %v", err)
	}

	token, err := encodeTaskToken(history.TaskToken{
		NamespaceID:      nsID,
		WorkflowID:       task.WorkflowID,
		RunID:            task.RunID,
		ScheduledEventID: task.ScheduledEventID,
		Type:             "workflow",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encode task token: %v", err)
	}

	return &workflowservicev1.PollWorkflowTaskQueueResponse{
		TaskToken: token,
		WorkflowExecution: &commonv1.WorkflowExecution{
			WorkflowId: task.WorkflowID,
			RunId:      task.RunID,
		},
		WorkflowType:         &commonv1.WorkflowType{Name: wf.WorkflowType},
		StartedEventId:       startedEventID,
		PreviousStartedEventId: task.ScheduledEventID,
		History:              hist,
	}, nil
}

// RespondWorkflowTaskCompleted responds to a workflow task.
func (h *Handler) RespondWorkflowTaskCompleted(ctx context.Context, req *workflowservicev1.RespondWorkflowTaskCompletedRequest) (*workflowservicev1.RespondWorkflowTaskCompletedResponse, error) {
	token, err := decodeTaskToken(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "decode task token: %v", err)
	}

	var commands []*history.Command
	for _, cmd := range req.Commands {
		switch cmd.CommandType {
		case enumsv1.CommandType_COMMAND_TYPE_SCHEDULE_ACTIVITY_TASK:
			attrs := cmd.GetScheduleActivityTaskCommandAttributes()
			if attrs == nil {
				continue
			}
			var inputBytes []byte
			if attrs.Input != nil {
				inputBytes, err = proto.Marshal(attrs.Input)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "marshal activity input: %v", err)
				}
			}
			var timeout time.Duration
			if attrs.StartToCloseTimeout != nil {
				timeout = attrs.StartToCloseTimeout.AsDuration()
			}
			actType := ""
			if attrs.ActivityType != nil {
				actType = attrs.ActivityType.Name
			}
			commands = append(commands, &history.Command{
				Type:                "schedule_activity",
				ActivityID:          attrs.ActivityId,
				ActivityType:        actType,
				ActivityTaskQueue:   attrs.TaskQueue,
				ActivityInput:       inputBytes,
				StartToCloseTimeout: timeout,
			})

		case enumsv1.CommandType_COMMAND_TYPE_COMPLETE_WORKFLOW_EXECUTION:
			attrs := cmd.GetCompleteWorkflowExecutionCommandAttributes()
			var resultBytes []byte
			if attrs != nil && attrs.Result != nil {
				resultBytes, err = proto.Marshal(attrs.Result)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "marshal result: %v", err)
				}
			}
			commands = append(commands, &history.Command{
				Type:   "complete_workflow",
				Result: resultBytes,
			})

		case enumsv1.CommandType_COMMAND_TYPE_FAIL_WORKFLOW_EXECUTION:
			attrs := cmd.GetFailWorkflowExecutionCommandAttributes()
			var f *history.Failure
			if attrs != nil && attrs.Failure != nil {
				f = &history.Failure{
					Message:    attrs.Failure.Message,
					Type:       attrs.Failure.Type,
					StackTrace: attrs.Failure.StackTrace,
				}
			}
			commands = append(commands, &history.Command{
				Type:    "fail_workflow",
				Failure: f,
			})
		}
	}

	// Find the started event ID from the token's scheduled event ID
	startedEventID, err := h.findWorkflowTaskStartedEventID(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find started event: %v", err)
	}

	if err := h.history.RespondWorkflowTaskCompleted(ctx, token, commands, startedEventID); err != nil {
		return nil, status.Errorf(codes.Internal, "respond workflow task: %v", err)
	}

	return &workflowservicev1.RespondWorkflowTaskCompletedResponse{}, nil
}

func (h *Handler) findWorkflowTaskStartedEventID(ctx context.Context, token history.TaskToken) (int64, error) {
	rows, err := h.store.GetHistoryEvents(ctx, token.NamespaceID, token.WorkflowID, token.RunID)
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		if int32(enumsv1.EventType_EVENT_TYPE_WORKFLOW_TASK_STARTED) == row.EventType {
			ev := &historyv1.HistoryEvent{}
			if err := proto.Unmarshal(row.EventData, ev); err != nil {
				continue
			}
			attrs := ev.GetWorkflowTaskStartedEventAttributes()
			if attrs != nil && attrs.ScheduledEventId == token.ScheduledEventID {
				return row.EventID, nil
			}
		}
	}
	return 0, nil
}

// PollActivityTaskQueue polls for an activity task.
func (h *Handler) PollActivityTaskQueue(ctx context.Context, req *workflowservicev1.PollActivityTaskQueueRequest) (*workflowservicev1.PollActivityTaskQueueResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolve namespace: %v", err)
	}

	task, err := h.matching.PollActivityTask(ctx, nsID, req.TaskQueue, pollTimeout)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "poll activity task: %v", err)
	}
	if task == nil {
		return &workflowservicev1.PollActivityTaskQueueResponse{}, nil
	}

	startedEventID, err := h.history.AppendActivityTaskStarted(ctx, nsID, task.WorkflowID, task.RunID, task.ScheduledEventID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "append activity started: %v", err)
	}

	token, err := encodeTaskToken(history.TaskToken{
		NamespaceID:      nsID,
		WorkflowID:       task.WorkflowID,
		RunID:            task.RunID,
		ScheduledEventID: task.ScheduledEventID,
		Type:             "activity",
		ActivityID:       task.ActivityID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encode token: %v", err)
	}

	// Parse input payloads
	var inputPayloads *commonv1.Payloads
	if len(task.Input) > 0 {
		inputPayloads = &commonv1.Payloads{}
		if err := proto.Unmarshal(task.Input, inputPayloads); err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal activity input payloads")
		}
	}

	return &workflowservicev1.PollActivityTaskQueueResponse{
		TaskToken: token,
		WorkflowExecution: &commonv1.WorkflowExecution{
			WorkflowId: task.WorkflowID,
			RunId:      task.RunID,
		},
		ActivityType:     &commonv1.ActivityType{Name: task.ActivityType},
		ActivityId:       task.ActivityID,
		Input:            inputPayloads,
		ScheduledEventId: task.ScheduledEventID,
		StartedEventId:   startedEventID,
	}, nil
}

// RespondActivityTaskCompleted responds to an activity task completion.
func (h *Handler) RespondActivityTaskCompleted(ctx context.Context, req *workflowservicev1.RespondActivityTaskCompletedRequest) (*workflowservicev1.RespondActivityTaskCompletedResponse, error) {
	token, err := decodeTaskToken(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "decode token: %v", err)
	}

	var resultBytes []byte
	if req.Result != nil {
		resultBytes, err = proto.Marshal(req.Result)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "marshal result: %v", err)
		}
	}

	if err := h.history.RespondActivityTaskCompleted(ctx, token, resultBytes); err != nil {
		return nil, status.Errorf(codes.Internal, "respond activity completed: %v", err)
	}

	return &workflowservicev1.RespondActivityTaskCompletedResponse{}, nil
}

// RespondActivityTaskFailed responds to an activity task failure.
func (h *Handler) RespondActivityTaskFailed(ctx context.Context, req *workflowservicev1.RespondActivityTaskFailedRequest) (*workflowservicev1.RespondActivityTaskFailedResponse, error) {
	token, err := decodeTaskToken(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "decode token: %v", err)
	}

	var f *history.Failure
	if req.Failure != nil {
		f = &history.Failure{
			Message:    req.Failure.Message,
			Type:       req.Failure.Type,
			StackTrace: req.Failure.StackTrace,
		}
	}

	if err := h.history.RespondActivityTaskFailed(ctx, token, f); err != nil {
		return nil, status.Errorf(codes.Internal, "respond activity failed: %v", err)
	}

	return &workflowservicev1.RespondActivityTaskFailedResponse{}, nil
}

// RecordActivityTaskHeartbeat records a heartbeat for an activity task.
func (h *Handler) RecordActivityTaskHeartbeat(ctx context.Context, req *workflowservicev1.RecordActivityTaskHeartbeatRequest) (*workflowservicev1.RecordActivityTaskHeartbeatResponse, error) {
	_, err := decodeTaskToken(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "decode token: %v", err)
	}
	return &workflowservicev1.RecordActivityTaskHeartbeatResponse{CancelRequested: false}, nil
}

// Ensure unused import doesn't cause error
var _ = (*failurev1.Failure)(nil)

// SignalWorkflowExecution sends a signal to a running workflow.
func (h *Handler) SignalWorkflowExecution(ctx context.Context, req *workflowservicev1.SignalWorkflowExecutionRequest) (*workflowservicev1.SignalWorkflowExecutionResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	exec := req.WorkflowExecution
	if exec == nil {
		return nil, status.Error(codes.InvalidArgument, "workflow_execution is required")
	}
	inputBytes, _ := proto.Marshal(req.Input)
	if err := h.history.SignalWorkflowExecution(ctx, nsID, exec.WorkflowId, exec.RunId, req.SignalName, inputBytes); err != nil {
		return nil, status.Errorf(codes.Internal, "signal workflow: %v", err)
	}
	return &workflowservicev1.SignalWorkflowExecutionResponse{}, nil
}

// RequestCancelWorkflowExecution requests graceful cancellation.
func (h *Handler) RequestCancelWorkflowExecution(ctx context.Context, req *workflowservicev1.RequestCancelWorkflowExecutionRequest) (*workflowservicev1.RequestCancelWorkflowExecutionResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	exec := req.WorkflowExecution
	if exec == nil {
		return nil, status.Error(codes.InvalidArgument, "workflow_execution is required")
	}
	if err := h.history.RequestCancelWorkflowExecution(ctx, nsID, exec.WorkflowId, exec.RunId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "cancel workflow: %v", err)
	}
	return &workflowservicev1.RequestCancelWorkflowExecutionResponse{}, nil
}

// TerminateWorkflowExecution forcefully terminates a workflow.
func (h *Handler) TerminateWorkflowExecution(ctx context.Context, req *workflowservicev1.TerminateWorkflowExecutionRequest) (*workflowservicev1.TerminateWorkflowExecutionResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	exec := req.WorkflowExecution
	if exec == nil {
		return nil, status.Error(codes.InvalidArgument, "workflow_execution is required")
	}
	if err := h.history.TerminateWorkflowExecution(ctx, nsID, exec.WorkflowId, exec.RunId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "terminate workflow: %v", err)
	}
	return &workflowservicev1.TerminateWorkflowExecutionResponse{}, nil
}

// QueryWorkflow is not yet implemented; returns unimplemented.
func (h *Handler) QueryWorkflow(ctx context.Context, req *workflowservicev1.QueryWorkflowRequest) (*workflowservicev1.QueryWorkflowResponse, error) {
	return nil, status.Error(codes.Unimplemented, "QueryWorkflow not yet implemented")
}

// CountWorkflowExecutions returns the count of workflows in a namespace.
func (h *Handler) CountWorkflowExecutions(ctx context.Context, req *workflowservicev1.CountWorkflowExecutionsRequest) (*workflowservicev1.CountWorkflowExecutionsResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	count, err := h.history.CountWorkflowExecutions(ctx, nsID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count workflows: %v", err)
	}
	return &workflowservicev1.CountWorkflowExecutionsResponse{Count: count}, nil
}

// ListWorkflowExecutions lists workflow executions in a namespace.
func (h *Handler) ListWorkflowExecutions(ctx context.Context, req *workflowservicev1.ListWorkflowExecutionsRequest) (*workflowservicev1.ListWorkflowExecutionsResponse, error) {
	nsID, err := h.resolveNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	wfs, nextToken, err := h.history.ListWorkflowExecutions(ctx, nsID, req.PageSize, req.NextPageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list workflows: %v", err)
	}
	var infos []*commonv1.WorkflowExecutionInfo
	for _, wf := range wfs {
		info := &commonv1.WorkflowExecutionInfo{
			Execution: &commonv1.WorkflowExecution{
				WorkflowId: wf.WorkflowID,
				RunId:      wf.RunID,
			},
			Type:      &commonv1.WorkflowType{Name: wf.WorkflowType},
			TaskQueue: wf.TaskQueue,
			Status:    enumsv1.WorkflowExecutionStatus(wf.Status),
		}
		if !wf.StartedAt.IsZero() {
			ts := timestamppb.New(wf.StartedAt)
			info.StartTime = ts
		}
		if wf.ClosedAt != nil {
			info.CloseTime = timestamppb.New(*wf.ClosedAt)
		}
		infos = append(infos, info)
	}
	return &workflowservicev1.ListWorkflowExecutionsResponse{
		Executions:    infos,
		NextPageToken: nextToken,
	}, nil
}
