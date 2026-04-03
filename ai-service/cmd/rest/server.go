package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/ai-service/drivers/validator"
)

type Config struct {
	Port         int           `validate:"required"`
	ReadTimeout  time.Duration `validate:"required"`
	WriteTimeout time.Duration `validate:"required"`
	IdleTimeout  time.Duration `validate:"required"`
	Handler      http.Handler  `validate:"required"`
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

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      cfg.Handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("REST server listening on %s\n", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}