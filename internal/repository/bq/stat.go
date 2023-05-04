package bq

import (
	"Soloway/internal/domain/entity"
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"time"
)

type StatRepository struct {
	db     *bigquery.Client
	table  *bigquery.Table
	logger *zerolog.Logger
}

func NewStatRepository(client *bigquery.Client, datasetID string, tableID string, logger *zerolog.Logger) *StatRepository {
	repoLogger := logger.With().Str("repository", "placement").Str("storage", "bq").Logger()

	dataset := client.Dataset(datasetID)
	table := dataset.Table(tableID)

	return &StatRepository{
		db:     client,
		table:  table,
		logger: &repoLogger,
	}
}

type PlacementStatDTO struct {
	Client        string     `bigquery:"client_name"`
	Clicks        int        `bigquery:"clicks"`
	Cost          int        `bigquery:"cost"`
	PlacementID   string     `bigquery:"placement_id"`
	PlacementName string     `bigquery:"placement_name"`
	Exposures     int        `bigquery:"exposures"`
	Date          civil.Date `bigquery:"date"`
	DateUpdate    time.Time  `bigquery:"date_update"`
}

func PlacementStatDTOFromEntity(client string, placementStat entity.StatPlacement, dateUpdate time.Time) *PlacementStatDTO {
	return &PlacementStatDTO{
		Client:        client,
		Clicks:        placementStat.Clicks,
		Cost:          placementStat.Cost,
		PlacementID:   placementStat.PlacementID,
		PlacementName: placementStat.PlacementName,
		Exposures:     placementStat.Exposures,
		Date:          civil.DateOf(placementStat.Date),
		DateUpdate:    dateUpdate,
	}
}

func (sr StatRepository) SendPlacementStatByDay(ctx context.Context, client string, stat []entity.StatPlacement) error {
	sr.logger.Trace().Msg("SendPlacementStatByDay")

	result := make([]PlacementStatDTO, 0, len(stat))
	dateUpdate := time.Now()

	for _, item := range stat {
		conv := PlacementStatDTOFromEntity(client, item, dateUpdate)
		result = append(result, *conv)
	}

	u := sr.table.Inserter()
	if err := u.Put(ctx, result); err != nil {
		return fmt.Errorf("failed to insert data into table %s.%s: %w", sr.table.DatasetID, sr.table.TableID, err)
	}

	return nil
}

func (sr StatRepository) DeleteByDateColumn(ctx context.Context, client string, dateStart time.Time, dateFinish time.Time) error {
	dateColumn := "date"

	err := DeleteByDateColumn(ctx, sr.db, sr.table, client, dateColumn, dateStart, dateFinish)
	if err != nil {
		return fmt.Errorf("DeleteByDateColumn: %w", err)
	}

	return nil
}

func (sr StatRepository) TableExists(ctx context.Context) error {
	sr.logger.Trace().Msg("TableExists")

	err := TableExists(ctx, sr.table)
	if err != nil {
		return err
	}

	return nil
}

func (sr StatRepository) CreateTable(ctx context.Context) error {
	sr.logger.Trace().Msg("createTable")

	err := CreateTable(ctx, PlacementStatDTO{}, sr.table, nil, nil)
	if err != nil {
		return fmt.Errorf("createTable: %w", err)
	}

	return nil
}
