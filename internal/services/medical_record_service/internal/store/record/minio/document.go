package minio

import (
	"context"
	"io"

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

func (r *Repository) Get(key string) (io.ReadCloser, int64, string, error) {
	stat, err := r.db.StatObject(
		context.Background(),
		r.bucket,
		key,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return nil, 0, "", err
	}

	object, err := r.db.GetObject(
		context.Background(),
		r.bucket,
		key,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, 0, "", err
	}

	return object, stat.Size, stat.ContentType, nil
}

func (r *Repository) Delete(key string) error {
	return r.db.RemoveObject(
		context.Background(),
		r.bucket,
		key,
		minio.RemoveObjectOptions{},
	)
}
