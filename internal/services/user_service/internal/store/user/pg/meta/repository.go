package meta

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
)

type Repository struct {
	db     *postgres.Postgres
	logger logger.Logger
}

func New(db *postgres.Postgres, logger logger.Logger) store.MetaRepository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}
