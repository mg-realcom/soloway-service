package report

import (
	"errors"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"

	errmsg "soloway/pkg/errors"
)

func validatePushPlacementStatByDayToBQ(request interface{}) (*pb.PushPlacementStatByDayToBQRequest, error) {
	req, ok := request.(*pb.PushPlacementStatByDayToBQRequest)
	if !ok {
		err := errors.New(errmsg.ErrMsgFailedCastRequest)

		return nil, err
	}

	if req.GetBqConfig() == nil {
		return nil, errors.New("bq_config is nil")
	}

	if req.BqConfig.ProjectId == "" {
		return nil, errors.New("bq_config.project_id is empty")
	}

	if req.BqConfig.DatasetId == "" {
		return nil, errors.New("bq_config.dataset_id is empty")
	}

	if req.BqConfig.TableId == "" {
		return nil, errors.New("bq_config.table_id is empty")
	}

	if req.GetCsConfig() == nil {
		return nil, errors.New("cs_config is nil")
	}

	if req.CsConfig.BucketName == "" {
		return nil, errors.New("cs_config.bucket_name is empty")
	}

	if req.GetGsConfig() == nil {
		return nil, errors.New("gs_config is nil")
	}

	if req.GsConfig.SpreadsheetId == "" {
		return nil, errors.New("gs_config.spreadsheet_id is empty")
	}

	if req.GsConfig.ServiceKey == "" {
		return nil, errors.New("gs_config.service_key is empty")
	}

	if req.GetPeriod() == nil {
		return nil, errors.New("period is nil")
	}

	if req.GetPeriod().GetDateFrom() == "" {
		return nil, errors.New("period.date_from is empty")
	}

	if req.GetPeriod().GetDateTill() == "" {
		return nil, errors.New("period.date_till is empty")
	}

	return req, nil
}
