package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	failurev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/failure/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/pi/internal/event"
	"github.com/tempiex/pi/internal/metrics"
	"github.com/tempiex/pi/internal/session"
	"github.com/tempiex/pi/internal/tool"
)

type Executor struct {
	registry *tool.Registry
	bus      *event.Bus
	sessions *session.Manager
	metrics  *metrics.Metrics
	tracer   trace.Tracer
	log      zerolog.Logger
}

func NewExecutor(
	registry *tool.Registry,
	bus *event.Bus,
	sessions *session.Manager,
	m *metrics.Metrics,
	tracer trace.Tracer,
	log zerolog.Logger,
) *Executor {
	return &Executor{
		registry: registry,
		bus:      bus,
		sessions: sessions,
		metrics:  m,
		tracer:   tracer,
		log:      log,
	}
}

func (e *Executor) Execute(
	ctx context.Context,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	taskQueue string,
	client workflowservicev1.WorkflowServiceClient,
) error {
	toolName := ""
	if task.ActivityType != nil {
		toolName = task.ActivityType.Name
	}

	ctx, span := e.tracer.Start(ctx, fmt.Sprintf("activity/%s", toolName),
		trace.WithAttributes(attribute.String("task_queue", taskQueue)))
	defer span.End()

	start := time.Now()

	t, ok := e.registry.Get(toolName)
	if !ok {
		e.log.Warn().Str("tool", toolName).Msg("tool not found")
		_, err := client.RespondActivityTaskFailed(ctx, &workflowservicev1.RespondActivityTaskFailedRequest{
			TaskToken: task.TaskToken,
			Failure:   &failurev1.Failure{Message: fmt.Sprintf("tool not found: %s", toolName)},
		})
		return err
	}

	wfExec := task.WorkflowExecution
	workflowID := ""
	runID := ""
	if wfExec != nil {
		workflowID = wfExec.WorkflowId
		runID = wfExec.RunId
	}
	sessionID := workflowID + "/" + runID

	e.bus.Publish(event.Event{
		ID:         task.ActivityId,
		Type:       event.EventExecutionStart,
		SessionID:  sessionID,
		WorkflowID: workflowID,
		RunID:      runID,
		ToolName:   toolName,
		Timestamp:  time.Now(),
	})

	var inputJSON []byte
	if task.Input != nil && len(task.Input.Payloads) > 0 {
		inputJSON = task.Input.Payloads[0].Data
	}
	if inputJSON == nil {
		inputJSON = []byte("{}")
	}

	outputJSON, execErr := t.Run(ctx, inputJSON)

	duration := time.Since(start).Seconds() * 1000
	e.metrics.ActivityExecutions.Add(ctx, 1, metric.WithAttributes(attribute.String("tool", toolName)))
	e.metrics.ActivityDuration.Record(ctx, duration, metric.WithAttributes(attribute.String("tool", toolName)))

	if execErr != nil {
		e.bus.Publish(event.Event{
			ID:        task.ActivityId + "_err",
			Type:      event.EventExecutionError,
			SessionID: sessionID,
			ToolName:  toolName,
			Timestamp: time.Now(),
			Error:     execErr.Error(),
		})
		_, err := client.RespondActivityTaskFailed(ctx, &workflowservicev1.RespondActivityTaskFailedRequest{
			TaskToken: task.TaskToken,
			Failure:   &failurev1.Failure{Message: execErr.Error()},
		})
		return err
	}

	e.bus.Publish(event.Event{
		ID:        task.ActivityId + "_end",
		Type:      event.EventExecutionEnd,
		SessionID: sessionID,
		ToolName:  toolName,
		Timestamp: time.Now(),
	})

	_, err := client.RespondActivityTaskCompleted(ctx, &workflowservicev1.RespondActivityTaskCompletedRequest{
		TaskToken: task.TaskToken,
		Result: &commonv1.Payloads{
			Payloads: []*commonv1.Payload{{Data: outputJSON}},
		},
	})
	return err
}
