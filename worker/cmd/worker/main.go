// Command worker is a standalone, generic Tempiex activity worker binary.
// It polls the configured task queues and dispatches to a built-in "echo"
// ActivityHandler that returns whatever input it received — a reference
// implementation for anyone embedding github.com/tempiex/worker directly,
// and a smoke-test target for grpcurl (schedule an "echo" activity and
// confirm the response matches the input).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	tempiexworker "github.com/tempiex/worker"
	"github.com/tempiex/worker/internal/config"
)

func main() {
	configPath := flag.String("config", "config/development.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}

	logger := buildLogger(cfg.Log)
	log.Logger = logger

	var dialOpts []grpc.DialOption
	if !cfg.Tempiex.TLS.Enabled {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(cfg.Tempiex.Address, dialOpts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("grpc connect")
	}
	defer conn.Close()

	client := workflowservicev1.NewWorkflowServiceClient(conn)

	// worker.Config is the module's own Go API type, kept separate from
	// this binary's YAML config schema (see internal/config) — mirrors the
	// same translation pi/internal/agent performs.
	workerCfgs := make([]tempiexworker.Config, len(cfg.Workers))
	for i, w := range cfg.Workers {
		workerCfgs[i] = tempiexworker.Config{
			TaskQueue:                  w.TaskQueue,
			MaxConcurrentActivities:    w.MaxConcurrentActivities,
			MaxConcurrentWorkflowTasks: w.MaxConcurrentWorkflowTasks,
		}
	}

	pool := tempiexworker.New(workerCfgs, client, &echoHandler{log: logger}, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool.Start(ctx)
	logger.Info().Str("address", cfg.Tempiex.Address).Msg("worker polling started")

	<-ctx.Done()
	logger.Info().Msg("shutting down")
	pool.Stop()
}

func buildLogger(cfg config.LogConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	if cfg.Format == "console" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(level).With().Timestamp().Logger()
	}
	return zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
}

// echoHandler implements worker.ActivityHandler by completing every task
// with its own input as the result, unchanged.
type echoHandler struct {
	log zerolog.Logger
}

func (h *echoHandler) Execute(
	ctx context.Context,
	task *workflowservicev1.PollActivityTaskQueueResponse,
	taskQueue string,
	client workflowservicev1.WorkflowServiceClient,
) error {
	h.log.Info().
		Str("task_queue", taskQueue).
		Str("activity_id", task.ActivityId).
		Msg("echoing activity input back as result")

	result := task.Input
	if result == nil {
		result = &commonv1.Payloads{}
	}

	_, err := client.RespondActivityTaskCompleted(ctx, &workflowservicev1.RespondActivityTaskCompletedRequest{
		TaskToken: task.TaskToken,
		Result:    result,
	})
	return err
}
