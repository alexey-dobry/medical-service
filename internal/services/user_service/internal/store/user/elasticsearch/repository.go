package elasticsearch

import (
	"context"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/elastic/go-elasticsearch/v8"
)

type Repository struct {
	db     *elasticsearch.Client
	index  string
	logger logger.Logger
}

func New(db *elasticsearch.Client, logger logger.Logger, index string) store.SearchRepository {
	return &Repository{
		db:     db,
		index:  index,
		logger: logger,
	}
}

func (r *Repository) Close() error {
	return r.db.Close(context.Background())
}
