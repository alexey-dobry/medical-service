package credentials

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolationCode — SQLSTATE Postgres для unique_violation.
const uniqueViolationCode = "23505"

type CredentialsRepository struct {
	*postgres.Postgres
}

func NewCredentialsRepository(pg *postgres.Postgres) *CredentialsRepository {
	return &CredentialsRepository{pg}
}

func (r *CredentialsRepository) Add(ctx context.Context, credentials model.Credentials) error {
	sql, args, err := r.Builder.
		Insert("credentials").
		Columns("user_id", "email", "password_hash", "role").
		Values(
			credentials.UserID,
			credentials.Email,
			credentials.PasswordHash,
			credentials.Role,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("credentialsRepository - Add() - failed to build query string: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return repository.ErrAlreadyExists
	} else if err != nil {
		return fmt.Errorf("credentialsRepository - Add() - failed to add new entry: %w", err)
	}

	return nil
}

func (r *CredentialsRepository) GetOneByMail(ctx context.Context, email string) (model.Credentials, error) {
	sql, args, err := r.Builder.
		Select("user_id", "email", "password_hash", "role").
		From("credentials").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return model.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByMail() - failed to build query string: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)

	var c model.Credentials
	if err := row.Scan(&c.UserID, &c.Email, &c.PasswordHash, &c.Role); errors.Is(err, pgx.ErrNoRows) {
		return model.Credentials{}, repository.ErrNotFound
	} else if err != nil {
		return model.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByMail() - failed to query: %w", err)
	}

	return c, nil
}

func (r *CredentialsRepository) GetOneByID(ctx context.Context, id uuid.UUID) (model.Credentials, error) {
	sql, args, err := r.Builder.
		Select("user_id", "email", "password_hash", "role").
		From("credentials").
		Where(sq.Eq{"user_id": id.String()}).
		ToSql()
	if err != nil {
		return model.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByID() - failed to build query string: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)

	var c model.Credentials
	if err := row.Scan(&c.UserID, &c.Email, &c.PasswordHash, &c.Role); errors.Is(err, pgx.ErrNoRows) {
		return model.Credentials{}, repository.ErrNotFound
	} else if err != nil {
		return model.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByID() - failed to query: %w", err)
	}

	return c, nil
}

func (r *CredentialsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.Builder.
		Delete("credentials").
		Where(sq.Eq{"user_id": id.String()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("credentialsRepository - Delete() - failed to build query string: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("credentialsRepository - Delete() - failed to delete entry: %w", err)
	}

	return nil
}
