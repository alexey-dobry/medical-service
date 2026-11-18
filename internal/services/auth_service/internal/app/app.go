package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/health"
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger/zap"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/pkg/redis"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/blacklist"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/credentials"
	authrpc "github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/grpc"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/rest"
	"google.golang.org/grpc"
)

type App interface {
	Run(context.Context) error
}

type app struct {
	GRPCServer        *grpc.Server
	RESTServer        *rest.RESTServer
	GRPCServerAddress string
	RESTServerAddress string

	postgres *postgres.Postgres
	redis    *redis.Redis
	health   *health.Checker

	logger          logger.Logger
	shutdownTimeout time.Duration
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.GRPCServerAddress = fmt.Sprintf(":%d", cfg.GRPC.Port)
	a.RESTServerAddress = fmt.Sprintf(":%d", cfg.REST.Port)
	a.shutdownTimeout = cfg.ShutdownTimeout

	a.postgres, err = connectToPostgres(cfg.CredentialsRepository, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to postgres: %s", err)
	}

	a.redis, err = connectToRedis(cfg.BlacklistRepository, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to redis: %s", err)
	}

	a.health = health.New(2*time.Second,
		health.Check{Name: "postgres", Critical: true, Fn: a.postgres.Ping},
		health.Check{Name: "redis", Critical: true, Fn: a.redis.Ping},
	)

	a.health.SetReady(true)

	credentialsRepo := credentials.NewCredentialsRepository(a.postgres)
	blacklistRepo := blacklist.NewBlacklistRepository(a.redis)

	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	created, err := ensureAdmin(bootstrapCtx, credentialsRepo, cfg.Admin)
	cancel()
	if err != nil {
		a.logger.Fatalf("Failed to bootstrap admin: %s", err)
	}
	if created {
		a.logger.Infof("Admin account %s created", cfg.Admin.Email)
	}

	jwtHandler, err := jwt.New(cfg.JWT)
	if err != nil {
		a.logger.Fatalf("Failed to create jwt handler: %s", err)
	}

	a.GRPCServer = authrpc.New(a.logger, credentialsRepo, blacklistRepo, jwtHandler)
	a.RESTServer = rest.New(a.logger, credentialsRepo, blacklistRepo, jwtHandler, a.health, cfg.REST)

	a.logger.Info("app was built")
	return &a
}

func (a *app) Run(ctx context.Context) error {
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

	serveErr := make(chan error, 2)

	go func() {
		a.logger.Infof("Starting gRPC server at address %s...", a.GRPCServerAddress)
		if err := a.GRPCServer.Serve(grpcListener); err != nil {
			serveErr <- fmt.Errorf("gRPC server: %w", err)
		}
	}()

	go func() {
		a.logger.Infof("Starting REST server at address %s...", a.RESTServerAddress)
		if err := a.RESTServer.Listener(restListener); err != nil {
			serveErr <- fmt.Errorf("REST server: %w", err)
		}
	}()

	a.logger.Info("App is running...")

	var runErr error
	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case runErr = <-serveErr:
		a.logger.Errorf("server failed, shutting down: %s", runErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	return errors.Join(runErr, a.shutdown(shutdownCtx))
}
