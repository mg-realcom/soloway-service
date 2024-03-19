package expanded_errors

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"soloway/pkg/tracing"
)

/*
OutputError struct that realize Errorer interface
errorMessage is field for human-understandable error message
errorDetail is field for detail error message for logging
grpcStatusCode is field for grpc response code
httpStatusCode is field for http response code.
*/
type OutputError struct {
	errorMessage   string
	errorDetail    error
	grpcStatusCode codes.Code
	httpStatusCode int
}

// Error returns OutputError.errorMessage field.
func (e *OutputError) Error() string {
	return e.errorMessage
}

// ErrorDetail returns OutputError.errorDetail field.
func (e *OutputError) ErrorDetail() error {
	return e.errorDetail
}

// New is the constructor for OutputError.
func New(errorMessage string, errorDetail error, grpcCode codes.Code, httpCode int) *OutputError {
	return &OutputError{
		errorMessage:   errorMessage,
		errorDetail:    errorDetail,
		grpcStatusCode: grpcCode,
		httpStatusCode: httpCode,
	}
}

// GetGRPC returns formed grpc status error by OutputError info.
func (e *OutputError) GetGRPC() error {
	return status.Error(e.grpcStatusCode, e.errorMessage)
}

// EncodeResponseGRPC is a function that forms the grpc status error response by OutputError info.
//
// T generic must be pointer type.
func EncodeResponseGRPC[T proto.Message](ctx context.Context, response interface{}, err error) (T, error) {
	var nilT T

	if err != nil {
		span := tracing.SpanFromContext(ctx)
		spanWithEnd := false

		var outputError *OutputError

		if errors.As(err, &outputError) {
			tracing.EndSpanError(span, outputError.errorDetail, outputError.errorMessage, spanWithEnd)

			return nilT, outputError.GetGRPC()
		} else {
			tracing.EndSpanError(span, err, err.Error(), spanWithEnd)

			return nilT, New(err.Error(), err, codes.Unknown, http.StatusInternalServerError).GetGRPC()
		}
	}

	return response.(T), nil
}
