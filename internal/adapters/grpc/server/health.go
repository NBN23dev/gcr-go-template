package server

import (
	"context"

	health "google.golang.org/grpc/health/grpc_health_v1"
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

	health.UnimplementedHealthServer
}

func NewHealhCheck() *HealhCheck {
	return &HealhCheck{Status: HealthCheckStatus_SERVING}
}

// If the requested service is unknown, the call will fail with status
// NOT_FOUND.
func (hc *HealhCheck) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_ServingStatus(hc.Status)}, nil
}
