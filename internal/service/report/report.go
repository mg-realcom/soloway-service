package report

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"soloway/internal/entity"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"github.com/rs/zerolog"
	solowaysdk "github.com/zfullio/soloway-sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"soloway/internal/config"
	repository "soloway/internal/repository/report"
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

func (s *Service) SendReportToStorage(ctx context.Context, req *pb.SendReportToStorageRequest, repo repository.IRepository) (*pb.SendReportToStorageResponse, error) {
	serviceLogger := s.logger.With().Str("Source", "SendReportToStorage").Logger()

	serviceLogger.Info().Msg("in progress")

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

	var files []entity.File

	transport := http.Client{
		Timeout: getClientTimeout,
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(users))

	mux := &sync.Mutex{}

	for _, user := range users {
		go func(user entity.User) {
			defer wg.Done()

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
				mux.Lock()
				warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))
				mux.Unlock()

				return
			}

			err = solClient.Whoami(ctx)
			if err != nil {
				msg := fmt.Sprintf("can't whoami: %v", err)
				userLogger.Error().Err(err).Msg(msg)
				mux.Lock()
				warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))
				mux.Unlock()

				return
			}

			userLogger.Info().Msg("collect stat")

			gotFiles, err := repo.GetStatPlacementByDay(ctx, solClient, dateStart, dateFinish, s.cfg.AttachmentsDir)
			if err != nil {
				msg := fmt.Sprintf("can't get stat: %v", err)
				userLogger.Error().Err(err).Msg(msg)
				mux.Lock()
				warnings = append(warnings, fmt.Sprintf("clent `%s`: %v", user.Name, msg))
				mux.Unlock()

				return
			}

			mux.Lock()
			files = append(files, gotFiles...)
			mux.Unlock()
		}(user)
	}

	wg.Wait()

	tempFiles := addTempFiles(files)

	defer clearTempFiles(serviceLogger, tempFiles)

	dates := make(map[time.Time]bool)

	for _, file := range files {
		if dates[file.Date] {
			continue
		} else {
			dates[file.Date] = true
		}
	}

	for date := range dates {
		err := repo.StorageClearByDate(ctx, req.Storage.GetYandexStorage().GetFolderName(), req.Storage.GetYandexStorage().GetBucketName(), date)
		if err != nil {
			msg := fmt.Sprintf("failed to clear old files: %v", err)
			serviceLogger.Error().Msg(msg)

			return nil, status.Error(codes.Internal, msg)
		}
	}

	for _, file := range files {
		err := repo.UploadToStorage(ctx, req.Storage.GetYandexStorage().GetFolderName(), req.Storage.GetYandexStorage().GetBucketName(), file.Path, file.Date)
		if err != nil {
			msg := fmt.Sprintf("failed to send report: %v", err)
			serviceLogger.Error().Msg(msg)

			return nil, status.Error(codes.Internal, msg)
		}
	}

	return &pb.SendReportToStorageResponse{Warnings: warnings}, nil
}

func clearTempFiles(logger zerolog.Logger, files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			logger.Error().Err(err).Msg("can't remove file")
		}
	}
}

func addTempFiles(files []entity.File) []string {
	bucketFiles := make([]string, 0, len(files))
	for _, file := range files {
		bucketFiles = append(bucketFiles, file.Path)
	}
	return bucketFiles
}
