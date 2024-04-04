package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// UnaryErrorHandler is a custom error handler for unary requests
func UnaryErrorHandler(ctx context.Context, sm *runtime.ServeMux, ma runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {
	sts := status.Convert(err)
	code := runtime.HTTPStatusFromCode(sts.Code())

	// Set response headers
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)

	// Return error message as JSON response
	json.NewEncoder(rw).Encode(map[string]any{
		"code":    code,
		"message": sts.Message(),
	})
}
