package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	"github.com/tempiex/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tempiex/pi/internal/activity"
	"github.com/tempiex/pi/internal/api"
	"github.com/tempiex/pi/internal/config"
	"github.com/tempiex/pi/internal/event"
	"github.com/tempiex/pi/internal/metrics"
	"github.com/tempiex/pi/internal/session"
	"github.com/tempiex/pi/internal/tool"
)

type Options struct {
	Config   *config.Config
	Registry *tool.Registry
}

type Agent struct {
	opts Options
}

func New(opts Options) *Agent {
	return &Agent{opts: opts}
}

func (a *Agent) Run(ctx context.Context) error {
	cfg := a.opts.Config

	logger := buildLogger(cfg.Log)
	log.Logger = logger

	m, err := metrics.Setup(ctx, cfg.Metrics, logger)
	if err != nil {
		return fmt.Errorf("metrics setup: %w", err)
	}

	store, err := session.Open(cfg.Session.StorePath)
	if err != nil {
		return fmt.Errorf("session store: %w", err)
	}
	defer store.Close()
	mgr := session.NewManager(store)

	bus := event.New()

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

	exec := activity.NewExecutor(
		a.opts.Registry,
		bus,
		mgr,
		m,
		m.Tracer(),
		logger,
	)

	// The generic worker.Pool speaks worker.Config, not pi's own
	// config.WorkerConfig — translate field-for-field so pi's YAML schema
	// stays independent of the worker module's Go API.
	workerCfgs := make([]worker.Config, len(cfg.Workers))
	for i, w := range cfg.Workers {
		workerCfgs[i] = worker.Config{
			TaskQueue:                  w.TaskQueue,
			MaxConcurrentActivities:    w.MaxConcurrentActivities,
			MaxConcurrentWorkflowTasks: w.MaxConcurrentWorkflowTasks,
		}
	}
	pool := worker.New(workerCfgs, client, exec, logger)

	router := api.NewRouter(mgr, bus, a.opts.Registry, client, cfg.API.CORSOrigins, logger)
	addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool.Start(ctx)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	go func() {
		logger.Info().Str("addr", addr).Msg("HTTP API listening")
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("shutting down")

	pool.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	_ = m.Shutdown(shutdownCtx)

	return nil
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
