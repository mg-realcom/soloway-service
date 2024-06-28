package report

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	repository "soloway/internal/repository/report"
	"soloway/internal/service/report"
	errmsg "soloway/pkg/errors"
	"soloway/pkg/helpers"
	"soloway/pkg/repository/bq"
	"soloway/pkg/repository/cs"
	"soloway/pkg/tracing"
)

func makePushPlacementStatByDayToBQ(s report.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := helpers.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str(helpers.RequestIDKey, reqID).Str("Source", "makePushReviseMvideoToBQ").Logger()

		span := tracing.SpanFromContext(ctx)
		tracing.SetSpanAttribute(span, tracing.AttributeRequestID, reqID)

		req, err := validatePushPlacementStatByDayToBQ(request)
		if err != nil {
			serviceLogger.Err(err).Stack().Msg(errmsg.ErrMsgFailedValidateRequest)

			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s: %s", errmsg.ErrMsgFailedValidateRequest, err.Error()))
		}

		tracing.SetSpanAttribute(span, tracing.AttributeProjectID, req.BqConfig.ProjectId)
		tracing.SetSpanAttributesByStructFields(span, req, tracing.AttributeRequest)
		serviceLogger = serviceLogger.With().Str("project_id", req.BqConfig.ProjectId).Logger()

		cfg := s.GetConfig()

		gServiceKey := cfg.KeysDir + "/" + req.BqConfig.ServiceKey

		bd, err := bq.NewClient(ctx, s.GetLogger(), req.BqConfig.ProjectId, gServiceKey)
		if err != nil {
			msg := errmsg.ErrMsgFailedInitBigQuery
			serviceLogger.Error().Stack().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}

		defer func(bqClient *bq.Client) {
			err := bqClient.Close()
			if err != nil {
				serviceLogger.Error().Stack().Err(err).Msg(errmsg.ErrMsgFailedCloseBigQuery)
			}
		}(bd)

		storage, err := cs.NewClient(ctx, s.GetLogger(), req.CsConfig.BucketName, gServiceKey)
		if err != nil {
			msg := errmsg.ErrMsgFailedInitCloudStorage
			serviceLogger.Error().Stack().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}

		defer func(csClient *cs.Client) {
			err := csClient.Close()
			if err != nil {
				serviceLogger.Error().Stack().Err(err).Msg(errmsg.ErrMsgFailedCloseCloudStorage)
			}
		}(storage)

		gsServiceKey := cfg.KeysDir + "/" + req.GsConfig.ServiceKey

		ghSrv, err := sheets.NewService(ctx, option.WithCredentialsFile(gsServiceKey))
		if err != nil {
			msg := errmsg.ErrMsgFailedInitGoogleSheets
			serviceLogger.Error().Stack().Err(err).Msg(msg)

			return nil, status.Error(codes.InvalidArgument, msg)
		}

		repoLogger := s.GetLogger().With().Str("repo", "report").Logger()
		repo := repository.NewRepository(&repoLogger, bd, storage, ghSrv)

		resp, err := s.PushPlacementStatByDayToBQ(ctx, req, repo)
		if err != nil {
			return nil, err
		}

		tracing.SetSpanAttributesByStructFields(span, resp, tracing.AttributeResponse)

		return resp, nil
	}
}
