package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	adapters "github.com/NBN23dev/gcr-go-template/internal/adapters/grpc"
	"github.com/NBN23dev/gcr-go-template/internal/adapters/grpc/interceptors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	gs *grpc.Server
	hc *HealhCheck
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
func grpcHandlerFunc(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
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
	srv := grpc.NewServer([]grpc.ServerOption{
		grpc.ConnectionTimeout(time.Duration(10) * time.Second),
		grpc.ChainUnaryInterceptor(
			interceptors.MonitorUnary,
			interceptors.ValidationUnary,
			interceptors.HeadersUnary,
		),
		grpc.ChainStreamInterceptor(interceptors.MonitorStream),
	}...)

	// Register rpc's
	// TODO: Register GRPC service - pb.Register${ServiceName}ServiceServer(srv, adapter)

	// Health check
	hc := NewHealhCheck()

	health.RegisterHealthServer(srv, hc)

	// Reflection
	reflection.Register(srv)

	return &Server{gs: srv, hc: hc}, nil
}

// Start the server and listen for incoming requests.
func (srv *Server) Start(port int) error {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	}

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", port), opts...)

	if err != nil {
		return err
	}

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(unaryErrorHandler),
		runtime.WithStreamErrorHandler(streamErrorHandler),
		runtime.WithHealthEndpointAt(health.NewHealthClient(conn), "/"),
		runtime.WithOutgoingHeaderMatcher(func(header string) (string, bool) {
			if header == "trailer" {
				return "", false
			}

			return header, true
		}),
	)

	// Register rpc's handler
	// TODO: Register GRPC service - err = pb.Register${ServiceName}ServiceHandler(ctx, mux, conn)

	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), grpcHandlerFunc(srv.gs, mux))
}

// Stop shutdown the server gracefully.
func (srv *Server) Stop(cb func(os.Signal)) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	sig := <-done

	<-time.After(30 * time.Second)

	// Shutdown
	srv.hc.Status = HealthCheckStatus_NOT_SERVING

	srv.gs.GracefulStop()

	// Callback handler
	cb(sig)
}
