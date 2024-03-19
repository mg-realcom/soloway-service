package helpers

import (
	"context"
	"strings"

	guuid "github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const RequestIDKey = "X-Request-ID"

func GetRequestID(ctx context.Context) (reqID string, newCtx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	if val, ok := md[strings.ToLower(RequestIDKey)]; ok {
		if val[0] != "" {
			reqID = val[0]
		}
	}

	if reqID == "" {
		reqID = guuid.New().String()
	}

	newMd := metadata.New(map[string]string{RequestIDKey: reqID})
	newCtx = metadata.NewOutgoingContext(ctx, newMd)

	return reqID, newCtx
}
