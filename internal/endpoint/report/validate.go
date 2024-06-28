package report

import (
	"errors"
	"fmt"

	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"

	errmsg "soloway/pkg/errors"
)

type ValidateError struct {
	msg   string
	field string
}

func (err *ValidateError) Error() string {
	return fmt.Sprintf("%s: %s", err.field, err.msg)
}

func validatePushPlacementStatByDayToBQ(request interface{}) (*pb.PushPlacementStatByDayToBQRequest, error) {
	req, ok := request.(*pb.PushPlacementStatByDayToBQRequest)
	if !ok {
		err := errors.New(errmsg.ErrMsgFailedCastRequest)

		return nil, err
	}

	if err := validateBqConfig(req.GetBqConfig()); err != nil {
		return nil, err
	}

	if err := validateCsConfig(req.GetCsConfig()); err != nil {
		return nil, err
	}

	if err := validateGsConfig(req.GetGsConfig()); err != nil {
		return nil, err
	}

	if err := validatePeriod(req.GetPeriod()); err != nil {
		return nil, err
	}

	return req, nil
}

// validateBqConfig проверяет поля BqConfig на nil и пустые значения.
func validateBqConfig(config *pb.BqConfig) error {
	if config == nil {
		return &ValidateError{"is nil", "bq_config"}
	}

	if config.ProjectId == "" {
		return &ValidateError{"is empty", "bq_config.project_id"}
	}

	if config.DatasetId == "" {
		return &ValidateError{"is empty", "bq_config.dataset_id"}
	}

	if config.TableId == "" {
		return &ValidateError{"is empty", "bq_config.table_id"}
	}

	return nil
}

// validateCsConfig проверяет поля CsConfig на nil и пустые значения.
func validateCsConfig(config *pb.CsConfig) error {
	if config == nil {
		return &ValidateError{"is nil", "cs_config"}
	}

	if config.BucketName == "" {
		return &ValidateError{"is empty", "cs_config.bucket_name"}
	}

	return nil
}

func validateGsConfig(config *pb.GsConfig) error {
	if config == nil {
		return &ValidateError{"is nil", "gs_config"}
	}

	if config.SpreadsheetId == "" {
		return &ValidateError{"is empty", "gs_config.spreadsheet_id"}
	}

	if config.ServiceKey == "" {
		return &ValidateError{"is empty", "gs_config.service_key"}
	}

	return nil
}

// validatePeriod проверяет поля Period на nil и пустые значения.
func validatePeriod(period *pb.Period) error {
	if period == nil {
		return &ValidateError{"is nil", "period"}
	}

	if period.GetDateFrom() == "" {
		return &ValidateError{"is empty", "period.date_from"}
	}

	if period.GetDateTill() == "" {
		return &ValidateError{"is empty", "period.date_till"}
	}

	return nil
}
