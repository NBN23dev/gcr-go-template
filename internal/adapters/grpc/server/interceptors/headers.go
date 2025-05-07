package interceptors

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// eTag returns the ETag header value for a given byte slice.
func eTag(value []byte) string {
	hash := fmt.Sprintf("%x", sha1.Sum(value))

	return fmt.Sprintf("W/\"%d-%s\"", len(value), hash)
}

// HeadersUnary is a gRPC interceptor that adds useful headers to response.
func HeadersUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	res, err := handler(ctx, req)

	if err == nil {
		bytes, _ := json.Marshal(res)

		headers := metadata.New(map[string]string{
			"Cache-Control": "max-age=3600",
			"ETag":          eTag(bytes),
		})

		grpc.SendHeader(ctx, headers)
	}

	return res, err
}
