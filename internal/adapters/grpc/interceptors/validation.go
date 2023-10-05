package interceptors

import (
	"context"
	"fmt"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ValidationUnary is a gRPC interceptor that validates the request.
func ValidationUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	v, err := protovalidate.New()

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request, ok := (req).(proto.Message)

	if !ok {
		panic(fmt.Sprintf("request %s does not implement validation", info.FullMethod))
	}

	if err := v.Validate(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return handler(ctx, req)
}
