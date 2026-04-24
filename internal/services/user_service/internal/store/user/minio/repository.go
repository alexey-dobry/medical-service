package minio

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/minio/minio-go/v7"
)

type Repository struct {
	db     *minio.Client
	bucket string
	logger logger.Logger
}

func New(db *minio.Client, logger logger.Logger, bucket string) store.PhotosRepository {
	return &Repository{
		db:     db,
		bucket: bucket,
		logger: logger,
	}
}
