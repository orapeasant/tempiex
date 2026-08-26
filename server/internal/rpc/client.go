package rpc

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tempiex/server/internal/config"

	wfsvcv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"
)

func NewWorkflowServiceClient(cfg *config.Config) (wfsvcv1.WorkflowServiceClient, error) {
	conn, err := grpc.NewClient(cfg.TempiexGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Info().Str("addr", cfg.TempiexGrpcAddress).Msg("connected to tempiex gRPC")
	return wfsvcv1.NewWorkflowServiceClient(conn), nil
}
