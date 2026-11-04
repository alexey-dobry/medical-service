package rest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// photoURL возвращает presigned-ссылку на фото или пустую строку, если у пользователя нет фото.
func (s *RESTServer) photoURL(ctx context.Context, photoID string) (string, error) {
	if photoID == "" {
		return "", nil
	}

	id, err := uuid.Parse(photoID)
	if err != nil {
		return "", fmt.Errorf("invalid photo id %q: %w", photoID, err)
	}

	meta, err := s.store.Meta().GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	return s.store.Photos().Get(meta.StorageKey)
}
