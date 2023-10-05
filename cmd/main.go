package main

import (
	"fmt"
	"log"
	"os"

	adapters "github.com/NBN23dev/go-service-template/internal/adapters/grpc"
	"github.com/NBN23dev/go-service-template/internal/core/services"
	"github.com/NBN23dev/go-service-template/internal/plugins/logger"
	"github.com/NBN23dev/go-service-template/internal/plugins/tracer"
	"github.com/NBN23dev/go-service-template/internal/repositories"
	"github.com/NBN23dev/go-service-template/internal/server"
	"github.com/NBN23dev/go-service-template/internal/utils"
)

func main() {
	name, _ := utils.GetEnvOr("SERVICE_NAME", "unknown")

	// Logger
	logLevel, _ := utils.GetEnvOr("LOG_LEVEL", string(logger.LevelInfo))

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
	go server.GracefulShutdown(func(sig os.Signal) {
		tracer.Shutdown()

		logger.Info(fmt.Sprintf("'%s' service it is about to end", name), nil)
	})

	logger.Info(fmt.Sprintf("'%s' service it is about to start", name), nil)

	// Start server
	port, _ := utils.GetEnvOr("PORT", 8080)

	server.Start(port)
}
