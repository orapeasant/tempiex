package worker

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"

	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	"google.golang.org/grpc"
)

// fakeClient embeds the (large, generated) WorkflowServiceClient interface
// with a nil value and overrides only the methods this test exercises —
// PollActivityTaskQueue is the only one Pool ever calls directly.
type fakeClient struct {
	workflowservicev1.WorkflowServiceClient
	polled  chan struct{}
	dequeue chan *workflowservicev1.PollActivityTaskQueueResponse
}

func (f *fakeClient) PollActivityTaskQueue(
	ctx context.Context,
	_ *workflowservicev1.PollActivityTaskQueueRequest,
	_ ...grpc.CallOption,
) (*workflowservicev1.PollActivityTaskQueueResponse, error) {
	select {
	case f.polled <- struct{}{}:
	default:
	}
	select {
	case resp := <-f.dequeue:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// fakeHandler records every task it was asked to execute.
type fakeHandler struct {
	executed chan *workflowservicev1.PollActivityTaskQueueResponse
}

func (h *fakeHandler) Execute(
	_ context.Context,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	_ string,
	_ workflowservicev1.WorkflowServiceClient,
) error {
	h.executed <- task
	return nil
}

func TestPoolDispatchesPolledTaskToHandler(t *testing.T) {
	client := &fakeClient{
		polled:  make(chan struct{}, 1),
		dequeue: make(chan *workflowservicev1.PollActivityTaskQueueResponse, 1),
	}
	handler := &fakeHandler{executed: make(chan *workflowservicev1.PollActivityTaskQueueResponse, 1)}

	pool := New(
		[]Config{{TaskQueue: "test-queue"}},
		client,
		handler,
		zerolog.Nop(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	defer pool.Stop()

	select {
	case <-client.polled:
	case <-time.After(time.Second):
		t.Fatal("expected PollActivityTaskQueue to be called")
	}

	want := &workflowservicev1.PollActivityTaskQueueResponse{TaskToken: []byte("token-1")}
	client.dequeue <- want

	select {
	case got := <-handler.executed:
		if string(got.TaskToken) != string(want.TaskToken) {
			t.Fatalf("handler got task token %q, want %q", got.TaskToken, want.TaskToken)
		}
	case <-time.After(time.Second):
		t.Fatal("expected handler.Execute to be called with the polled task")
	}
}

func TestPoolSkipsEmptyPollResult(t *testing.T) {
	// An empty TaskToken represents a long-poll timeout with no work
	// available; Pool must not invoke the handler for it.
	client := &fakeClient{
		polled:  make(chan struct{}, 4),
		dequeue: make(chan *workflowservicev1.PollActivityTaskQueueResponse, 4),
	}
	handler := &fakeHandler{executed: make(chan *workflowservicev1.PollActivityTaskQueueResponse, 1)}

	pool := New([]Config{{TaskQueue: "test-queue"}}, client, handler, zerolog.Nop())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	client.dequeue <- &workflowservicev1.PollActivityTaskQueueResponse{}

	select {
	case <-handler.executed:
		t.Fatal("handler should not be invoked for an empty poll result")
	case <-time.After(200 * time.Millisecond):
	}

	pool.Stop()
}

func TestPoolStopWaitsForGoroutinesToExit(t *testing.T) {
	// dequeue is left permanently empty: both queue goroutines stay blocked
	// inside PollActivityTaskQueue until Stop() cancels the pool's context.
	client := &fakeClient{
		polled:  make(chan struct{}, 2),
		dequeue: make(chan *workflowservicev1.PollActivityTaskQueueResponse),
	}
	handler := &fakeHandler{executed: make(chan *workflowservicev1.PollActivityTaskQueueResponse, 1)}

	pool := New(
		[]Config{{TaskQueue: "q1"}, {TaskQueue: "q2"}},
		client,
		handler,
		zerolog.Nop(),
	)

	pool.Start(context.Background())

	stopped := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("expected Stop to return once both queue goroutines observed ctx cancellation")
	}
}
