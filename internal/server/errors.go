package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NBN23dev/gcr-go-template/internal/plugins/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// unaryErrorHandler is a custom error handler for unary requests
func unaryErrorHandler(ctx context.Context, sm *runtime.ServeMux, ma runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {
	sts := status.Convert(err)
	code := runtime.HTTPStatusFromCode(sts.Code())

	logger.Error(err.Error(), logger.Payload{
		"code":    fmt.Sprint(code),
		"message": sts.Message(),
	})

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)

	json.NewEncoder(rw).Encode(map[string]any{
		"code":    code,
		"message": sts.Message(),
	})
}

// streamErrorHandler is a custom error handler for stream requests
func streamErrorHandler(ctx context.Context, err error) *status.Status {
	sts := status.Convert(err)
	code := runtime.HTTPStatusFromCode(sts.Code())

	logger.Error(err.Error(), logger.Payload{
		"code":    fmt.Sprint(code),
		"message": sts.Message(),
	})

	return sts
}
