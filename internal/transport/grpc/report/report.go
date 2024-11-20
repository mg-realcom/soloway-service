package report

import (
	"context"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"soloway/pkg/expanded_errors"
)

func (s *RPCServer) SendReportToStorage(ctx context.Context, req *pb.SendReportToStorageRequest) (*pb.SendReportToStorageResponse, error) {
	return expanded_errors.EncodeResponseGRPC[*pb.SendReportToStorageResponse](s.sendReportToStorage.ServeGRPC(ctx, req))
}
