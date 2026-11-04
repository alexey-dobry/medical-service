package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	Run(context.Context)
}

type app struct {
	GRPCServer        *grpc.Server
	RESTServer        *rest.RESTServer
	GRPCServerAddress string
	RESTServerAddress string

	postgres *postgres.Postgres
	redis    *redis.Redis

	logger logger.Logger
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.GRPCServerAddress = fmt.Sprintf(":%d", cfg.GRPC.Port)
	a.RESTServerAddress = fmt.Sprintf(":%d", cfg.REST.Port)

	pgCfg := cfg.CredentialsRepository
	a.postgres, err = postgres.New(pgCfg.URL,
		postgres.MaxPoolSize(pgCfg.MaxPoolSize),
		postgres.ConnAttempts(pgCfg.ConnAttempts),
		postgres.ConnTimeout(pgCfg.ConnTimeout),
	)
	if err != nil {
		a.logger.Fatalf("Failed to connect to postgres: %s", err)
	}

	rdCfg := cfg.BlacklistRepository
	a.redis, err = redis.New(rdCfg.URL,
		redis.MaxCons(rdCfg.MaxCons),
		redis.ConnAttempts(rdCfg.ConnAttempts),
		redis.ConnTimeout(rdCfg.ConnTimeout),
	)
	if err != nil {
		a.logger.Fatalf("Failed to connect to redis: %s", err)
	}

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
	a.RESTServer = rest.New(a.logger, credentialsRepo, blacklistRepo, jwtHandler, cfg.REST)

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

	wg.Add(1)
	go func() {
		defer wg.Done()

		a.logger.Infof("Starting gRPC server at address %s...", a.GRPCServerAddress)
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

		a.logger.Infof("Starting REST server at address %s...", a.RESTServerAddress)
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
		a.logger.Warnf("gRPC listener closing ended with error: %s", err)
	}

	if err := restListener.Close(); err != nil {
		a.logger.Warnf("REST listener closing ended with error: %s", err)
	}

	wg.Wait()

	a.postgres.Close()
	a.redis.Close()

	a.logger.Info("app was gracefully shutdown")
}
