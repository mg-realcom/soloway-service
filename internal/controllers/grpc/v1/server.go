package v1

import (
	"Soloway/internal/config"
	"Soloway/internal/domain/policy"
	"Soloway/internal/domain/service"
	"Soloway/internal/repository/bq"
	"Soloway/internal/repository/sol"
	"Soloway/pb"
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zfullio/soloway-sdk"
	"google.golang.org/api/option"
)

type Server struct {
	cfg    config.ServerConfig
	logger *zerolog.Logger
	pb.UnimplementedSolowayServiceServer
}

func NewServer(cfg config.ServerConfig, logger *zerolog.Logger, srv pb.UnimplementedSolowayServiceServer) *Server {
	apiLogger := logger.With().Str("api", "grpc").Logger()

	return &Server{
		cfg:                               cfg,
		logger:                            &apiLogger,
		UnimplementedSolowayServiceServer: srv,
	}
}

func (s Server) initPolicy(ctx context.Context, solConfig config.Soloway, bqConfig config.BQ) (*policy.StatPolicy, error) {
	bqClient, err := bigquery.NewClient(ctx, bqConfig.ProjectID, option.WithCredentialsFile(bqConfig.ServiceKeyPath))
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации bq client: %w", err)
	}

	bqStatRepo := bq.NewStatRepository(bqClient, bqConfig.DatasetID, bqConfig.TableID, s.logger)

	solClient := solowaysdk.NewClient(solConfig.UserName, solConfig.Password)

	err = solClient.Login()
	if err != nil {
		return nil, fmt.Errorf("ошибка аутентификации Soloway: %w", err)
	}

	err = solClient.Whoami(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения инфо о клиенте Soloway: %w", err)
	}

	solPlacementRepo := sol.NewPlacementRepository(*solClient, s.logger)
	solStatRepo := sol.NewStatRepository(*solClient, s.logger)

	statSrv := service.NewStatService(*solStatRepo, bqStatRepo, s.logger)
	placementSrv := service.NewPlacementService(*solClient, *solPlacementRepo, s.logger)
	policyStat := policy.NewStatPolicy(*placementSrv, *statSrv, s.logger)

	return policyStat, nil
}
