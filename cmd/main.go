package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	adapters "github.com/NBN23dev/gcr-go-template/internal/adapters/grpc"
	"github.com/NBN23dev/gcr-go-template/internal/adapters/grpc/server"
	"github.com/NBN23dev/gcr-go-template/internal/core/services"
	"github.com/NBN23dev/gcr-go-template/internal/helpers"
	"github.com/NBN23dev/gcr-go-template/internal/repositories"
	"github.com/NBN23dev/lib-monitoring/logger"
	"github.com/NBN23dev/lib-monitoring/tracer"
)

// gracefulShutdown is a function that allows to shutdown the instance gracefully
func gracefulShutdown(callback func(os.Signal)) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	sig := <-done

	<-time.After(30 * time.Second)

	callback(sig)
}

func main() {
	// Context
	ctx := context.Background()

	// Service name
	name, _ := helpers.GetEnvOr("SERVICE_NAME", "unknown")

	// Logger
	logLevel, _ := helpers.GetEnvOr("LOG_LEVEL", logger.LevelInfo.String())

	if err := logger.Init(ctx, name, logger.LevelFrom(logLevel)); err != nil {
		log.Fatal(err)
	}

	// Tracer
	if err := tracer.Init(ctx, name); err != nil {
		logger.Fatal(ctx, err)
	}

	// Repositories
	repos, err := repositories.NewRepository()
	if err != nil {
		logger.Fatal(ctx, err)
	}

	// Service
	service, err := services.NewService(repos)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	// Adapter
	adapter := adapters.NewGRPCAdapter(service)

	// Create server
	server, err := server.NewServer(adapter)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	// Shutdown
	go gracefulShutdown(func(sig os.Signal) {
		server.Stop()

		tracer.Shutdown(ctx)

		logger.Info(ctx, fmt.Sprintf("'%s' service it is about to end", name), nil)
	})

	logger.Info(ctx, fmt.Sprintf("'%s' service it is about to start", name), nil)

	// Start server
	port, _ := helpers.GetEnvOr("PORT", 8080)

	err = server.Start(port)
	if err != nil {
		logger.Fatal(ctx, err)
	}
}
