package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	el "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	mn "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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
