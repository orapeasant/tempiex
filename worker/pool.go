// Package worker provides a generic, embeddable Tempiex activity worker:
// it polls one or more task queues via WorkflowService.PollActivityTaskQueue
// and dispatches each dequeued task to a caller-supplied ActivityHandler.
//
// This package intentionally knows nothing about *what* an activity does —
// tool dispatch, session tracking, and event streaming are host-application
// concerns (see github.com/tempiex/pi for an AI-agent-harness host). Pool
// only owns the polling/dispatch/shutdown lifecycle that is common to any
// Tempiex activity worker.
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
)

// ActivityHandler executes one dequeued activity task and is responsible for
// reporting its outcome back to Tempiex itself (RespondActivityTaskCompleted
// or RespondActivityTaskFailed). Pool stays agnostic to how task payloads are
// interpreted; it just hands the task off and moves on to the next poll.
type ActivityHandler interface {
	Execute(
		ctx context.Context,
		task *workflowservicev1.PollActivityTaskQueueResponse,
		taskQueue string,
		client workflowservicev1.WorkflowServiceClient,
	) error
}

// Config describes one task queue this worker should poll. A Pool runs one
// polling goroutine per Config entry, so a single process can service
// multiple task queues concurrently (e.g. separating CPU-bound work from
// I/O-bound work onto different queues).
type Config struct {
	TaskQueue string
	// MaxConcurrentActivities and MaxConcurrentWorkflowTasks bound how many
	// tasks this queue's poller may have in flight at once. Reserved for a
	// future semaphore-based dispatcher; the current poller processes one
	// task at a time per queue (see runQueue).
	MaxConcurrentActivities    int
	MaxConcurrentWorkflowTasks int
}

// Pool owns one polling goroutine per configured task queue and forwards
// dequeued activity tasks to the shared ActivityHandler.
type Pool struct {
	cfg     []Config
	client  workflowservicev1.WorkflowServiceClient
	handler ActivityHandler
	log     zerolog.Logger
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// New constructs a Pool. Start must be called to begin polling.
func New(
	cfg []Config,
	client workflowservicev1.WorkflowServiceClient,
	handler ActivityHandler,
	log zerolog.Logger,
) *Pool {
	return &Pool{
		cfg:     cfg,
		client:  client,
		handler: handler,
		log:     log,
	}
}

// Start launches one polling goroutine per configured task queue. It returns
// immediately; polling continues until Stop is called or ctx is cancelled.
func (p *Pool) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	for _, qCfg := range p.cfg {
		p.wg.Add(1)
		go p.runQueue(ctx, qCfg)
	}
}

// Stop signals all polling goroutines to exit and blocks until they have
// drained their current in-flight task (if any) and returned.
func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}

// runQueue is the long-poll loop for a single task queue: block on
// PollActivityTaskQueue (server-side long poll), dispatch whatever comes
// back to the handler, then poll again. On transient poll errors it backs
// off for a second before retrying rather than busy-looping.
func (p *Pool) runQueue(ctx context.Context, cfg Config) {
	defer p.wg.Done()
	log := p.log.With().Str("task_queue", cfg.TaskQueue).Logger()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := p.client.PollActivityTaskQueue(ctx, &workflowservicev1.PollActivityTaskQueueRequest{
			TaskQueue: cfg.TaskQueue,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error().Err(err).Msg("poll error; backing off")
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}

		// An empty task token means the long poll timed out with no work
		// available (not an error) — loop back and poll again immediately.
		if len(resp.TaskToken) == 0 {
			continue
		}

		if err := p.handler.Execute(ctx, resp, cfg.TaskQueue, p.client); err != nil {
			log.Error().Err(err).Msg("activity handler error")
		}
	}
}
