package report

import (
	"context"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"github.com/rs/zerolog"
	"soloway/internal/config"
	repository "soloway/internal/repository/report"
)

type IService interface {
	SendReportToStorage(ctx context.Context, req *pb.SendReportToStorageRequest, repo repository.IRepository) (*pb.SendReportToStorageResponse, error)

	GetLogger() *zerolog.Logger
	GetConfig() *config.Configuration
}

type Service struct {
	logger *zerolog.Logger
	cfg    *config.Configuration
}

func NewService(logger *zerolog.Logger, cfg *config.Configuration) IService {
	return &Service{
		logger: logger,
		cfg:    cfg,
	}
}

// GetLogger is a method of business logic layer that gets a logger for logging events in a upper layer.
func (s *Service) GetLogger() *zerolog.Logger {
	return s.logger
}

// GetConfig is a method of business logic layer that gets a configuration from an upper layer.
func (s *Service) GetConfig() *config.Configuration {
	return s.cfg
}
