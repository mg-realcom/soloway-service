package main

import (
	"net"
	"time"

	"soloway/pkg/interceptors"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	googlegrpc "google.golang.org/grpc"
	"soloway/internal/config"
	"soloway/internal/endpoint"
	reportEP "soloway/internal/endpoint/report"
	"soloway/internal/service/report"
	tpGRPC "soloway/internal/transport/grpc"
	tpReport "soloway/internal/transport/grpc/report"
	"soloway/pkg/metrics"
)

func initKitGRPC(appConfig *config.Configuration, endpoints endpoint.ServicesEndpoints, netLogger zerolog.Logger, listenErr chan error) (*googlegrpc.Server, net.Listener) {
	var serverOptions []kitgrpc.ServerOption

	grpcServer := googlegrpc.NewServer(googlegrpc.StatsHandler(otelgrpc.NewServerHandler()), googlegrpc.UnaryInterceptor(interceptors.AuthInterceptor(appConfig.Token)))

	newServer := tpReport.NewServer(endpoints.ReportEP, serverOptions)

	pb.RegisterSolowayServiceServer(grpcServer, newServer)

	l, err := net.Listen(appConfig.GRPC.Network, appConfig.GRPC.Address)
	if err != nil {
		netLogger.Fatal().Err(err).Msg("failed to init net.Listen for grpc")
	} else {
		netLogger.Info().Msg("successful net.Listen for grpc init")
	}

	go tpGRPC.RunGRPCServer(grpcServer, l, netLogger, listenErr)
	time.Sleep(10 * time.Millisecond)

	return grpcServer, l
}

func initEndpoints(appConfig *config.Configuration, apiLogger zerolog.Logger) endpoint.ServicesEndpoints {
	mailService := report.NewService(&apiLogger, appConfig)

	return endpoint.ServicesEndpoints{
		ReportEP: reportEP.MakeEndpoints(mailService),
	}
}

func initPrometheus(prometheusAddr string, coreLogger zerolog.Logger) {
	if prometheusAddr != "" {
		coreLogger.Info().Msg("Сервис Prometheus запущен")
		err := metrics.Listen(prometheusAddr)
		if err != nil {
			coreLogger.Fatal().Err(err).Msg("Ошибка в сервисе: Prometheus")
		}
	} else {
		coreLogger.Warn().Msg("Сервис Prometheus не запущен")
	}
}
