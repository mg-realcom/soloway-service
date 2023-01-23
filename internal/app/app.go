package app

import (
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"github.com/nikoksr/notify"
	"github.com/rs/zerolog"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/api/option"
	"solowayStat/internal/config"
	"solowayStat/internal/repository"
	"solowayStat/internal/services"
	"solowayStat/pkg/solowaysdk"
	"time"
)

func Run(cfg *config.Config, logger *zerolog.Logger) {
	err := GetPlacementStatByDay(cfg, logger)
	if err != nil {
		logger.Warn().Err(err).Msg("ошибка при выполнении скрипта")
		err = notify.Send(context.Background(), cfg.Name, fmt.Sprintf("Ошибка: %s", err))
		if err != nil {
			logger.Warn().Err(err).Msg("ошибка сервиса уведомлений")
		}
	}
}

func GetPlacementStatByDay(cfg *config.Config, logger *zerolog.Logger) (err error) {
	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, cfg.BQ.ProjectID, option.WithCredentialsFile(cfg.BQ.ServiceKeyPath))
	if err != nil {
		return fmt.Errorf("ошибка инициализации bq: %w", err)
	}

	repo := repository.NewRepository(*bqClient)
	client := solowaysdk.NewClient(cfg.Soloway.UserName, cfg.Soloway.Password)
	srv := services.NewService(*client, *repo)

	err = srv.Info.Login()
	if err != nil {
		return fmt.Errorf("ошибка аутентификации: %w", err)
	}

	logger.Info().Msg("Информация о клиенте")
	err = srv.Info.Whoami()
	if err != nil {
		return fmt.Errorf("ошибка получения инфо о клиенте: %w", err)
	}

	placements, err := srv.Info.GetPlacements()
	if err != nil {
		return fmt.Errorf("ошибка получения площадок: %w", err)
	}

	err = srv.Stat.Login()
	if err != nil {
		return fmt.Errorf("ошибка аутентификации: %w", err)
	}
	logger.Info().Msg("Статистика площадок по дням")
	resultData := make([]solowaysdk.PlacementsStatByDay, 0, len(placements.List))
	timeFinish := time.Now().AddDate(0, 0, -1)
	timeStart := timeFinish.AddDate(-cfg.App.DeltaYear, -cfg.App.DeltaMonth, -cfg.App.DeltaDay)
	bar := progressbar.Default(int64(len(placements.List)), "Получение данных по площадкам")
	for _, place := range placements.List {
		err := bar.Add(1)
		if err != nil {
			logger.Warn().Err(err).Msg("ошибка прогресс бара")
		}
		stat, err := srv.Stat.GetPlacementStatByDay(place.Doc.Guid, timeStart, timeFinish)
		if err != nil {
			return fmt.Errorf("ошибка получения площадок по дням: %w", err)
		}
		resultData = append(resultData, stat)
	}

	logger.Info().Msg("Отправка в BQ: Статистика площадок по дням")
	err = srv.Stat.SendAnyPlacementStatByDay(cfg.BQ.DatasetID, cfg.BQ.TableID, resultData, placements)
	if err != nil {
		return fmt.Errorf("ошибка отправки данных в bq: %w", err)
	}
	return nil

}
