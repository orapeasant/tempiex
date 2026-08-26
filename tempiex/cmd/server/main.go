package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
	"github.com/tempiex/tempiex/internal/config"
	"github.com/tempiex/tempiex/internal/frontend"
	"github.com/tempiex/tempiex/internal/history"
	"github.com/tempiex/tempiex/internal/matching"
	"github.com/tempiex/tempiex/internal/persistence"
	"github.com/tempiex/tempiex/internal/persistence/postgresql"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := fx.New(
		fx.Provide(
			loadConfig,
			newStore,
			newMatchingEngine,
			newHistoryEngine,
			newHandler,
			newGRPCServer,
		),
		fx.Invoke(startServer),
	)
	app.Run()
}

func loadConfig() (*config.Config, error) {
	path := os.Getenv("TEMPIEX_CONFIG")
	if path == "" {
		path = "config/development.yaml"
	}
	return config.Load(path)
}

func newStore(cfg *config.Config, lc fx.Lifecycle) (persistence.Store, error) {
	store, err := postgresql.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info().Msg("initializing database schema")
			return store.InitSchema(ctx)
		},
	})
	return store, nil
}

func newMatchingEngine(store persistence.Store) *matching.Engine {
	return matching.NewEngine(store)
}

func newHistoryEngine(store persistence.Store, match *matching.Engine) *history.Engine {
	return history.NewEngine(store, match)
}

func newHandler(hist *history.Engine, match *matching.Engine, store persistence.Store) *frontend.Handler {
	return frontend.NewHandler(hist, match, store)
}

func newGRPCServer(handler *frontend.Handler) *grpc.Server {
	srv := grpc.NewServer()
	workflowservicev1.RegisterWorkflowServiceServer(srv, handler)

	// Health check
	healthSvc := health.NewServer()
	healthSvc.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, healthSvc)

	reflection.Register(srv)
	return srv
}

type serverParams struct {
	fx.In
	Config *config.Config
	Server *grpc.Server
	LC     fx.Lifecycle
}

func startServer(p serverParams) {
	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := fmt.Sprintf(":%d", p.Config.Server.GRPCPort)
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", addr, err)
			}

			go func() {
				log.Info().Str("addr", addr).Msg("gRPC server listening")
				if err := p.Server.Serve(lis); err != nil {
					log.Error().Err(err).Msg("gRPC server stopped")
				}
			}()

			// Handle OS signals
			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt, syscall.SIGTERM)
				<-c
				p.Server.GracefulStop()
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Server.GracefulStop()
			return nil
		},
	})
}
