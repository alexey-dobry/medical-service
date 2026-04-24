package meta

import (
	"database/sql"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
)

type Repository struct {
	db     *sql.DB
	logger logger.Logger
}

func New(db *sql.DB, logger logger.Logger) store.MetaRepository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) Close() error {
	return r.db.Close()
}
