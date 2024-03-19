package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	solowaysdk "github.com/zfullio/soloway-sdk"
	"golang.org/x/sync/errgroup"
	"soloway/internal/entity"
)

const readRange = "// Config!A2:C"

func (r Repository) SendFromStorage(ctx context.Context, destination Destination, dateStart, dateFinish time.Time, bucketName string, file string, clientName string) (err error) {
	repoLogger := r.logger.With().Str("Source", "SendFromStorage").Str("type", "bq").Logger()

	schema := PlacementStatDTO{}

	err = r.storage.SendFile(ctx, file)
	if err != nil {
		repoLogger.Error().Err(err).Msg("error send storage")

		return fmt.Errorf("error send storage: %w", err)
	}

	err = r.bd.TableExists(ctx, destination)
	if err != nil {
		err = r.bd.CreateTable(ctx, destination, schema)
		if err != nil {
			repoLogger.Error().Err(err).Msg("error create BQ table")

			return fmt.Errorf("creation BQ table error: %w", err)
		}
	} else {
		err = r.bd.DeleteByDateColumn(ctx, destination, clientName, "date", dateStart, dateFinish)
		if err != nil {
			repoLogger.Error().Err(err).Msg("error delete bq")

			return fmt.Errorf("error delete bq: %w", err)
		}
	}

	err = r.bd.ImportFromCS(ctx, destination, bucketName, file, schema)
	if err != nil {
		repoLogger.Error().Err(err).Msg("error push to BQ from storage")

		return fmt.Errorf("error push to BQ from storage: %w", err)
	}

	return nil
}

func (r Repository) GetUsers(ctx context.Context, spreadsheetID string) ([]entity.User, error) {
	repoLogger := r.logger.With().Str("Source", "GetUsers").Str("type", "gsheets").Logger()

	resp, err := r.spreadsheetSrv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		repoLogger.Error().Err(err).Msg("api error")

		return nil, fmt.Errorf("api error: %w", err)
	}

	users := make([]entity.User, 0, len(resp.Values))

	for _, row := range resp.Values {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			client, ok := row[0].(string)
			if !ok {
				msg := "can't get: 'Клиент' in gsheets"
				repoLogger.Error().Msg(msg)

				return nil, errors.New(msg)
			}

			client = strings.ToLower(client)

			loginStr, ok := row[1].(string)
			if !ok {
				msg := "can't get: 'Логин' in gsheets"
				repoLogger.Error().Msg(msg)

				return nil, errors.New(msg)
			}

			login := strings.ToLower(loginStr)

			passStr, ok := row[2].(string)
			if !ok {
				msg := "can't get: 'Пароль' in gsheets"
				repoLogger.Error().Msg(msg)

				return nil, errors.New(msg)
			}

			pass := passStr

			user := entity.User{
				Name:     client,
				Login:    login,
				Password: pass,
			}
			users = append(users, user)
		}
	}

	return users, nil
}

func (r Repository) GetStatPlacementByDay(ctx context.Context, client *solowaysdk.Client, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error) {
	repoLogger := r.logger.With().Str("Source", "GetStatPlacementByDay").Str("type", "soloway-api").Str("username", client.Username).Logger()

	data, err := client.GetPlacements(ctx)
	if err != nil {
		msg := "can't get placements"
		repoLogger.Error().Err(err).Msg(msg)

		return stat, errors.New(msg)
	}

	placements := make([]entity.Placement, 0, len(data.List))

	for i := 0; i < len(data.List); i++ {
		placements = append(placements, *newPlacement(data.List[i]))
	}

	g, ctx := errgroup.WithContext(ctx)
	statCh := make(chan []entity.StatPlacement, len(placements))

	for _, placement := range placements {
		g.Go(func() error {
			placementStat, err := client.GetPlacementStatByDay(ctx, placement.GUID, startDate, stopDate)
			if err != nil {
				msg := "can't get stat placement by day"
				repoLogger.Error().Err(err).Msg(msg)

				return errors.New(msg)
			}

			var data []entity.StatPlacement

			for _, item := range placementStat.List {
				data = append(data, *placementStatFromDTO(item, placement.Name))
			}

			statCh <- data

			return nil
		})
	}

	go func() {
		g.Wait()
		close(statCh)
	}()

	for i := range statCh {
		stat = append(stat, i...)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return stat, nil
}

func newPlacement(placement solowaysdk.Placement) *entity.Placement {
	return &entity.Placement{
		GUID: placement.Doc.GUID,
		Name: placement.Doc.Name,
	}
}

func placementStatFromDTO(pStat solowaysdk.PerformanceStat, placementName string) *entity.StatPlacement {
	date, err := time.Parse(time.DateOnly, pStat.Date)
	if err != nil {
		log.Println(err)
	}

	return &entity.StatPlacement{
		Clicks:        pStat.Clicks,
		Cost:          pStat.Cost,
		PlacementID:   pStat.PlacementID,
		PlacementName: placementName,
		Exposures:     pStat.Exposures,
		Date:          date,
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
