package service

import (
	"Soloway/internal/domain/entity"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"time"
)

type SolStatRepository interface {
	GetPlacementStatByDay(ctx context.Context, placement *entity.Placement, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error)
}

type StatRepositoryBQ interface {
	SendPlacementStatByDay(ctx context.Context, client string, stat []entity.StatPlacement) (err error)
	DeleteByDateColumn(ctx context.Context, client string, dateStart time.Time, dateFinish time.Time) (err error)
	CreateTable(ctx context.Context) (err error)
	TableExists(ctx context.Context) (err error)
}

type StatService struct {
	solRepo SolStatRepository
	bqRepo  StatRepositoryBQ
	logger  *zerolog.Logger
}

func NewStatService(solRepo SolStatRepository, bqRepo StatRepositoryBQ, logger *zerolog.Logger) *StatService {
	serviceLogger := logger.With().Str("service", "stat").Logger()

	return &StatService{
		solRepo: solRepo,
		bqRepo:  bqRepo,
		logger:  &serviceLogger,
	}
}

func (s *StatService) GetStatPlacementByDay(ctx context.Context, placement *entity.Placement, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error) {
	stat, err = s.solRepo.GetPlacementStatByDay(ctx, placement, startDate, stopDate)
	if err != nil {
		return stat, fmt.Errorf("cannot get stat placement by day from Comagic: %w", err)
	}

	return stat, nil
}

func (s *StatService) PushStatPlacementByDayToBQ(ctx context.Context, client string, stat []entity.StatPlacement,
	dateFrom time.Time, dateTill time.Time) error {
	s.logger.Trace().Str("client", client).Msg("PushStatPlacementByDayToBQ")

	err := s.bqRepo.TableExists(ctx)
	if err != nil {
		err = s.bqRepo.CreateTable(ctx)
		if err != nil {
			return fmt.Errorf("ошибка создания bq таблицы: %w", err)
		}
	}

	s.logger.Info().Msgf("%s :Удаление за %s -- %s", client, dateFrom.Format(time.DateOnly), dateTill.Format(time.DateOnly))

	errBq := s.bqRepo.DeleteByDateColumn(ctx, client, dateFrom, dateTill)
	if errBq != nil {
		return fmt.Errorf("can`t delete by date column")
	}

	err = s.bqRepo.SendPlacementStatByDay(ctx, client, stat)
	if err != nil {
		return fmt.Errorf("cannot send stat placement by day to BQ: %w", err)
	}

	return nil
}
