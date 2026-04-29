package minio

import (
	"context"
	"io"
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/minio/minio-go/v7"
)

func (r *Repository) Put(
	key string,
	reader io.Reader,
	size int64,
	contentType string,
) error {

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := r.db.PutObject(
		context.Background(),
		r.bucket,
		key,
		reader,
		size,
		opts,
	)

	return err
}

func (r *Repository) Get(key string) (string, error) {
	_, err := r.db.StatObject(
		context.Background(),
		r.bucket,
		key,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return "", err
	}

	presignedURL, err := r.db.PresignedGetObject(
		context.Background(),
		r.bucket,
		key,
		7*24*time.Hour,
		nil,
	)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (r *Repository) Delete(key string) error {
	return r.db.RemoveObject(
		context.Background(),
		r.bucket,
		key,
		minio.RemoveObjectOptions{},
	)
}

func (r *Repository) Stat(key string) (model.StorageObjectInfo, error) {
	info, err := r.db.StatObject(
		context.Background(),
		r.bucket,
		key,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return model.StorageObjectInfo{}, err
	}

	return model.StorageObjectInfo{
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}
