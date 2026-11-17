package app

import (
	"context"
	"errors"
	"fmt"
)

func (a *app) shutdown(ctx context.Context) error {
	a.logger.Info("stopping all services")

	var errs []error

	if err := a.RESTServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("REST shutdown: %w", err))
	}

	if err := a.client.Close(); err != nil {
		errs = append(errs, fmt.Errorf("auth client close: %w", err))
	}

	if err := a.store.Close(); err != nil {
		errs = append(errs, fmt.Errorf("store close: %w", err))
	}

	if err := errors.Join(errs...); err != nil {
		a.logger.Warnf("app stopped with errors: %s", err)
		return err
	}

	a.logger.Info("app was gracefully shutdown")
	return nil
}
