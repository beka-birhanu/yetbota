package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
)

type GrpcServer interface {
	RegisterService(srv grpc.ServiceRegistrar)
}

type Config struct {
	Port                int           `validate:"required"`
	KeepaliveTime       time.Duration `validate:"required"`
	KeepaliveTimeout    time.Duration `validate:"required"`
	KeepaliveMinTime    time.Duration `validate:"required"`
	PermitWithoutStream bool
	SessionManager      domainAuth.SessionManager `validate:"required"`
	GrpcServers         []GrpcServer              `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func RunServer(ctx context.Context, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.KeepaliveTime,
			Timeout: cfg.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             cfg.KeepaliveMinTime,
			PermitWithoutStream: cfg.PermitWithoutStream,
		}),
		grpc.ChainUnaryInterceptor(makeLoggingInterceptor(), makeUnaryServerInterceptor(cfg.SessionManager)),
		grpc.StreamInterceptor(makeStreamServerInterceptor(cfg.SessionManager)),
	)

	for _, server := range cfg.GrpcServers {
		server.RegisterService(srv)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("gRPC server listening on %s\n", addr)
		errCh <- srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		srv.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
