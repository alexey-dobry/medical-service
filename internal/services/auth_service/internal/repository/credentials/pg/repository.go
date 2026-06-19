package pg

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
)

type PostgresCredentialsRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func New(logger logger.Logger, cfg PostgresConfig) (*PostgresCredentialsRepository, error) {
	logger = logger.WithFields("layer", "pgstore")

	// pgs connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DatabaseName)

	// opening sql connection with retry mechanism
	var db *sql.DB
	var err error
	for range cfg.MaxRetries {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			break
		}

		time.Sleep(time.Second * time.Duration(cfg.RetryDelay))
	}
	if err != nil {
		return nil, err
	}

	// try to ping db
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	logger.Info("repository was connected")

	return &PostgresCredentialsRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (cr *PostgresCredentialsRepository) Close() error {
	return cr.db.Close()
}
