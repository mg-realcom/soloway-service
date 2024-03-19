package report

import (
	"github.com/go-kit/kit/endpoint"
	"soloway/internal/service/report"
)

type Endpoints struct {
	PushPlacementStatByDayToBQ endpoint.Endpoint
}

func MakeEndpoints(s report.IService) Endpoints {
	return Endpoints{
		PushPlacementStatByDayToBQ: makePushPlacementStatByDayToBQ(s),
	}
}
