package v1

import (
	"Soloway/internal/config"
	"Soloway/internal/domain/service"
	"Soloway/internal/repository/gs"
	"Soloway/pb"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func (s Server) PushPlacementStatByDayToBQ(ctx context.Context, req *pb.PushPlacementStatByDayToBQRequest) (*pb.PushPlacementStatByDayToBQResponse, error) {
	methodLogger := s.logger.With().Str("method", "PushPlacementStatByDayToBQ").Logger()

	methodLogger.Info().Msg(msgMethodPrepared)

	defer methodLogger.Info().Msg(msgMethodFinished)

	bqServiceKey := s.cfg.KeysDir + "/" + req.BqConfig.ServiceKey
	gsServiceKey := s.cfg.KeysDir + "/" + req.GsConfig.ServiceKey

	if req.Period == nil {
		err := errors.New("wrong value in field 'period'")
		methodLogger.Error().Err(err).Msg(msgErrMethod)

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: false,
		}, err
	}

	dateFrom, err := pbDateNormalize(req.Period.DateFrom)
	if err != nil {
		methodLogger.Error().Err(err).Msg(msgErrMethod)

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: false,
		}, fmt.Errorf("wrong value in field 'dateFrom' : %w", err)
	}

	dateTill, err := pbDateNormalize(req.Period.DateTill)
	if err != nil {
		methodLogger.Error().Err(err).Msg(msgErrMethod)

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: false,
		}, fmt.Errorf("wrong value in field 'dateTill' : %w", err)
	}

	bqConfig := config.BQ{
		ServiceKeyPath: bqServiceKey,
		ProjectID:      req.BqConfig.ProjectId,
		DatasetID:      req.BqConfig.DatasetId,
		TableID:        req.BqConfig.TableId,
	}

	ghSrv, err := sheets.NewService(ctx, option.WithCredentialsFile(gsServiceKey))
	if err != nil {
		methodLogger.Error().Err(err).Msg(msgErrMethod)

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: false,
		}, fmt.Errorf("ошибка инициализации gs client")
	}

	userRepo := gs.NewUserRepository(*ghSrv, s.logger)
	userSrv := service.NewUserService(userRepo, s.logger)

	users, err := userSrv.GetAll(req.GsConfig.SpreadsheetId)
	if err != nil {
		methodLogger.Error().Err(err).Msg(msgErrMethod)

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: true,
		}, fmt.Errorf("ошибка получения пользователей")
	}

	if len(users) == 0 {
		methodLogger.Info().Msg("Ни один клиент не собран")

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk: false,
		}, fmt.Errorf("отсутсвуют пользователи в справочнике")
	}

	warnings := make([]string, 0, len(users))

	methodLogger.Info().Msg(msgMethodStarted)

	for _, user := range users {
		solConfig := config.Soloway{
			UserName: user.Login,
			Password: user.Password,
		}

		policyStat, err := s.initPolicy(ctx, solConfig, bqConfig)
		if err != nil {
			methodLogger.Warn().Err(err).Msg("ошибка инициализации policy")

			warnings = append(warnings, fmt.Sprintf("ошибка инициализации policy для %s", user.Name))

			continue
		}

		err = policyStat.PushPlacementStatByDayToBQ(ctx, user.Name, dateFrom, dateTill)
		if err != nil {
			methodLogger.Warn().Err(err).Msg("ошибка отправки отчета")
			warnings = append(warnings, fmt.Sprintf("ошибка отправки отчета: %s", err))

			continue
		}

		s.logger.Trace().Msgf("Статистика по '%s' отправлена в BQ", user.Name)
	}

	if len(warnings) >= len(users) {
		methodLogger.Warn().Msg("Ни один клиент не собран")

		return &pb.PushPlacementStatByDayToBQResponse{
			IsOk:     false,
			Warnings: warnings,
		}, errors.New("ни один клиент не собран : %s")
	}

	return &pb.PushPlacementStatByDayToBQResponse{
		IsOk:     true,
		Warnings: warnings,
	}, nil
}
