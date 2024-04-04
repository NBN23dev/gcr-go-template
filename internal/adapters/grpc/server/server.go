package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	adapters "github.com/NBN23dev/gcr-go-template/internal/adapters/grpc"
	"github.com/NBN23dev/gcr-go-template/internal/adapters/grpc/server/interceptors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	gs *grpc.Server
	hc *HealhCheck
}

// grpcHandler returns an http.Handler that delegates to grpcServer on incoming gRPC
func grpcHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)

			return
		}

		if r.Method == "OPTIONS" {
			return
		}

		httpHandler.ServeHTTP(w, r)
	}), &http2.Server{})
}

// NewServer
func NewServer(adapter *adapters.GRPCAdapter) (*Server, error) {
	// Create a new grpc server
	srv := grpc.NewServer([]grpc.ServerOption{
		grpc.ConnectionTimeout(10 * time.Second),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: false,
		}),
		grpc.ChainStreamInterceptor(
			interceptors.MonitorStream,
		),
	}...)

	// Register rpc's
	// TODO: Register GRPC service
	// pb.RegisterPublishServiceServer(srv, adapter)

	// Health check
	hc := NewHealhCheck()

	health.RegisterHealthServer(srv, hc)

	// Reflection
	reflection.Register(srv)

	return &Server{gs: srv, hc: hc}, nil
}

// Start the server and listen for incoming requests.
func (srv *Server) Start(port int) error {
	_, cancel := context.WithCancel(context.Background())

	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	}

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port), opts...)
	if err != nil {
		return err
	}

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(UnaryErrorHandler),
		runtime.WithHealthEndpointAt(health.NewHealthClient(conn), "/"),
		runtime.WithIncomingHeaderMatcher(func(header string) (string, bool) {
			if strings.EqualFold(header, "traceparent") {
				return header, true
			}

			return "", false
		}),
		runtime.WithOutgoingHeaderMatcher(func(header string) (string, bool) {
			if strings.EqualFold(header, "trailer") {
				return "", false
			}

			return header, true
		}),
	)

	// Register rpc's handler
	// TODO: Register GRPC service
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// err = pb.Register${ServiceName}ServiceHandler(ctx, mux, conn)
	// if err != nil {
	// 	return err
	// }

	// GRPC
	handler := grpcHandler(srv.gs, mux)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}

// Stop shutdown the server gracefully.
func (srv *Server) Stop() {
	// Set the server to not available for serving
	srv.hc.Status = HealthCheckStatus_NOT_SERVING

	srv.gs.GracefulStop()
}
