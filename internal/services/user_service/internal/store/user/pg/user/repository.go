package user

import (
	"errors"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolationCode is the Postgres SQLSTATE for unique_violation.
const uniqueViolationCode = "23505"

type Repository struct {
	db     *postgres.Postgres
	logger logger.Logger
}

func New(db *postgres.Postgres, logger logger.Logger) store.UserRepository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return store.ErrAlreadyExists
	}
	return err
}
