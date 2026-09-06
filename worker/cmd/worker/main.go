// Command worker is the Tempiex activity worker binary.
//
// It polls the configured task queues, dispatches each dequeued task to the
// typed activity registry, tracks executions as runs, and serves the HTTP
// inspection API on :8090. The built-in activities (echo, shell, http,
// file_read, file_write) make a fresh checkout smoke-testable without writing
// any handler code.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	tempiexworker "github.com/tempiex/worker"
	"github.com/tempiex/worker/activities/file"
	"github.com/tempiex/worker/activities/httptool"
	"github.com/tempiex/worker/activities/shell"
	"github.com/tempiex/worker/activity"
	"github.com/tempiex/worker/api"
	"github.com/tempiex/worker/internal/config"
	"github.com/tempiex/worker/metrics"
	"github.com/tempiex/worker/run"
)

const shutdownTimeout = 30 * time.Second

func main() {
	configPath := flag.String("config", "config/development.yaml", "path to config file")
	flag.Parse()

	if err := runWorker(context.Background(), *configPath); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func runWorker(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := buildLogger(cfg.Log)
	log.Logger = logger

	m, err := metrics.Setup(ctx, cfg.Metrics.OTel.Endpoint, cfg.Metrics.OTel.Insecure, logger)
	if err != nil {
		return fmt.Errorf("metrics setup: %w", err)
	}

	store, err := run.Open(cfg.Runs.StorePath)
	if err != nil {
		return fmt.Errorf("run store: %w", err)
	}
	defer store.Close()

	bus := run.NewBus()
	bus.OnDrop = func(runID string) {
		m.EventsDropped.Add(ctx, 1, metric.WithAttributes(attribute.String("subscriber", "sse")))
	}
	runs := run.NewManager(store, bus, cfg.Runs.MaxCompletedRuns)

	registry := activity.NewRegistry()
	registerBuiltins(registry)

	var dialOpts []grpc.DialOption
	if !cfg.Tempiex.TLS.Enabled {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(cfg.Tempiex.Address, dialOpts...)
	if err != nil {
		return fmt.Errorf("grpc connect: %w", err)
	}
	defer conn.Close()

	client := workflowservicev1.NewWorkflowServiceClient(conn)

	exec := activity.NewExecutor(registry, runs, m, m.Tracer(), cfg.Tempiex.Namespace, logger)

	// worker.Config is the library's own Go API type, deliberately separate
	// from this binary's YAML schema; translate field-for-field.
	workerCfgs := make([]tempiexworker.Config, len(cfg.Workers))
	for i, w := range cfg.Workers {
		workerCfgs[i] = tempiexworker.Config{
			TaskQueue:                  w.TaskQueue,
			MaxConcurrentActivities:    w.MaxConcurrentActivities,
			MaxConcurrentWorkflowTasks: w.MaxConcurrentWorkflowTasks,
		}
	}
	pool := tempiexworker.New(workerCfgs, client, exec, logger)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool.Start(ctx)
	logger.Info().
		Str("address", cfg.Tempiex.Address).
		Str("namespace", cfg.Tempiex.Namespace).
		Int("queues", len(workerCfgs)).
		Msg("worker polling started")

	srv, err := startAPI(cfg, runs, registry, client, logger)
	if err != nil {
		pool.Stop()
		return err
	}

	<-ctx.Done()
	logger.Info().Msg("shutting down")

	// Stop polling first, then wait for in-flight activities to finish.
	pool.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if srv != nil {
		_ = srv.Shutdown(shutdownCtx)
	}
	_ = m.Shutdown(shutdownCtx)

	return nil
}

func startAPI(
	cfg *config.Config,
	runs *run.Manager,
	registry *activity.Registry,
	client workflowservicev1.WorkflowServiceClient,
	logger zerolog.Logger,
) (*http.Server, error) {
	if !cfg.API.Enabled {
		return nil, nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	router := api.NewRouter(api.Options{
		Runs:        runs,
		Registry:    registry,
		Client:      client,
		Namespace:   cfg.Tempiex.Namespace,
		CORSOrigins: cfg.API.CORSOrigins,
		Log:         logger,
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}

	srv := &http.Server{Handler: router, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		logger.Info().Str("addr", addr).Msg("HTTP API listening")
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("HTTP server error")
		}
	}()
	return srv, nil
}

// registerBuiltins registers the activities shipped with the reference binary.
// echo exists purely as a smoke test: schedule an "echo" activity and confirm
// the result payload matches the input.
func registerBuiltins(registry *activity.Registry) {
	activity.Register(registry, activity.Activity[map[string]any, map[string]any]{
		Name:          "echo",
		Description:   "Return the input unchanged",
		ExecutionMode: "parallel",
		Execute: func(_ activity.Context, input map[string]any) (map[string]any, error) {
			return input, nil
		},
	})
	registry.RegisterAny(shell.New())
	registry.RegisterAny(httptool.New())
	registry.RegisterAny(file.NewReadActivity())
	registry.RegisterAny(file.NewWriteActivity())
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
