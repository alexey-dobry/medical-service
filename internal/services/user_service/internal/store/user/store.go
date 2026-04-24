package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	el "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	mn "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg/meta"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg/user"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Store struct {
	metaRepository   store.MetaRepository
	userRepository   store.UserRepository
	searchRepository store.SearchRepository
	photosRepository store.PhotosRepository
}

func New(cfg Config, logger logger.Logger) (store.Store, error) {
	logger = logger.WithFields("layer", "store")

	pgDB, err := connectToPostgres(cfg.pgConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	mnDB, err := connectToMinio(cfg.minioConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %w", err)
	}

	elDB, err := connectToElastic(cfg.elasticsearchConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	metaRepository := meta.New(pgDB, logger)
	userRepository := user.New(pgDB, logger)

	photosRepository := mn.New(mnDB, logger, cfg.minioConfig.Bucket)

	searchRepository := el.New(elDB, logger, cfg.elasticsearchConfig.DoctorIndex)

	return &Store{
		metaRepository:   metaRepository,
		userRepository:   userRepository,
		photosRepository: photosRepository,
		searchRepository: searchRepository,
	}, nil
}

func (s *Store) User() store.UserRepository {
	return s.userRepository
}

func (s *Store) Meta() store.MetaRepository {
	return s.metaRepository
}

func (s *Store) Search() store.SearchRepository {
	return s.searchRepository
}

func (s *Store) Photos() store.PhotosRepository {
	return s.photosRepository
}

func (s *Store) Close() error {
	err := s.userRepository.Close()
	if err != nil {
		return err
	}

	err = s.metaRepository.Close()
	if err != nil {
		return err
	}

	err = s.searchRepository.Close()
	if err != nil {
		return err
	}

	return nil
}

func connectToPostgres(cfg pg.Config, logger logger.Logger) (*sql.DB, error) {
	logger = logger.WithFields("layer", "pg_store")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DatabaseName)

	var db *sql.DB
	var err error
	for range cfg.MaxRetries {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			break
		}

		logger.Warnf("connect retry: %w", err)

		time.Sleep(time.Second * time.Duration(cfg.RetryDelay))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	logger.Info("connection established")

	return db, nil
}

func connectToMinio(cfg mn.Config, logger logger.Logger) (*minio.Client, error) {
	logger = logger.WithFields("layer", "minio_store")

	dsn := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	accessKey := cfg.AccessKey
	secretAccessKey := cfg.SecretKey
	useSSL := false

	var db *minio.Client
	var err error
	for range cfg.MaxRetries {
		db, err = minio.New(dsn, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretAccessKey, ""),
			Secure: useSSL,
		})
		if err == nil {
			break
		}

		logger.Warnf("connect retry: %w", err)

		time.Sleep(time.Second * time.Duration(cfg.RetryDelay))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	exists, err := db.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exist: %w", err)
	}

	if !exists {
		if err := db.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	logger.Info("connection established")

	return db, nil
}

func connectToElastic(cfg el.Config, logger logger.Logger) (*elasticsearch.Client, error) {
	logger = logger.WithFields("layer", "elasticsearch_store")

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

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

	err = initDoctorsIndex(client)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	logger.Info("connection established")

	return client, nil
}

func initDoctorsIndex(client *elasticsearch.Client) error {
	ctx := context.Background()

	existsRes, err := client.Indices.Exists(
		[]string{"doctors"},
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
		"doctors",
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
