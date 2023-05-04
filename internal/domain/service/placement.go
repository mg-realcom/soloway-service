package service

import (
	"Soloway/internal/domain/entity"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zfullio/soloway-sdk"
)

type PlacementsRepository interface {
	GetAll(ctx context.Context) ([]entity.Placement, error)
}

type PlacementService struct {
	client solowaysdk.Client
	repo   PlacementsRepository
	logger *zerolog.Logger
}

func NewPlacementService(solowayClient solowaysdk.Client, repo PlacementsRepository, logger *zerolog.Logger) *PlacementService {
	serviceLogger := logger.With().Str("service", "placement").Logger()

	return &PlacementService{
		client: solowayClient,
		repo:   repo,
		logger: &serviceLogger,
	}
}

func (ps *PlacementService) GetPlacements(ctx context.Context) (placements []entity.Placement, err error) {
	ps.logger.Trace().Msg("GetPlacements")

	placements, err = ps.repo.GetAll(ctx)
	if err != nil {
		return placements, fmt.Errorf("PlacementService.GetPlacements: %w", err)
	}

	return placements, nil
}
