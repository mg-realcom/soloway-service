package report

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"soloway/internal/endpoint/report"
)

type RPCServer struct {
	sendReportToStorage kitgrpc.Handler

	pb.UnimplementedSolowayServiceServer
}

// NewServer is a constructor for creating a new instance of a gRPC server(RPCServer structure).
func NewServer(endpoints report.Endpoints, serverOptions []kitgrpc.ServerOption) pb.SolowayServiceServer {
	return &RPCServer{
		sendReportToStorage: kitgrpc.NewServer(endpoints.SendReportToStorage, DecodeRequest, EncodeResponse, serverOptions...),

		UnimplementedSolowayServiceServer: pb.UnimplementedSolowayServiceServer{},
	}
}

func DecodeRequest(_ context.Context, request interface{}) (interface{}, error) {
	return request, nil
}

func EncodeResponse(_ context.Context, response interface{}) (interface{}, error) {
	return response, nil
}
