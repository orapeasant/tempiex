package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	failurev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/failure/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/worker/metrics"
	"github.com/tempiex/worker/run"
)

// Failure type strings reported to Tempiex in Failure.Type. They are stable
// identifiers a workflow (or an operator reading history) can match on.
// payloadEncoding matches what sdk-python's converter writes, so activity
// inputs and results are interchangeable across SDKs.
const payloadEncoding = "json/plain"

const (
	FailureTypeNotRegistered = "ActivityTypeNotRegistered"
	FailureTypeBadInput      = "ActivityInputDecodeError"
	FailureTypeCancelled     = "ActivityCancelled"
	FailureTypeExecution     = "ActivityExecutionError"
)

// Executor implements worker.ActivityHandler on top of a Registry: it resolves
// the dequeued task's activity type, tracks the execution as a run, runs the
// activity, and reports the outcome back to Tempiex.
type Executor struct {
	registry  *activityRegistry
	runs      *run.Manager
	metrics   *metrics.Metrics
	tracer    trace.Tracer
	log       zerolog.Logger
	namespace string
}

// activityRegistry is an alias so the struct field reads clearly while keeping
// Registry as the exported name callers construct.
type activityRegistry = Registry

// NewExecutor builds an Executor. namespace is the Tempiex namespace this
// worker polls; it is recorded on runs because the activity task itself does
// not carry it.
func NewExecutor(
	registry *Registry,
	runs *run.Manager,
	m *metrics.Metrics,
	tracer trace.Tracer,
	namespace string,
	log zerolog.Logger,
) *Executor {
	return &Executor{
		registry:  registry,
		runs:      runs,
		metrics:   m,
		tracer:    tracer,
		namespace: namespace,
		log:       log,
	}
}

// Execute satisfies worker.ActivityHandler.
func (e *Executor) Execute(
	ctx context.Context,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	taskQueue string,
	client workflowservicev1.WorkflowServiceClient,
) error {
	activityType := task.GetActivityType().GetName()
	workflowID := task.GetWorkflowExecution().GetWorkflowId()
	workflowRunID := task.GetWorkflowExecution().GetRunId()
	runID := runIDFor(workflowID, workflowRunID, task.GetActivityId())

	ctx, span := e.tracer.Start(ctx, fmt.Sprintf("activity/%s", activityType),
		trace.WithAttributes(
			attribute.String("task_queue", taskQueue),
			attribute.String("activity_type", activityType),
			attribute.String("workflow_id", workflowID),
			attribute.String("run_id", workflowRunID),
		))
	defer span.End()

	log := e.log.With().
		Str("task_queue", taskQueue).
		Str("activity_type", activityType).
		Str("workflow_id", workflowID).
		Str("run_id", workflowRunID).
		Str("worker_run_id", runID).
		Logger()

	act, ok := e.registry.Get(activityType)
	if !ok {
		log.Warn().Msg("activity type not registered")
		return e.fail(ctx, client, task, FailureTypeNotRegistered,
			fmt.Sprintf("%v: %s", ErrNotRegistered, activityType))
	}

	r := run.Run{
		ID:            runID,
		Namespace:     e.namespace,
		TaskQueue:     taskQueue,
		WorkflowID:    workflowID,
		WorkflowRunID: workflowRunID,
		ActivityID:    task.GetActivityId(),
		ActivityType:  activityType,
		StartedAt:     time.Now().UTC(),
	}
	if err := e.runs.Open(ctx, r); err != nil {
		log.Error().Err(err).Msg("open run")
	}
	e.metrics.RunsActive.Add(ctx, 1, metric.WithAttributes(attribute.String("task_queue", taskQueue)))
	defer e.metrics.RunsActive.Add(ctx, -1, metric.WithAttributes(attribute.String("task_queue", taskQueue)))

	inputJSON := inputPayload(task)
	e.publish(ctx, r, run.EventExecutionStart, rawJSON(inputJSON), "")

	actx, cancel := e.newContext(ctx, task, r, client)
	defer cancel()

	start := time.Now()
	outputJSON, execErr := act.Run(actx, inputJSON)
	e.recordDuration(ctx, activityType, taskQueue, start, execErr)

	switch {
	case execErr == nil:
		e.publish(ctx, r, run.EventExecutionEnd, rawJSON(outputJSON), "")
		e.closeRun(ctx, runID, run.StatusCompleted, log)
		_, err := client.RespondActivityTaskCompleted(ctx, &workflowservicev1.RespondActivityTaskCompletedRequest{
			TaskToken: task.TaskToken,
			Result: &commonv1.Payloads{Payloads: []*commonv1.Payload{
				{Encoding: payloadEncoding, Data: outputJSON},
			}},
		})
		return err

	case errors.Is(execErr, ErrCancelled), errors.Is(execErr, context.Canceled):
		// Tempiex has no RespondActivityTaskCanceled RPC, so a cancelled
		// activity is reported as a failure carrying the cancellation type.
		e.publish(ctx, r, run.EventExecutionCancelled, nil, execErr.Error())
		e.closeRun(ctx, runID, run.StatusCancelled, log)
		return e.fail(ctx, client, task, FailureTypeCancelled, execErr.Error())

	default:
		failureType := FailureTypeExecution
		var inputErr *InputError
		if errors.As(execErr, &inputErr) {
			failureType = FailureTypeBadInput
		}
		log.Warn().Err(execErr).Str("failure_type", failureType).Msg("activity failed")
		e.publish(ctx, r, run.EventExecutionError, nil, execErr.Error())
		e.closeRun(ctx, runID, run.StatusFailed, log)
		return e.fail(ctx, client, task, failureType, execErr.Error())
	}
}

func (e *Executor) newContext(
	ctx context.Context,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	r run.Run,
	client workflowservicev1.WorkflowServiceClient,
) (*activityContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	info := Info{
		Namespace:        r.Namespace,
		TaskQueue:        r.TaskQueue,
		WorkflowID:       r.WorkflowID,
		WorkflowRunID:    r.WorkflowRunID,
		ActivityID:       r.ActivityID,
		ActivityType:     r.ActivityType,
		ScheduledEventID: task.GetScheduledEventId(),
		StartedEventID:   task.GetStartedEventId(),
		RunID:            r.ID,
	}
	if d := task.GetStartToCloseTimeout(); d != nil {
		info.StartToCloseTimeout = d.AsDuration()
	}
	if hb := task.GetHeartbeatDetails(); hb != nil && len(hb.Payloads) > 0 {
		info.HeartbeatDetails = hb.Payloads[0].Data
	}

	actx := &activityContext{
		Context:   ctx,
		cancel:    cancel,
		info:      info,
		taskToken: task.TaskToken,
		client:    client,
		publish: func(payload any) {
			e.publish(ctx, r, run.EventExecutionUpdate, payload, "")
		},
	}
	return actx, cancel
}

func (e *Executor) publish(ctx context.Context, r run.Run, typ run.EventType, payload any, errMsg string) {
	ev := run.Event{
		ID:            fmt.Sprintf("%s/%s/%d", r.ID, typ, time.Now().UnixNano()),
		Type:          typ,
		RunID:         r.ID,
		WorkflowID:    r.WorkflowID,
		WorkflowRunID: r.WorkflowRunID,
		ActivityID:    r.ActivityID,
		ActivityType:  r.ActivityType,
		Timestamp:     time.Now().UTC(),
		Payload:       payload,
		Error:         errMsg,
	}
	if err := e.runs.Publish(ctx, ev); err != nil {
		e.log.Warn().Err(err).Str("worker_run_id", r.ID).Msg("persist event")
	}
}

func (e *Executor) closeRun(ctx context.Context, runID string, status run.Status, log zerolog.Logger) {
	if err := e.runs.Close(ctx, runID, status); err != nil {
		log.Warn().Err(err).Msg("close run")
	}
}

func (e *Executor) recordDuration(ctx context.Context, activityType, taskQueue string, start time.Time, execErr error) {
	status := "success"
	switch {
	case errors.Is(execErr, ErrCancelled), errors.Is(execErr, context.Canceled):
		status = "cancelled"
	case execErr != nil:
		status = "error"
	}
	attrs := metric.WithAttributes(
		attribute.String("activity_type", activityType),
		attribute.String("task_queue", taskQueue),
		attribute.String("status", status),
	)
	e.metrics.ActivityExecutions.Add(ctx, 1, attrs)
	e.metrics.ActivityDuration.Record(ctx, float64(time.Since(start).Milliseconds()), attrs)
}

func (e *Executor) fail(
	ctx context.Context,
	client workflowservicev1.WorkflowServiceClient,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	failureType, message string,
) error {
	_, err := client.RespondActivityTaskFailed(ctx, &workflowservicev1.RespondActivityTaskFailedRequest{
		TaskToken: task.TaskToken,
		Failure:   &failurev1.Failure{Message: message, Type: failureType},
	})
	return err
}

// runIDFor derives a stable worker-local run id. Keying on the activity id
// within a workflow run means a re-delivered task reuses its run record
// instead of creating a duplicate.
func runIDFor(workflowID, workflowRunID, activityID string) string {
	return fmt.Sprintf("%s/%s/%s", workflowID, workflowRunID, activityID)
}

func inputPayload(task *workflowservicev1.PollActivityTaskQueueResponse) []byte {
	if in := task.GetInput(); in != nil && len(in.Payloads) > 0 && in.Payloads[0].Data != nil {
		return in.Payloads[0].Data
	}
	return []byte("{}")
}

// rawJSON lets already-encoded payloads ride through the event's `any` field
// without a decode/re-encode round trip.
func rawJSON(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return jsonRaw(data)
}

type jsonRaw []byte

func (r jsonRaw) MarshalJSON() ([]byte, error) { return r, nil }
