package activity

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
)

// ErrCancelled is returned by Context.Heartbeat once Tempiex reports that the
// activity's workflow has requested cancellation. The activity's context is
// cancelled at the same time, so an activity that watches ctx.Done() observes
// cancellation without checking the heartbeat's return value.
var ErrCancelled = errors.New("activity cancelled")

// ErrNotRegistered is reported when a dequeued task names an activity type
// that is not in the Registry.
var ErrNotRegistered = errors.New("activity type not registered")

// InputError wraps a failure to decode a task's input payload. Decoding will
// fail identically on every retry, so callers surface it as a distinct
// failure type rather than a transient error.
type InputError struct{ Err error }

func (e *InputError) Error() string { return "decode activity input: " + e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }

// Info carries the Tempiex task details an activity may legitimately read.
//
// Namespace comes from worker configuration rather than the task: the current
// PollActivityTaskQueueResponse does not carry it. There is likewise no
// attempt counter on the task, so retries are not distinguishable from first
// attempts by the worker today.
type Info struct {
	Namespace           string
	TaskQueue           string
	WorkflowID          string
	WorkflowRunID       string
	ActivityID          string
	ActivityType        string
	ScheduledEventID    int64
	StartedEventID      int64
	StartToCloseTimeout time.Duration
	HeartbeatDetails    []byte
	// RunID is the worker-local run id for this execution.
	RunID string
}

// Context is the context passed to an activity implementation. It embeds the
// standard context (cancelled on worker shutdown, task timeout, or workflow
// cancellation) and adds task info, heartbeating, and progress publication.
type Context interface {
	context.Context
	Info() Info
	// Heartbeat reports liveness to Tempiex and returns ErrCancelled when the
	// server reports that cancellation was requested.
	Heartbeat(details any) error
	// Publish emits an execution_update event on this run's event stream.
	Publish(payload any)
}

type activityContext struct {
	context.Context
	cancel    context.CancelFunc
	info      Info
	taskToken []byte
	client    workflowservicev1.WorkflowServiceClient
	publish   func(payload any)
}

func (c *activityContext) Info() Info { return c.info }

func (c *activityContext) Publish(payload any) {
	if c.publish != nil {
		c.publish(payload)
	}
}

func (c *activityContext) Heartbeat(details any) error {
	req := &workflowservicev1.RecordActivityTaskHeartbeatRequest{TaskToken: c.taskToken}
	if details != nil {
		data, err := json.Marshal(details)
		if err != nil {
			return err
		}
		req.Details = &commonv1.Payloads{Payloads: []*commonv1.Payload{{Data: data}}}
	}

	resp, err := c.client.RecordActivityTaskHeartbeat(c.Context, req)
	if err != nil {
		return err
	}
	if resp.GetCancelRequested() {
		c.cancel()
		return ErrCancelled
	}
	return nil
}
