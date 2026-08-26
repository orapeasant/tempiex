package activity

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/pi/internal/event"
	"github.com/tempiex/pi/internal/metrics"
	"github.com/tempiex/pi/internal/session"
	"github.com/tempiex/pi/internal/tool"
)

type mockClient struct {
	workflowservicev1.WorkflowServiceClient
	failedCalled    bool
	completedCalled bool
}

func (m *mockClient) RespondActivityTaskFailed(_ context.Context, _ *workflowservicev1.RespondActivityTaskFailedRequest, _ ...grpc.CallOption) (*workflowservicev1.RespondActivityTaskFailedResponse, error) {
	m.failedCalled = true
	return &workflowservicev1.RespondActivityTaskFailedResponse{}, nil
}

func (m *mockClient) RespondActivityTaskCompleted(_ context.Context, _ *workflowservicev1.RespondActivityTaskCompletedRequest, _ ...grpc.CallOption) (*workflowservicev1.RespondActivityTaskCompletedResponse, error) {
	m.completedCalled = true
	return &workflowservicev1.RespondActivityTaskCompletedResponse{}, nil
}

func newTestExecutor(t *testing.T) *Executor {
	t.Helper()
	store, err := session.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	mgr := session.NewManager(store)

	nm := noop.NewMeterProvider().Meter("test")
	m := &metrics.Metrics{}
	m.ActivityExecutions, _ = nm.Int64Counter("a")
	m.ActivityDuration, _ = nm.Float64Histogram("b")
	m.SessionsActive, _ = nm.Int64UpDownCounter("c")
	m.WorkerPollLatency, _ = nm.Float64Histogram("d")

	return NewExecutor(
		tool.NewRegistry(),
		event.New(),
		mgr,
		m,
		tracenoop.NewTracerProvider().Tracer("test"),
		zerolog.Nop(),
	)
}

func TestExecuteToolNotFound(t *testing.T) {
	exec := newTestExecutor(t)
	client := &mockClient{}

	task := &workflowservicev1.PollActivityTaskQueueResponse{
		TaskToken:    []byte("token"),
		ActivityId:   "act1",
		ActivityType: &commonv1.ActivityType{Name: "nonexistent"},
	}

	if err := exec.Execute(context.Background(), task, "default", client); err != nil {
		t.Fatal(err)
	}
	if !client.failedCalled {
		t.Fatal("expected RespondActivityTaskFailed to be called")
	}
}
