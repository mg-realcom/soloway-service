package stat

import (
	"Soloway/internal/domain/entity"
	"Soloway/internal/domain/service"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"time"
)

type UseCase struct {
	statSrv      service.StatService
	userSrv      service.UserService
	placementSrv service.PlacementService
	logger       *zerolog.Logger
}

func NewUseCase(statSrv service.StatService, placementSrv service.PlacementService, logger *zerolog.Logger) *UseCase {
	useCaseLogger := logger.With().Str("useCase", "stat").Logger()

	return &UseCase{statSrv: statSrv, placementSrv: placementSrv, logger: &useCaseLogger}
}

func (su UseCase) GetStatPlacementByDay(ctx context.Context, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error) {
	su.logger.Trace().Str("startDate", startDate.Format(time.RFC3339)).Str("stopDate", stopDate.Format(time.RFC3339)).Msg("GetStatPlacementByDay")

	placements, err := su.placementSrv.GetPlacements(ctx)
	if err != nil {
		return stat, fmt.Errorf("cannot get placements from soloway: %w", err)
	}

	for _, placement := range placements {
		placement := placement

		placementStat, err := su.statSrv.GetStatPlacementByDay(ctx, &placement, startDate, stopDate)
		if err != nil {
			return stat, fmt.Errorf("cannot get stat placement by day from soloway: %w", err)
		}

		stat = append(stat, placementStat...)
	}

	return stat, nil
}

func (su UseCase) PushPlacementStatByDayToBQ(ctx context.Context, client string, startDate time.Time, stopDate time.Time) error {
	su.logger.Trace().Str("client", client).Str("startDate", startDate.Format(time.RFC3339)).Str("stopDate", stopDate.Format(time.RFC3339)).Msg("PushPlacementStatByDayToBQ")

	stat, err := su.GetStatPlacementByDay(ctx, startDate, stopDate)
	if err != nil {
		return fmt.Errorf("cannot get stat placement by day from soloway: %w", err)
	}

	err = su.statSrv.Send(ctx, client, startDate, stopDate, stat)
	if err != nil {
		return fmt.Errorf("ошибка отправки статистики по дням в BQ: %w", err)
	}

	return nil
}
