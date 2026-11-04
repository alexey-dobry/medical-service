package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger/zap"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	grpcClient "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client/grpc"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/service"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user"
)

type App interface {
	Run(context.Context)
}

type app struct {
	RESTServer        *rest.RESTServer
	RESTServerAddress string

	store  store.Store
	client client.Client

	logger logger.Logger
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.RESTServerAddress = fmt.Sprintf(":%d", cfg.REST.Port)

	a.store, err = user.New(cfg.Store, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to create store instance: %s", err)
	}

	a.client, err = grpcClient.New(a.logger, cfg.Client)
	if err != nil {
		a.logger.Fatalf("Failed to create auth client: %s", err)
	}

	tokenValidator := middleware.NewAuthServiceValidator(a.client)

	registration := service.NewRegistration(a.client, a.store, a.logger, cfg.Compensation)

	a.RESTServer = rest.New(a.logger, a.store, a.client, registration, tokenValidator, cfg.REST)

	a.logger.Info("app was built")
	return &a
}

func (a *app) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	restListener, err := net.Listen("tcp", a.RESTServerAddress)
	if err != nil {
		a.logger.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		a.logger.Infof("Starting rest server at address %s...", a.RESTServerAddress)
		if err := a.RESTServer.Listener(restListener); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				a.logger.Errorf("Rest server error: %s", err)
				cancel()
			}
		}
	}()
	a.logger.Info("App is running...")

	select {
	case <-quit:
		a.logger.Info("shutdown signal received")
	case <-ctx.Done():
		a.logger.Info("context canceled")
	}

	a.logger.Info("stopping all services")

	cancel()

	if err := restListener.Close(); err != nil {
		a.logger.Warnf("REST listener closing ended with error: %s", err)
	}

	wg.Wait()

	if err := a.client.Close(); err != nil {
		a.logger.Warnf("auth client closing ended with error: %s", err)
	}

	if err := a.store.Close(); err != nil {
		a.logger.Warnf("store closing ended with error: %s", err)
	}

	a.logger.Info("app was gracefully shutdown")
}
