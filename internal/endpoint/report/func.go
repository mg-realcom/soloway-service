package report

import (
	"context"
	"fmt"

	"github.com/mg-realcom/s3utils"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	repository "soloway/internal/repository/report"
	"soloway/internal/service/report"
	errmsg "soloway/pkg/errors"
	"soloway/pkg/helpers"
	"soloway/pkg/tracing"
)

func makeSendReportToStorage(s report.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := helpers.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str(helpers.RequestIDKey, reqID).Str("Source", "makeSendReportToStorage").Logger()

		span := tracing.SpanFromContext(ctx)
		tracing.SetSpanAttribute(span, tracing.AttributeRequestID, reqID)

		req, err := validateSendReportToStorage(request)
		if err != nil {
			serviceLogger.Err(err).Stack().Msg(errmsg.ErrMsgFailedValidateRequest)

			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s: %s", errmsg.ErrMsgFailedValidateRequest, err.Error()))
		}

		tracing.SetSpanAttributesByStructFields(span, req, tracing.AttributeRequest)
		serviceLogger = serviceLogger.With().Str("project_id", req.Storage.GetYandexStorage().BucketName).Logger()

		storage, err := s3utils.NewClient(ctx, "ru-central1")
		if err != nil {
			msg := errmsg.ErrMsgFailedInitS3
			serviceLogger.Error().Stack().Err(err).Msg(msg)
			return nil, err
		}

		cfg := s.GetConfig()

		gServiceKey := cfg.KeysDir + "/" + req.GsConfig.ServiceKey

		ghSrv, err := sheets.NewService(ctx, option.WithCredentialsFile(gServiceKey))
		if err != nil {
			msg := errmsg.ErrMsgFailedInitGoogleSheets
			serviceLogger.Error().Stack().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}

		repoLogger := s.GetLogger().With().Str("repo", "report").Logger()
		repo := repository.NewRepository(&repoLogger, storage, ghSrv)

		resp, err := s.SendReportToStorage(ctx, req, repo)
		if err != nil {
			return nil, err
		}

		tracing.SetSpanAttributesByStructFields(span, resp, tracing.AttributeResponse)

		return resp, nil
	}
}
