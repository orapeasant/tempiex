package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"github.com/tempiex/server/internal/config"
	"github.com/tempiex/server/internal/rpc"
	"github.com/tempiex/server/internal/server"

	wfsvcv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			newDBPool,
			rpc.NewWorkflowServiceClient,
			server.New,
		),
		fx.Invoke(startServer),
	).Run()
}

func newDBPool(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DbDsn)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("connected to database")
	return pool, nil
}

func startServer(lc fx.Lifecycle, e *echo.Echo, cfg *config.Config, _ wfsvcv1.WorkflowServiceClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Start(e, cfg); err != nil {
					log.Error().Err(err).Msg("server stopped")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})
}
