package meta

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) Create(ctx context.Context, photo model.Photo) error {
	_, err := r.db.Pool.Exec(ctx,
		"INSERT INTO photo_meta (id,name,mime_type,size,user_id,storage_key) VALUES ($1,$2,$3,$4,$5,$6)",
		photo.ID,
		photo.Name,
		photo.MimeType,
		photo.Size,
		photo.UserID,
		photo.StorageKey,
	)
	if err != nil {
		return fmt.Errorf("metaRepository - Create(): %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, ID uuid.UUID) (model.Photo, error) {
	row := r.db.Pool.QueryRow(ctx, "SELECT id,name,mime_type,size,user_id,storage_key FROM photo_meta WHERE id = $1", ID)

	var p model.Photo
	err := row.Scan(&p.ID, &p.Name, &p.MimeType, &p.Size, &p.UserID, &p.StorageKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Photo{}, store.ErrNotFound
	} else if err != nil {
		return model.Photo{}, fmt.Errorf("metaRepository - GetByID(): %w", err)
	}

	return p, nil
}

// DeleteByUserID удаляет метаданные фото пользователя и возвращает ключ
// объекта в хранилище, чтобы вызывающий мог удалить сам файл.
func (r *Repository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	var key string
	err := r.db.Pool.QueryRow(ctx, "DELETE FROM photo_meta WHERE user_id = $1 RETURNING storage_key", userID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", fmt.Errorf("metaRepository - DeleteByUserID(): %w", err)
	}

	return key, nil
}
