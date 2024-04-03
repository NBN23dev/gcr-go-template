package main

import (
	"fmt"
	"log"
	"os"

	adapters "github.com/NBN23dev/gcr-go-template/internal/adapters/grpc"
	"github.com/NBN23dev/gcr-go-template/internal/core/services"
	"github.com/NBN23dev/gcr-go-template/internal/helpers"
	"github.com/NBN23dev/gcr-go-template/internal/plugins/logger"
	"github.com/NBN23dev/gcr-go-template/internal/plugins/tracer"
	"github.com/NBN23dev/gcr-go-template/internal/repositories"
	"github.com/NBN23dev/gcr-go-template/internal/server"
)

func main() {
	name, _ := helpers.GetEnvOr("SERVICE_NAME", "unknown")

	// Logger
	logLevel, _ := helpers.GetEnvOr("LOG_LEVEL", string(logger.LevelInfo))

	if err := logger.Init(name, logger.Level(logLevel)); err != nil {
		log.Fatal(err)
	}

	// Tracer
	if err := tracer.Init(name); err != nil {
		logger.Fatal(err)
	}

	// Service
	repos, err := repositories.NewRepository()

	if err != nil {
		logger.Fatal(err)
	}

	service, err := services.NewService(repos)

	if err != nil {
		logger.Fatal(err)
	}

	adapter := adapters.NewGRPCAdapter(service)

	// Create server
	server, err := server.NewServer(adapter)

	if err != nil {
		logger.Fatal(err)
	}

	// Shutdown
	go server.Stop(func(sig os.Signal) {
		tracer.Shutdown()

		logger.Info(fmt.Sprintf("'%s' service it is about to end", name), nil)
	})

	logger.Info(fmt.Sprintf("'%s' service it is about to start", name), nil)

	// Start server
	port, _ := helpers.GetEnvOr("PORT", 8080)

	server.Start(port)
}
