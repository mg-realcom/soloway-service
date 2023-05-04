package sol

import (
	"Soloway/internal/domain/entity"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zfullio/soloway-sdk"
	"log"
	"time"
)

type StatRepository struct {
	client solowaysdk.Client
	logger *zerolog.Logger
}

func NewStatRepository(client solowaysdk.Client, logger *zerolog.Logger) *StatRepository {
	repoLogger := logger.With().Str("repository", "stat").Str("storage", "soloway").Logger()

	return &StatRepository{client: client, logger: &repoLogger}
}

func (pr StatRepository) GetPlacementStatByDay(ctx context.Context, placement *entity.Placement, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error) {
	pr.logger.Trace().Str("placement", placement.GUID).Msg("GetPlacementStatByDay")

	rawStat, err := pr.client.GetPlacementStatByDay(ctx, placement.GUID, startDate, stopDate)
	if err != nil {
		return stat, fmt.Errorf("StatRepository.GetPlacementStatByDay: %w", err)
	}

	for _, item := range rawStat.List {
		stat = append(stat, *PlacementStatFromDTO(item, *placement))
	}

	return stat, err
}

func PlacementStatFromDTO(pStat solowaysdk.PerformanceStat, placement entity.Placement) *entity.StatPlacement {
	date, err := time.Parse(time.DateOnly, pStat.Date)
	if err != nil {
		log.Println(err)
	}

	return &entity.StatPlacement{
		Clicks:        pStat.Clicks,
		Cost:          pStat.Cost,
		PlacementID:   pStat.PlacementID,
		PlacementName: placement.Name,
		Exposures:     pStat.Exposures,
		Date:          date,
	}
}
