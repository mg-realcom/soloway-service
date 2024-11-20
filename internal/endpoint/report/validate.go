package report

import (
	"errors"
	"fmt"
	"strings"

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

func validateSendReportToStorage(request interface{}) (*pb.SendReportToStorageRequest, error) {
	req, ok := request.(*pb.SendReportToStorageRequest)
	if !ok {
		err := errors.New(errmsg.ErrMsgFailedCastRequest)

		return nil, err
	}

	if err := validateStorageConfig(req.GetStorage()); err != nil {
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

// validateStorageConfig проверяет поля Storage на nil empty и некорректные значения.
func validateStorageConfig(config *pb.Storage) error {
	if config == nil {
		return &ValidateError{msg: "is nil", field: "storage"}
	}

	if config.GetYandexStorage() == nil {
		return &ValidateError{msg: "is nil", field: "storage.yandex_storage"}
	}

	if config.GetYandexStorage().GetBucketName() == "" {
		return &ValidateError{msg: "is empty", field: "storage.yandex_storage.bucket_name"}
	}

	if config.GetYandexStorage().GetFolderName() == "" {
		return &ValidateError{msg: "is empty", field: "storage.yandex_storage.folder_name"}
	}

	if strings.HasPrefix(config.GetYandexStorage().GetFolderName(), "/") {
		return &ValidateError{msg: "is invalid, start with /", field: "storage.yandex_storage.folder_name"}
	}

	if strings.HasSuffix(config.GetYandexStorage().GetFolderName(), "/") {
		return &ValidateError{msg: "is invalid, end with /", field: "storage.yandex_storage.folder_name"}
	}

	return nil
}
