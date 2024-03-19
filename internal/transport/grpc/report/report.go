package report

import (
	"context"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"soloway/pkg/expanded_errors"
)

func (s *RPCServer) PushPlacementStatByDayToBQ(ctx context.Context, req *pb.PushPlacementStatByDayToBQRequest) (*pb.PushPlacementStatByDayToBQResponse, error) {
	return expanded_errors.EncodeResponseGRPC[*pb.PushPlacementStatByDayToBQResponse](s.pushPlacementStatByDayToBQ.ServeGRPC(ctx, req))
}
