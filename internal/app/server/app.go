package server

import (
	"Soloway/internal/config"
	"Soloway/internal/controllers/grpc/v1"
	"Soloway/pb"
	"context"
	"fmt"
	"github.com/nikoksr/notify"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type App struct {
	cfg           *config.ServerConfig
	grpcServer    *grpc.Server
	solowayServer pb.SolowayServiceServer
	logger        *zerolog.Logger
	Notify        notify.Notifier
}

func NewApp(cfg *config.ServerConfig, logger *zerolog.Logger, notify notify.Notifier) *App {
	grpcServer := v1.NewServer(*cfg, logger, pb.UnimplementedSolowayServiceServer{})

	return &App{
		cfg:           cfg,
		grpcServer:    nil,
		solowayServer: grpcServer,
		logger:        logger,
		Notify:        notify,
	}
}

func (a App) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return a.StartGRPC(a.solowayServer)
	})

	return grp.Wait()
}

func (a App) StartGRPC(server pb.SolowayServiceServer) error {
	addr := net.JoinHostPort(a.cfg.GRPC.IP, strconv.Itoa(a.cfg.GRPC.Port))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("failed to create listener")
	}

	a.grpcServer = grpc.NewServer()
	pb.RegisterSolowayServiceServer(a.grpcServer, server)

	a.logger.Info().Msg(fmt.Sprintf("GRPC запущен на %s:%d", a.cfg.GRPC.IP, a.cfg.GRPC.Port))

	err = a.Notify.Send(context.Background(), "Soloway Service", fmt.Sprintf("gRPC запущен на %v:%v", a.cfg.GRPC.IP, a.cfg.GRPC.Port))
	if err != nil {
		a.logger.Fatal().Err(err).Msg("ошибка отправки уведомления")
	}

	if err := a.grpcServer.Serve(lis); err != nil {
		a.logger.Fatal().Err(err).Msg("failed to serve")
	}

	return nil
}
