package minio

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/minio/minio-go/v7"
)

type Repository struct {
	db *minio.Client
	// presigner подписывает URL для публичного эндпоинта; к MinIO он не обращается,
	// потому что регион задан явно.
	presigner *minio.Client
	bucket    string
	urlExpiry time.Duration
	logger    logger.Logger
}

func New(db, presigner *minio.Client, logger logger.Logger, bucket string, urlExpiry time.Duration) store.PhotosRepository {
	return &Repository{
		db:        db,
		presigner: presigner,
		bucket:    bucket,
		urlExpiry: urlExpiry,
		logger:    logger,
	}
}
