package interceptors

import (
	"context"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ValidationUnary is a gRPC interceptor that validates the request.
func ValidationUnary() (grpc.UnaryServerInterceptor, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		request, _ := (req).(proto.Message)

		if err := v.Validate(request); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return handler(ctx, req)
	}, nil
}
