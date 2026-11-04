package service

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

// MaxPhotoSize — максимальный размер фото профиля.
const MaxPhotoSize = 5 << 20 // 5 MiB

// ErrInvalidPhoto — загруженный файл не является допустимым изображением.
var ErrInvalidPhoto = errors.New("invalid photo")

// PhotoUpload — фото профиля, полученное от клиента.
type PhotoUpload struct {
	Name     string
	MimeType string
	Size     int64
	Content  io.Reader
}

// Validate проверяет тип и размер.
func (p *PhotoUpload) Validate() error {
	if !strings.HasPrefix(p.MimeType, "image/") {
		return fmt.Errorf("%w: unsupported content type %q", ErrInvalidPhoto, p.MimeType)
	}
	if p.Size <= 0 || p.Size > MaxPhotoSize {
		return fmt.Errorf("%w: size must be between 1 byte and %d bytes", ErrInvalidPhoto, MaxPhotoSize)
	}
	return nil
}

// uploadPhoto сохраняет файл в объектное хранилище и возвращает метаданные.
func (r *Registration) uploadPhoto(upload *PhotoUpload, userID string) (model.Photo, error) {
	photoID := uuid.NewString()
	photo := model.Photo{
		ID:         photoID,
		Name:       upload.Name,
		MimeType:   upload.MimeType,
		Size:       upload.Size,
		UserID:     userID,
		StorageKey: fmt.Sprintf("photos/%s", photoID),
	}

	if err := r.store.Photos().Put(photo.StorageKey, upload.Content, photo.Size, photo.MimeType); err != nil {
		return model.Photo{}, fmt.Errorf("failed to upload photo: %w", err)
	}

	return photo, nil
}

// removePhoto — best-effort удаление загруженного объекта.
func (r *Registration) removePhoto(photo model.Photo) {
	if err := r.store.Photos().Delete(photo.StorageKey); err != nil {
		r.logger.Errorf("Failed to remove orphaned photo %s: %s", photo.StorageKey, err)
	}
}
