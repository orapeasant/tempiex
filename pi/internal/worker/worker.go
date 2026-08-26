package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/pi/internal/config"
	"github.com/tempiex/pi/internal/worker/activity"
)

type WorkerPool struct {
	cfg    []config.WorkerConfig
	client workflowservicev1.WorkflowServiceClient
	exec   *activity.Executor
	log    zerolog.Logger
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func New(
	cfg []config.WorkerConfig,
	client workflowservicev1.WorkflowServiceClient,
	exec *activity.Executor,
	log zerolog.Logger,
) *WorkerPool {
	return &WorkerPool{
		cfg:    cfg,
		client: client,
		exec:   exec,
		log:    log,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	wp.cancel = cancel

	for _, wCfg := range wp.cfg {
		wp.wg.Add(1)
		go wp.runWorker(ctx, wCfg)
	}
}

func (wp *WorkerPool) Stop() {
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.wg.Wait()
}

func (wp *WorkerPool) runWorker(ctx context.Context, cfg config.WorkerConfig) {
	defer wp.wg.Done()
	log := wp.log.With().Str("task_queue", cfg.TaskQueue).Logger()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := wp.client.PollActivityTaskQueue(ctx, &workflowservicev1.PollActivityTaskQueueRequest{
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

		if len(resp.TaskToken) == 0 {
			continue
		}

		if err := wp.exec.Execute(ctx, resp, cfg.TaskQueue, wp.client); err != nil {
			log.Error().Err(err).Msg("activity execute error")
		}
	}
}
