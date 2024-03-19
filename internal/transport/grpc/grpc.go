package grpc

import (
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func RunGRPCServer(grpcSrv *grpc.Server, l net.Listener, log zerolog.Logger, listenErr chan error) {
	reflection.Register(grpcSrv)
	log.Info().Msgf("starting grpc server on %s", l.Addr())

	if err := grpcSrv.Serve(l); err != nil {
		listenErr <- err
	}
}
