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
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	authrpc "github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/grpc"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/rest"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store"
	"google.golang.org/grpc"
)

type App interface {
	Run(context.Context)
}

type app struct {
	GRPCServer        *grpc.Server
	RESTServer        *rest.RESTServer
	GRPCServerAddress string
	RESTServerAddress string

	store store.Store

	logger logger.Logger
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.GRPCServerAddress = fmt.Sprintf(":%s", cfg.GRPC.Port)
	a.RESTServerAddress = fmt.Sprintf(":%s", cfg.REST.Port)

	a.store, err = store.New(a.logger, cfg.Store)
	if err != nil {
		a.logger.Fatalf("Failed to create store instance: %s", err)
	}

	jwtHandler, err := jwt.NewHandler(cfg.JWT)
	if err != nil {
		a.logger.Fatalf("Failed to create jwt handler: %s", err)
	}

	a.GRPCServer = authrpc.New(a.logger, a.store.Credentials(), jwtHandler)
	a.RESTServer = rest.New(a.logger, a.store, jwtHandler, cfg.REST)

	a.logger.Info("app was built")
	return &a
}

func (a *app) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	grpcListener, err := net.Listen("tcp", a.GRPCServerAddress)
	if err != nil {
		a.logger.Fatal(err)
	}

	restListener, err := net.Listen("tcp", a.RESTServerAddress)
	if err != nil {
		a.logger.Fatal(err)
	}

	var wg sync.WaitGroup

	// gRPC servers
	wg.Add(1)
	go func() {
		defer wg.Done()

		a.logger.Infof("Starting public grpc server at address %s...", a.GRPCServerAddress)
		if err := a.GRPCServer.Serve(grpcListener); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				a.logger.Errorf("Grpc server error: %s", err)
				cancel()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		a.logger.Infof("Starting internal grpc server at address %s...", a.RESTServerAddress)
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
	if err := grpcListener.Close(); err != nil {
		a.logger.Warnf("Internal net listener closing ended with error: %s", err)
	}

	if err := restListener.Close(); err != nil {
		a.logger.Warnf("Public net listener closing ended with error: %s", err)
	}

	wg.Wait()

	if err := a.store.Close(); err != nil {
		a.logger.Warnf("store closing ended with error: %s", err)
	}

	a.logger.Info("app was gracefully shutdown")
}
