package activity

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/worker/metrics"
	"github.com/tempiex/worker/run"
)

type mockClient struct {
	workflowservicev1.WorkflowServiceClient
	failed          *workflowservicev1.RespondActivityTaskFailedRequest
	completed       *workflowservicev1.RespondActivityTaskCompletedRequest
	heartbeats      int
	cancelRequested bool
}

func (m *mockClient) RespondActivityTaskFailed(_ context.Context, req *workflowservicev1.RespondActivityTaskFailedRequest, _ ...grpc.CallOption) (*workflowservicev1.RespondActivityTaskFailedResponse, error) {
	m.failed = req
	return &workflowservicev1.RespondActivityTaskFailedResponse{}, nil
}

func (m *mockClient) RespondActivityTaskCompleted(_ context.Context, req *workflowservicev1.RespondActivityTaskCompletedRequest, _ ...grpc.CallOption) (*workflowservicev1.RespondActivityTaskCompletedResponse, error) {
	m.completed = req
	return &workflowservicev1.RespondActivityTaskCompletedResponse{}, nil
}

func (m *mockClient) RecordActivityTaskHeartbeat(_ context.Context, _ *workflowservicev1.RecordActivityTaskHeartbeatRequest, _ ...grpc.CallOption) (*workflowservicev1.RecordActivityTaskHeartbeatResponse, error) {
	m.heartbeats++
	return &workflowservicev1.RecordActivityTaskHeartbeatResponse{CancelRequested: m.cancelRequested}, nil
}

func newTestExecutor(t *testing.T, registry *Registry) (*Executor, *run.Manager) {
	t.Helper()

	store, err := run.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	runs := run.NewManager(store, run.NewBus(), 100)
	exec := NewExecutor(registry, runs, metrics.Noop(),
		tracenoop.NewTracerProvider().Tracer("test"), "default", zerolog.Nop())
	return exec, runs
}

func testTask(activityType string, input any) *workflowservicev1.PollActivityTaskQueueResponse {
	task := &workflowservicev1.PollActivityTaskQueueResponse{
		TaskToken:         []byte("token"),
		ActivityId:        "act1",
		ActivityType:      &commonv1.ActivityType{Name: activityType},
		WorkflowExecution: &commonv1.WorkflowExecution{WorkflowId: "wf1", RunId: "wfr1"},
	}
	if input != nil {
		data, _ := json.Marshal(input)
		task.Input = &commonv1.Payloads{Payloads: []*commonv1.Payload{{Data: data}}}
	}
	return task
}

func TestExecuteActivityNotRegistered(t *testing.T) {
	exec, _ := newTestExecutor(t, NewRegistry())
	client := &mockClient{}

	if err := exec.Execute(context.Background(), testTask("nonexistent", nil), "default", client); err != nil {
		t.Fatal(err)
	}
	if client.failed == nil {
		t.Fatal("expected RespondActivityTaskFailed to be called")
	}
	if got := client.failed.Failure.Type; got != FailureTypeNotRegistered {
		t.Fatalf("expected %s failure type, got %s", FailureTypeNotRegistered, got)
	}
}

func TestExecuteCompletesAndRecordsRun(t *testing.T) {
	registry := NewRegistry()
	Register(registry, echoActivity())
	exec, runs := newTestExecutor(t, registry)
	client := &mockClient{}

	task := testTask("test-activity", testInput{Value: "hello"})
	if err := exec.Execute(context.Background(), task, "queue-a", client); err != nil {
		t.Fatal(err)
	}

	if client.completed == nil {
		t.Fatal("expected RespondActivityTaskCompleted to be called")
	}
	var out testOutput
	_ = json.Unmarshal(client.completed.Result.Payloads[0].Data, &out)
	if out.Echo != "hello" {
		t.Fatalf("unexpected result payload: %s", out.Echo)
	}

	list, err := runs.List(context.Background(), run.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 tracked run, got %d", len(list))
	}
	r := list[0]
	if r.Status != run.StatusCompleted {
		t.Fatalf("expected completed, got %s", r.Status)
	}
	if r.WorkflowID != "wf1" || r.WorkflowRunID != "wfr1" || r.TaskQueue != "queue-a" || r.Namespace != "default" {
		t.Fatalf("run not correlated to the task: %+v", r)
	}

	events, _ := runs.LoadEvents(context.Background(), r.ID)
	if len(events) < 2 {
		t.Fatalf("expected start and end events, got %d", len(events))
	}
	if events[0].Type != run.EventExecutionStart {
		t.Fatalf("expected first event to be start, got %s", events[0].Type)
	}
	if events[len(events)-1].Type != run.EventExecutionEnd {
		t.Fatalf("expected last event to be end, got %s", events[len(events)-1].Type)
	}
}

func TestExecuteFailureMarksRunFailed(t *testing.T) {
	registry := NewRegistry()
	Register(registry, Activity[testInput, testOutput]{
		Name: "boom",
		Execute: func(_ Context, _ testInput) (testOutput, error) {
			return testOutput{}, errors.New("exploded")
		},
	})
	exec, runs := newTestExecutor(t, registry)
	client := &mockClient{}

	if err := exec.Execute(context.Background(), testTask("boom", testInput{}), "default", client); err != nil {
		t.Fatal(err)
	}

	if client.failed == nil || client.failed.Failure.Message != "exploded" {
		t.Fatalf("expected the activity error to be reported, got %+v", client.failed)
	}
	if client.failed.Failure.Type != FailureTypeExecution {
		t.Fatalf("unexpected failure type: %s", client.failed.Failure.Type)
	}

	list, _ := runs.List(context.Background(), run.Filter{})
	if len(list) != 1 || list[0].Status != run.StatusFailed {
		t.Fatalf("expected a failed run, got %+v", list)
	}
}

func TestExecuteBadInputIsReportedDistinctly(t *testing.T) {
	registry := NewRegistry()
	Register(registry, echoActivity())
	exec, _ := newTestExecutor(t, registry)
	client := &mockClient{}

	task := testTask("test-activity", nil)
	task.Input = &commonv1.Payloads{Payloads: []*commonv1.Payload{{Data: []byte("not json")}}}

	if err := exec.Execute(context.Background(), task, "default", client); err != nil {
		t.Fatal(err)
	}
	if client.failed == nil || client.failed.Failure.Type != FailureTypeBadInput {
		t.Fatalf("expected %s, got %+v", FailureTypeBadInput, client.failed)
	}
}

func TestHeartbeatReportsCancellation(t *testing.T) {
	registry := NewRegistry()
	Register(registry, Activity[testInput, testOutput]{
		Name: "heartbeater",
		Execute: func(ctx Context, _ testInput) (testOutput, error) {
			if err := ctx.Heartbeat(map[string]string{"progress": "half"}); err != nil {
				return testOutput{}, err
			}
			return testOutput{}, nil
		},
	})
	exec, runs := newTestExecutor(t, registry)
	client := &mockClient{cancelRequested: true}

	if err := exec.Execute(context.Background(), testTask("heartbeater", testInput{}), "default", client); err != nil {
		t.Fatal(err)
	}

	if client.heartbeats != 1 {
		t.Fatalf("expected 1 heartbeat, got %d", client.heartbeats)
	}
	if client.failed == nil || client.failed.Failure.Type != FailureTypeCancelled {
		t.Fatalf("expected a cancellation failure, got %+v", client.failed)
	}

	list, _ := runs.List(context.Background(), run.Filter{})
	if len(list) != 1 || list[0].Status != run.StatusCancelled {
		t.Fatalf("expected a cancelled run, got %+v", list)
	}
}

func TestPublishEmitsUpdateEvents(t *testing.T) {
	registry := NewRegistry()
	Register(registry, Activity[testInput, testOutput]{
		Name: "chatty",
		Execute: func(ctx Context, _ testInput) (testOutput, error) {
			ctx.Publish(map[string]string{"stdout": "line 1"})
			return testOutput{}, nil
		},
	})
	exec, runs := newTestExecutor(t, registry)

	if err := exec.Execute(context.Background(), testTask("chatty", testInput{}), "default", &mockClient{}); err != nil {
		t.Fatal(err)
	}

	list, _ := runs.List(context.Background(), run.Filter{})
	events, _ := runs.LoadEvents(context.Background(), list[0].ID)

	var updates int
	for _, e := range events {
		if e.Type == run.EventExecutionUpdate {
			updates++
		}
	}
	if updates != 1 {
		t.Fatalf("expected 1 update event, got %d in %v", updates, events)
	}
}

func TestInfoCarriesTaskDetails(t *testing.T) {
	var got Info
	registry := NewRegistry()
	Register(registry, Activity[testInput, testOutput]{
		Name: "inspector",
		Execute: func(ctx Context, _ testInput) (testOutput, error) {
			got = ctx.Info()
			return testOutput{}, nil
		},
	})
	exec, _ := newTestExecutor(t, registry)

	task := testTask("inspector", testInput{})
	task.ScheduledEventId = 7
	task.StartedEventId = 8

	if err := exec.Execute(context.Background(), task, "queue-b", &mockClient{}); err != nil {
		t.Fatal(err)
	}

	if got.Namespace != "default" || got.TaskQueue != "queue-b" {
		t.Fatalf("unexpected namespace/queue: %+v", got)
	}
	if got.WorkflowID != "wf1" || got.WorkflowRunID != "wfr1" || got.ActivityID != "act1" {
		t.Fatalf("unexpected correlation ids: %+v", got)
	}
	if got.ScheduledEventID != 7 || got.StartedEventID != 8 {
		t.Fatalf("unexpected event ids: %+v", got)
	}
}
