package report

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"github.com/rs/zerolog"
	solowaysdk "github.com/zfullio/soloway-sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"soloway/internal/config"
	repository "soloway/internal/repository/report"
	"soloway/pkg/converters"
)

const (
	getClientTimeout = 10 * time.Minute
)

type SendDataCfg struct {
	DateStart   time.Time
	DateFinish  time.Time
	Destination repository.Destination
	BucketName  string
	Filename    string
	ClientName  string
}

func (s *Service) PushPlacementStatByDayToBQ(ctx context.Context, req *pb.PushPlacementStatByDayToBQRequest, repo repository.IRepository) (*pb.PushPlacementStatByDayToBQResponse, error) {
	serviceLogger := s.logger.With().Str("Source", "PushPlacementStatByDayToBQ").Logger()

	var dateStart, dateFinish time.Time

	var err error

	if req.GetPeriod() != nil {
		dateStart, err = time.Parse(time.DateOnly, req.Period.DateFrom)
		if err != nil {
			msg := fmt.Sprintf("can't parse date from: %v", err)
			serviceLogger.Error().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}

		dateFinish, err = time.Parse(time.DateOnly, req.Period.DateTill)
		if err != nil {
			msg := fmt.Sprintf("can't parse date to: %v", err)
			serviceLogger.Error().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "period is required")
	}

	users, err := repo.GetUsers(ctx, req.GsConfig.SpreadsheetId)
	if err != nil {
		msg := fmt.Sprintf("ошибка получения пользователей: %v", err)
		serviceLogger.Error().Err(err).Msg(msg)

		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if len(users) == 0 {
		msg := "users not found"
		serviceLogger.Error().Msg(msg)

		return nil, status.Error(codes.NotFound, msg)
	}

	var warnings []string

	var bucketFiles []string

	dateUpdate := time.Now()

	params := make([]SendDataCfg, 0, len(users))
	destination := repository.Destination{
		ProjectID: req.BqConfig.ProjectId,
		DatasetID: req.BqConfig.DatasetId,
		TableID:   req.BqConfig.TableId,
	}

	transport := http.Client{
		Timeout: getClientTimeout,
	}

	for _, user := range users {
		userLogger := serviceLogger.With().Str("user", user.Name).Logger()
		solConfig := config.Soloway{
			UserName: user.Login,
			Password: user.Password,
		}

		solClient := solowaysdk.NewClient(transport, solConfig.UserName, solConfig.Password)

		err = solClient.Login(ctx)
		if err != nil {
			msg := fmt.Sprintf("can't login: %v", err)
			userLogger.Error().Err(err).Msg(msg)
			warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))

			continue
		}

		err = solClient.Whoami(ctx)
		if err != nil {
			msg := fmt.Sprintf("can't whoami: %v", err)
			userLogger.Error().Err(err).Msg(msg)
			warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))

			continue
		}

		userLogger.Info().Msg("collect stat")

		stat, err := repo.GetStatPlacementByDay(ctx, solClient, dateStart, dateFinish)
		if err != nil {
			msg := fmt.Sprintf("can't get stat: %v", err)
			userLogger.Error().Err(err).Msg(msg)
			warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))

			continue
		}

		filename, err := converters.GeneratePlacementStatByDayJSON(s.cfg.AttachmentsDir, stat, user.Login, dateUpdate)
		if err != nil {
			msg := fmt.Sprintf("can't generate json: %v", err)
			userLogger.Error().Err(err).Msg(msg)
			warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))

			continue
		}

		bucketFiles = append(bucketFiles, filename)
		cfg := SendDataCfg{
			DateStart:   dateStart,
			DateFinish:  dateFinish,
			Destination: destination,
			BucketName:  req.CsConfig.BucketName,
			Filename:    filename,
			ClientName:  user.Name,
		}

		params = append(params, cfg)
	}

	defer clearBucketFiles(serviceLogger, bucketFiles)
	for i := 0; i < len(params); i++ {
		select {
		case <-ctx.Done():
			serviceLogger.Warn().Msg("context canceled")

			return nil, status.Error(codes.Canceled, "canceled")
		default:
			serviceLogger.Info().Str("user", params[i].ClientName).Msg("sending")

			err = repo.SendFromStorage(ctx, params[i].Destination, params[i].DateStart, params[i].DateFinish, params[i].BucketName, params[i].Filename, params[i].ClientName)
			if err != nil {
				msg := fmt.Sprintf("can't send to bq: %v", err)
				serviceLogger.Error().Err(err).Str("user", params[i].ClientName).Msg(msg)

				warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", params[i].ClientName, msg))

				continue
			}
		}
	}

	return &pb.PushPlacementStatByDayToBQResponse{Warnings: warnings}, nil
}

func clearBucketFiles(logger zerolog.Logger, files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			logger.Error().Err(err).Msg("can't remove file")
		}
	}
}
