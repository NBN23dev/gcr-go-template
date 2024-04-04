package interceptors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NBN23dev/lib-monitoring/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MonitorUnary is a gRPC interceptor that logs the request and response.
func MonitorUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Call the handler and get the response
	res, err := handler(ctx, req)
	// Log the error if there is one
	if err != nil {
		// Add the attributes to the log
		attrs := map[string]string{
			"attr.name": info.FullMethod,
		}

		// Get the request body
		if body, err := json.Marshal(req); err == nil {
			attrs["attr.req"] = string(body)
		}

		// Get the metadata from the context
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			attrs["attr.md"] = fmt.Sprintf("%v", md)
		}

		// Transform the error to get the status code
		status, ok := status.FromError(err)

		if !ok {
			logger.Error(ctx, err.Error(), attrs)

			return res, err
		}

		switch status.Code() {
		case codes.InvalidArgument, codes.NotFound, codes.PermissionDenied:
			logger.Warn(ctx, err.Error(), attrs)
		default:
			logger.Error(ctx, err.Error(), attrs)
		}
	}

	return res, err
}

// MonitorStream is a gRPC interceptor that logs the request and response.
func MonitorStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	// Call the handler and get the response
	err := handler(srv, ss)

	// Log the error if there is one
	if err != nil {
		// Check if the context was canceled
		if err == context.Canceled {
			return nil
		}

		// Add the attributes to the log
		attrs := map[string]string{
			"attr.name": info.FullMethod,
		}

		// Get the metadata from the context
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			attrs["attr.md"] = fmt.Sprintf("%v", md)
		}

		// Transform the error to get the status code
		status, ok := status.FromError(err)

		if !ok {
			logger.Error(ctx, err.Error(), attrs)

			return err
		}

		switch status.Code() {
		case codes.InvalidArgument, codes.NotFound, codes.PermissionDenied:
			logger.Warn(ctx, err.Error(), attrs)
		default:
			logger.Error(ctx, err.Error(), attrs)
		}
	}

	return err
}
