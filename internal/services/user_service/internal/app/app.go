package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	els "github.com/alexey-dobry/medical-service/internal/pkg/elasticsearch"
	"github.com/alexey-dobry/medical-service/internal/pkg/health"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger/zap"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
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
	Run(context.Context) error
}

type app struct {
	RESTServer        *rest.RESTServer
	RESTServerAddress string

	store         store.Store
	postgres      *postgres.Postgres
	elasticsearch *els.Elasticsearch
	client        client.Client

	health *health.Checker

	logger          logger.Logger
	shutdownTimeout time.Duration
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.RESTServerAddress = fmt.Sprintf(":%d", cfg.REST.Port)
	a.shutdownTimeout = cfg.ShutdownTimeout

	a.postgres, err = connectToPostgres(cfg.Store.Postgres, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to postgres: %s", err)
	}

	mnDB, mnPresigner, err := connectToMinio(cfg.Store.Minio, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to redis: %s", err)
	}

	a.elasticsearch, err = connectToElasticsearch(cfg.Store.Elasticsearch.ElasticsearchConfig, cfg.Store.Elasticsearch.DoctorIndex, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to elasticsearch: %s", err)
	}

	a.health = health.New(2*time.Second,
		health.Check{Name: "postgres", Critical: true, Fn: a.postgres.Ping},
		health.Check{Name: "elasticsearch", Critical: true, Fn: a.elasticsearch.Ping},
	)

	a.health.SetReady(true)

	store, err := user.New(cfg.Store, a.logger, a.postgres, mnDB, mnPresigner, a.elasticsearch)
	if err != nil {
		a.logger.Fatalf("Failed to create store instance: %s", err)
	}

	a.client, err = grpcClient.New(a.logger, cfg.Client)
	if err != nil {
		a.logger.Fatalf("Failed to create auth client: %s", err)
	}

	tokenValidator := middleware.NewAuthServiceValidator(a.client)

	registration := service.NewRegistration(a.client, store, a.logger, cfg.Compensation)

	a.RESTServer = rest.New(a.logger, store, a.client, registration, tokenValidator, a.health, cfg.REST)

	a.logger.Info("app was built")
	return &a
}

func (a *app) Run(ctx context.Context) error {
	restListener, err := net.Listen("tcp", a.RESTServerAddress)
	if err != nil {
		return fmt.Errorf("listen REST %s: %w", a.RESTServerAddress, err)
	}

	serveErr := make(chan error, 1)

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
