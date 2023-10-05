package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	adapters "github.com/NBN23dev/go-service-template/internal/adapters/grpc"
	"github.com/NBN23dev/go-service-template/internal/adapters/grpc/interceptors"
	"github.com/NBN23dev/go-service-template/internal/plugins/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Server struct {
	gs *grpc.Server
	hc *HealhCheck
}

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

func streamErrorHandler(ctx context.Context, err error) *status.Status {
	sts := status.Convert(err)
	code := runtime.HTTPStatusFromCode(sts.Code())

	logger.Error(err.Error(), logger.Payload{
		"code":    fmt.Sprint(code),
		"message": sts.Message(),
	})

	return sts
}

func grpcHandlerFunc(grpcServer *grpc.Server, httHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)

			return
		}

		httHandler.ServeHTTP(w, r)
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

// StartServer
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
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{
				EmitDefaultValues: false,
			}},
		),
	)

	// Register rpc's handler
	// TODO: Register GRPC service - err = pb.Register${ServiceName}ServiceHandler(ctx, mux, conn)

	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), grpcHandlerFunc(srv.gs, mux))
}

// GracefulShutdown
func (srv *Server) GracefulShutdown(cb func(os.Signal)) {
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
