package elasticsearch

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/elasticsearch"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
)

type Repository struct {
	db     *elasticsearch.Elasticsearch
	index  string
	logger logger.Logger
}

func New(db *elasticsearch.Elasticsearch, logger logger.Logger, index string) store.SearchRepository {
	return &Repository{
		db:     db,
		index:  index,
		logger: logger,
	}
}

func (r *Repository) Close() error {
	return r.db.Close()
}
