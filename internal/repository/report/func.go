package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"soloway/pkg/tracing"

	"cloud.google.com/go/civil"
	solowaysdk "github.com/zfullio/soloway-sdk"
	"golang.org/x/sync/errgroup"
	"soloway/internal/entity"
)

const readRange = "// Config!A2:C"

func (r Repository) SendFromStorage(ctx context.Context, destination Destination, dateStart, dateFinish time.Time, bucketName string, file string, clientName string) (err error) {
	repoLogger := r.logger.With().Str("Source", "SendFromStorage").Str("type", "bq").Logger()

	ctx, span := tracing.CreateSpan(ctx, "bq", "SendFromStorage")

	tracing.SetSpanAttribute(span, tracing.AttributeUser, clientName)
	tracing.SetSpanAttribute(span, "file", file)

	schema := PlacementStatDTO{}

	err = r.storage.SendFile(ctx, file)
	if err != nil {
		msg := "error send file"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return fmt.Errorf("error send storage: %w", err)
	}

	err = r.bd.TableExists(ctx, destination)
	if err != nil {
		err = r.bd.CreateTable(ctx, destination, schema)
		if err != nil {
			msg := "error create BQ table"
			repoLogger.Error().Err(err).Msg(msg)
			tracing.EndSpanError(span, err, msg, true)

			return fmt.Errorf("creation BQ table error: %w", err)
		}
	} else {
		err = r.bd.DeleteByDateColumn(ctx, destination, clientName, "date", dateStart, dateFinish)
		if err != nil {
			msg := "error delete bq"
			repoLogger.Error().Err(err).Msg(msg)
			tracing.EndSpanError(span, err, msg, true)

			return fmt.Errorf("error delete bq: %w", err)
		}
	}

	err = r.bd.ImportFromCS(ctx, destination, bucketName, file, schema)
	if err != nil {
		msg := "error push to BQ from storage"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return fmt.Errorf("error push to BQ from storage: %w", err)
	}

	tracing.EndSpanOk(span, "SendFromStorage", true)

	return nil
}

func (r Repository) GetUsers(ctx context.Context, spreadsheetID string) ([]entity.User, error) {
	repoLogger := r.logger.With().Str("Source", "GetUsers").Str("type", "gsheets").Logger()

	ctx, span := tracing.CreateSpan(ctx, "gsheets", "GetUsers")

	tracing.SetSpanAttribute(span, "spreadsheetID", spreadsheetID)

	resp, err := r.spreadsheetSrv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		msg := "api error"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	users := make([]entity.User, 0, len(resp.Values))

	for _, row := range resp.Values {
		select {
		case <-ctx.Done():
			tracing.EndSpanError(span, err, "context done", true)

			return nil, ctx.Err()
		default:
			client, ok := row[0].(string)
			if !ok {
				msg := "can't get: 'Клиент' in gsheets"
				repoLogger.Error().Msg(msg)
				tracing.EndSpanError(span, err, msg, true)

				return nil, errors.New(msg)
			}

			client = strings.ToLower(client)

			loginStr, ok := row[1].(string)
			if !ok {
				msg := "can't get: 'Логин' in gsheets"
				repoLogger.Error().Msg(msg)
				tracing.EndSpanError(span, err, msg, true)

				return nil, errors.New(msg)
			}

			login := strings.ToLower(loginStr)

			passStr, ok := row[2].(string)
			if !ok {
				msg := "can't get: 'Пароль' in gsheets"
				repoLogger.Error().Msg(msg)
				tracing.EndSpanError(span, err, msg, true)

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

	tracing.EndSpanOk(span, "GetUsers", true)

	return users, nil
}

func (r Repository) GetStatPlacementByDay(ctx context.Context, client *solowaysdk.Client, startDate time.Time, stopDate time.Time) (stat []entity.StatPlacement, err error) {
	repoLogger := r.logger.With().Str("Source", "GetStatPlacementByDay").Str("type", "soloway-api").Str("username", client.Username).Logger()

	ctx, span := tracing.CreateSpan(ctx, "soloway-api", "GetStatPlacementByDay")

	tracing.SetSpanAttribute(span, tracing.AttributeUser, client.Username)

	data, err := client.GetPlacements(ctx)
	if err != nil {
		msg := "can't get placements"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

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
		tracing.EndSpanError(span, err, err.Error(), true)

		return nil, err
	}

	tracing.EndSpanOk(span, "GetStatPlacementByDay", true)

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
