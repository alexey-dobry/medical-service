package app

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
)

func (a *app) shutdown(ctx context.Context) error {
	a.logger.Info("stopping all services")

	restDone := make(chan error, 1)
	go func() { restDone <- a.RESTServer.Shutdown(ctx) }()

	grpcErr := stopGRPC(ctx, a.GRPCServer)

	var errs []error
	if err := <-restDone; err != nil {
		errs = append(errs, fmt.Errorf("REST shutdown: %w", err))
	}
	if grpcErr != nil {
		errs = append(errs, fmt.Errorf("gRPC shutdown: %w", grpcErr))
	}

	a.postgres.Close()
	a.redis.Close()

	if err := errors.Join(errs...); err != nil {
		a.logger.Warnf("app stopped with errors: %s", err)
		return err
	}

	a.logger.Info("app was gracefully shutdown")
	return nil
}

func stopGRPC(ctx context.Context, srv *grpc.Server) error {
	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		go srv.Stop()
		return ctx.Err()
	}
}
