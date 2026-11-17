package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	els "github.com/alexey-dobry/medical-service/internal/pkg/elasticsearch"
	"github.com/alexey-dobry/medical-service/internal/pkg/health"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger/zap"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/pkg/redis"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	grpcClient "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client/grpc"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/service"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user"
	el "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	mn "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const shutdownDelay = time.Second * 5

type App interface {
	Run(context.Context)
}

type app struct {
	RESTServer        *rest.RESTServer
	RESTServerAddress string

	store         store.Store
	postgres      *postgres.Postgres
	redis         *redis.Redis
	elasticsearch *els.Elasticsearch
	client        client.Client

	health *health.Checker

	logger logger.Logger
}

func New(cfg config.Config) App {
	var a app
	var err error

	a.health = health.New(2*time.Second,
		health.Check{Name: "postgres", Critical: true, Fn: a.postgres.Ping},
		health.Check{Name: "redis", Critical: true, Fn: a.redis.Ping},
		health.Check{Name: "elasticsearch", Critical: true, Fn: a.elasticsearch.Ping},
	)

	a.health.SetReady(true)

	a.logger = zap.NewLogger(cfg.Logger).WithFields("layer", "app")

	a.RESTServerAddress = fmt.Sprintf(":%d", cfg.REST.Port)

	pgDB, err := connectToPostgres(cfg.Store.Postgres, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to postgres: %s", err)
	}

	mnDB, mnPresigner, err := connectToMinio(cfg.Store.Minio, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to redis: %s", err)
	}

	esDB, err := connectToElasticsearch(cfg.Store.Elasticsearch, a.logger)
	if err != nil {
		a.logger.Fatalf("Failed to connect to elasticsearch: %s", err)
	}

	store, err := user.New(cfg.Store, a.logger, pgDB, mnDB, mnPresigner, esDB)
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

func connectToPostgres(cfg pg.Config, logger logger.Logger) (*postgres.Postgres, error) {
	logger = logger.WithFields("layer", "pg_store")

	db, err := postgres.New(cfg.URL,
		postgres.MaxPoolSize(cfg.MaxPoolSize),
		postgres.ConnAttempts(cfg.ConnAttempts),
		postgres.ConnTimeout(cfg.ConnTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	logger.Info("connection established")

	return db, nil
}

func connectToMinio(cfg mn.Config, logger logger.Logger) (*minio.Client, *minio.Client, error) {
	logger = logger.WithFields("layer", "minio_store")

	opts := func() *minio.Options {
		return &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.UseSSL,
			Region: cfg.Region,
		}
	}

	db, err := minio.New(cfg.Endpoint, opts())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client: %w", err)
	}

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	presigner, err := minio.New(publicEndpoint, opts())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create presign client: %w", err)
	}

	var exists bool
	for attempt := 1; attempt <= cfg.ConnAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
		exists, err = db.BucketExists(ctx, cfg.Bucket)
		cancel()
		if err == nil {
			break
		}

		logger.Warnf("connect retry, attempts left %d: %s", cfg.ConnAttempts-attempt, err)

		time.Sleep(cfg.ConnTimeout)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check if bucket exist: %w", err)
	}

	if !exists {
		if err := db.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	logger.Info("connection established")

	return db, presigner, nil
}

func connectToElasticsearch(cfg el.Config, logger logger.Logger) (*elasticsearch.Client, error) {
	logger = logger.WithFields("layer", "elasticsearch_store")

	maxRetries := cfg.MaxRetries
	esCfg := elasticsearch.Config{
		Addresses:  cfg.Addresses,
		Username:   cfg.Username,
		Password:   cfg.Password,
		MaxRetries: maxRetries,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ping error [%s]", res.Status())
	}

	err = initDoctorsIndex(client, cfg.DoctorIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	logger.Info("connection established")

	return client, nil
}

func initDoctorsIndex(client *elasticsearch.Client, index string) error {
	ctx := context.Background()

	existsRes, err := client.Indices.Exists(
		[]string{index},
		client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("initDoctorsIndex: check existence: %w", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == http.StatusOK {
		return nil
	}

	if existsRes.StatusCode != http.StatusNotFound {
		return fmt.Errorf("initDoctorsIndex: unexpected status checking index [%s]", existsRes.Status())
	}

	mappingBody, err := json.Marshal(model.DoctorsIndexMapping)
	if err != nil {
		return fmt.Errorf("initDoctorsIndex: marshal mapping: %w", err)
	}

	createRes, err := client.Indices.Create(
		index,
		client.Indices.Create.WithBody(bytes.NewReader(mappingBody)),
		client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("initDoctorsIndex: create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("initDoctorsIndex: elasticsearch error [%s]", createRes.Status())
	}

	return nil
}
