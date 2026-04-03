package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beka-birhanu/yetbota/ai-service/drivers/config"
	"github.com/beka-birhanu/yetbota/ai-service/drivers/logger"
	"github.com/beka-birhanu/yetbota/ai-service/drivers/validator"

	cmdGrpc "github.com/beka-birhanu/yetbota/ai-service/cmd/grpc"
	cmdRest "github.com/beka-birhanu/yetbota/ai-service/cmd/rest"
	"github.com/beka-birhanu/yetbota/ai-service/internal/services/endpoint"
	aiGrpc "github.com/beka-birhanu/yetbota/ai-service/internal/transport/grpc/ai"
	restTransport "github.com/beka-birhanu/yetbota/ai-service/internal/transport/rest"
)

func main() {
	validator.InitValidator()
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error load config: %v", err))
	}

	ctx, cancel := signalContext()
	defer cancel()

	endpoints := endpoint.NewEndpoints()
	aiHandler := aiGrpc.NewHandler(endpoints)

	restHandler := restTransport.NewHandler(cfg)
	go func() {
		err := cmdRest.RunServer(ctx, &cmdRest.Config{
			Port:         cfg.Rest.Port,
			ReadTimeout:  time.Duration(cfg.Rest.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Rest.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Rest.IdleTimeout) * time.Second,
			Handler:      restHandler,
		})
		if err != nil {
			panic(fmt.Errorf("rest server error: %v", err))
		}
	}()

	if err := cmdGrpc.RunServer(ctx, &cmdGrpc.Config{
		Port:                cfg.Grpc.Port,
		KeepaliveTime:       time.Duration(cfg.Grpc.Keepalive.Time) * time.Second,
		KeepaliveTimeout:    time.Duration(cfg.Grpc.Keepalive.Timeout) * time.Second,
		KeepaliveMinTime:    time.Duration(cfg.Grpc.Keepalive.MinTime) * time.Second,
		PermitWithoutStream: cfg.Grpc.Keepalive.PermitWithoutStream,
		GrpcServers:         []cmdGrpc.GrpcServer{aiHandler},
	}); err != nil {
		panic(fmt.Errorf("gRPC server error: %v", err))
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ch
		cancel()
	}()

	return ctx, cancel
}