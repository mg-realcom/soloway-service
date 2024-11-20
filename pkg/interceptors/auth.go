package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	errMissingMetadata    = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken       = status.Errorf(codes.Unauthenticated, "invalid token")
	errInvalidTokenFormat = status.Errorf(codes.Unauthenticated, "invalid token format")
	errTokenIsRequired    = status.Errorf(codes.Unauthenticated, "token is required")
)

func AuthInterceptor(validToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, errTokenIsRequired
		}

		token := getTokenFromAuthHeader(authHeader[0])
		if token == "" {
			return nil, errInvalidTokenFormat
		}

		if token != validToken {
			return nil, errInvalidToken
		}

		return handler(ctx, req)
	}
}

func getTokenFromAuthHeader(authHeader string) string {
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
