package interceptors

import (
	"context"
	"encoding/json"

	"github.com/NBN23dev/go-service-template/internal/plugins/logger"
	"github.com/NBN23dev/go-service-template/internal/plugins/tracer"
	"google.golang.org/grpc"
)

// MonitorUnary is a gRPC interceptor that logs the request and response.
func MonitorUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	trace := tracer.Start(info.FullMethod)

	res, err := handler(ctx, req)

	defer trace.End(err)

	if err != nil {
		body, _ := json.Marshal(req)

		payload := map[string]string{
			"name": info.FullMethod,
			"req":  string(body),
		}

		trace.SetAttributes(payload)

		logger.Error(err.Error(), payload)
	}

	return res, err
}

func MonitorStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	trace := tracer.Start(info.FullMethod)

	err := handler(srv, ss)

	defer trace.End(err)

	if err != nil {
		logger.Error(err.Error(), nil)
	}

	return err
}
