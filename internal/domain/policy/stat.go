package policy

import (
	"Soloway/internal/domain/service"
	"Soloway/internal/domain/usecase/stat"
	"fmt"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
	"time"
)

type StatPolicy struct {
	PlacementService service.PlacementService
	StatService      service.StatService
	logger           *zerolog.Logger
}

func NewStatPolicy(placementService service.PlacementService, statService service.StatService, logger *zerolog.Logger) *StatPolicy {
	policyLogger := logger.With().Str("policy", "placement").Logger()

	return &StatPolicy{
		PlacementService: placementService,
		StatService:      statService,
		logger:           &policyLogger,
	}
}

func (sp StatPolicy) PushPlacementStatByDayToBQ(ctx context.Context, clientName string, startDate time.Time, stopDate time.Time) error {
	sp.logger.Trace().Msg("PushPlacementStatByDayToBQ")

	statUseCase := stat.NewUseCase(sp.StatService, sp.PlacementService, sp.logger)

	err := statUseCase.PushPlacementStatByDayToBQ(ctx, clientName, startDate, stopDate)
	if err != nil {
		return fmt.Errorf("usecase.PushPlacementStatByDayToBQ: %w", err)
	}

	return nil
}
