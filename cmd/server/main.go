package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"soloway/internal/config"
	"soloway/pkg/logger"
	"soloway/pkg/tracing"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("can`t load configuration")
	}

	err = cfg.Validation()
	if err != nil {
		log.Fatalf("not valid configuration: %s", err)
	}

	var baseLogger zerolog.Logger

	var loggerCloser io.WriteCloser

	baseLogger, loggerCloser, err = logger.NewLogger(os.Stdout, cfg.Log.Level)

	apiLogger := logger.NewComponentLogger(baseLogger, "api", 2)
	coreLogger := logger.NewComponentLogger(baseLogger, "core", 2)
	netLogger := logger.NewComponentLogger(baseLogger, "net", 2)

	defer func() {
		if loggerCloser != nil {
			err = loggerCloser.Close()
			if err != nil {
				log.Fatalf("error acquired while closing log writer: %+v", err)
			}
		}
	}()

	defer func() {
		coreLogger.Info().Msg("application stopped")
	}()

	coreLogger.Info().Msg("application started")

	go initPrometheus(cfg.PrometheusAddr, coreLogger)

	err = tracing.Init(ctx, cfg.Telemetry.ServerName, cfg.Telemetry.JaegerEndpoint)
	if err != nil {
		coreLogger.Fatal().Err(err).Msg("failed to create tracer")
	}

	serviceEndpoints := initEndpoints(cfg, apiLogger)

	listenErr := make(chan error, 1)

	grpcServer, grpcListener := initKitGRPC(cfg, serviceEndpoints, netLogger, listenErr)
	defer func() {
		err = grpcListener.Close()
		if err != nil {
			netLogger.Warn().Err(err).Msgf("failed to close net.Listen - %+v", err)
		}
	}()

	runApp(grpcServer, coreLogger, listenErr)
}

func runApp(grpcServer *grpc.Server, coreLogger zerolog.Logger, listenErr chan error) {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var err error
	runningApp := true

	for runningApp {
		select {
		// handle error channel
		case err = <-listenErr:
			if err != nil {
				coreLogger.Error().Err(err).Msg("received listener error")
				shutdownCh <- os.Kill
			}
		// handle os system signal
		case sig := <-shutdownCh:
			coreLogger.Info().Msgf("shutdown signal received: %s", sig.String())

			if err != nil {
				coreLogger.Error().Err(err).Msg("received http Shutdown error")
			}

			grpcServer.GracefulStop()
			coreLogger.Info().Msg("server stopped")

			runningApp = false

			break
		}
	}
}
