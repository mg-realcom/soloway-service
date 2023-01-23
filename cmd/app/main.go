package main

import (
	"context"
	"flag"
	"github.com/go-co-op/gocron"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
	"github.com/rs/zerolog"
	"log"
	"os"
	"runtime/debug"
	"solowayStat/internal/app"
	"solowayStat/internal/config"
	"time"
)

func main() {
	var fileConfig = flag.String("f", "config.yml", "configuration file")
	flag.Parse()
	f := *fileConfig
	buildInfo, _ := debug.ReadBuildInfo()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Int("pid", os.Getpid()).
		Str("go_version", buildInfo.GoVersion).
		Logger()

	cfg, err := config.NewConfig(f)
	if err != nil {
		logger.Fatal().Err(err).Msg("Ошибка в файле настроек")
	}

	telegramService, err := telegram.New(cfg.Token)
	if err != nil {
		logger.Fatal().Err(err).Msg("Ошибка в сервисе: Telegram")
	}
	telegramService.AddReceivers(cfg.Chat)
	notify.UseServices(telegramService)

	ctx := context.Background()
	err = notify.Send(ctx, cfg.Name, "Скрипт запущен")
	if err != nil {
		logger.Warn().Err(err).Msg("ошибка сервера уведомлений")
	}

	app.Run(cfg, &logger)
	if err != nil {
		log.Fatalf("Ошибка в приложении %s", err)
	}

	s := gocron.NewScheduler(time.UTC)
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Fatal("Error loading location")
	}
	s.ChangeLocation(location)
	s.Every(1).Day().At("10:00").Do(app.Run, *cfg)
	s.StartBlocking()
}
