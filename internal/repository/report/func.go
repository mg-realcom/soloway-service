package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"soloway/pkg/tracing"

	solowaysdk "github.com/zfullio/soloway-sdk"
	"golang.org/x/sync/errgroup"
	"soloway/internal/entity"
)

const readRange = "// Config!A2:C"

func (r Repository) UploadToStorage(ctx context.Context, directory string, bucketName string, filePath string, date time.Time) error {
	method := "UploadToStorage"
	storage := "YandexObjectStorage"
	repoLogger := r.logger.With().Str("Source", method).Str("type", storage).Logger()

	ctx, span := tracing.CreateSpan(ctx, storage, method)

	tracing.SetSpanAttribute(span, "bucket", bucketName)
	tracing.SetSpanAttribute(span, "directory", directory)
	tracing.SetSpanAttribute(span, "file", filePath)
	tracing.SetSpanAttribute(span, "date", date.Format(time.DateOnly))

	err := r.storage.UploadFileWithDateDestination(ctx, bucketName, directory, filePath, date)
	if err != nil {
		msg := "error send file"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return fmt.Errorf("error upload file to storage: %w", err)
	}

	tracing.EndSpanOk(span, method, true)

	return nil
}

func (r Repository) StorageClearByDate(ctx context.Context, directory string, bucketName string, date time.Time) error {
	method := "StorageClearByDate"
	storage := "YandexObjectStorage"
	repoLogger := r.logger.With().Str("Source", method).Str("type", storage).Logger()

	ctx, span := tracing.CreateSpan(ctx, storage, method)

	tracing.SetSpanAttribute(span, "bucket", bucketName)
	tracing.SetSpanAttribute(span, "directory", directory)
	tracing.SetSpanAttribute(span, "date", date.Format(time.DateOnly))

	err := r.storage.DeleteFolderByDate(ctx, bucketName, directory, date)
	if err != nil {
		msg := "error delete folder"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return fmt.Errorf("error delete folder: %w", err)
	}

	tracing.EndSpanOk(span, method, true)

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

func (r Repository) GetStatPlacementByDay(ctx context.Context, client *solowaysdk.Client, startDate time.Time, stopDate time.Time, attachmentDir string) ([]entity.File, error) {
	repoLogger := r.logger.With().Str("Source", "GetStatPlacementByDay").Str("type", "soloway-api").Str("username", client.Username).Logger()

	ctx, span := tracing.CreateSpan(ctx, "soloway-api", "GetStatPlacementByDay")

	tracing.SetSpanAttribute(span, tracing.AttributeUser, client.Username)

	data, err := client.GetPlacements(ctx)
	if err != nil {
		msg := "can't get placements"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return nil, errors.New(msg)
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
				data = append(data, *placementStatFromDTO(item, placement.Name, client.Username))
			}

			statCh <- data

			return nil
		})
	}

	go func() {
		g.Wait()
		close(statCh)
	}()

	stat := make([]entity.StatPlacement, 0)

	for i := range statCh {
		stat = append(stat, i...)
	}

	if err := g.Wait(); err != nil {
		tracing.EndSpanError(span, err, err.Error(), true)

		return nil, err
	}

	filenames, err := GenerateReportPlacementStatJSON(attachmentDir, stat)
	if err != nil {
		msg := "can't generate report"
		repoLogger.Error().Err(err).Msg(msg)
		tracing.EndSpanError(span, err, msg, true)

		return nil, fmt.Errorf("GenerateReport: %w", err)
	}

	tracing.EndSpanOk(span, "GetStatPlacementByDay", true)

	return filenames, nil
}

func newPlacement(placement solowaysdk.Placement) *entity.Placement {
	return &entity.Placement{
		GUID: placement.Doc.GUID,
		Name: placement.Doc.Name,
	}
}

func placementStatFromDTO(pStat solowaysdk.PerformanceStat, placementName string, client string) *entity.StatPlacement {
	return &entity.StatPlacement{
		Client:        client,
		Clicks:        pStat.Clicks,
		Cost:          pStat.Cost,
		PlacementID:   pStat.PlacementID,
		PlacementName: placementName,
		Exposures:     pStat.Exposures,
		Date:          pStat.Date,
	}
}
