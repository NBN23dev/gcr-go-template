package server

import (
	"context"

	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthCheckStatus int32

const (
	HealthCheckStatus_UNKNOWN         = 0
	HealthCheckStatus_SERVING         = 1
	HealthCheckStatus_NOT_SERVING     = 2
	HealthCheckStatus_SERVICE_UNKNOWN = 3 // Used only by the Watch method.
)

// HealhCheck
type HealhCheck struct {
	Status HealthCheckStatus
}

func NewHealhCheck() *HealhCheck {
	return &HealhCheck{Status: HealthCheckStatus_SERVING}
}

// If the requested service is unknown, the call will fail with status
// NOT_FOUND.
func (hc *HealhCheck) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_ServingStatus(hc.Status)}, nil
}

// Performs a watch for the serving status of the requested service.
// The server will immediately send back a message indicating the current
// serving status.  It will then subsequently send a new message whenever
// the service's serving status changes.
//
// If the requested service is unknown when the call is received, the
// server will send a message setting the serving status to
// SERVICE_UNKNOWN but will *not* terminate the call.  If at some
// future point, the serving status of the service becomes known, the
// server will send a new message with the service's serving status.
//
// If the call terminates with status UNIMPLEMENTED, then clients
// should assume this method is not supported and should not retry the
// call. If the call terminates with any other status (including OK),
// clients should retry the call with appropriate exponential backoff.
func (hc *HealhCheck) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	if hc.Status == HealthCheckStatus_SERVING {
		return nil
	}

	return status.Error(codes.Unavailable, "unavailable")
}
