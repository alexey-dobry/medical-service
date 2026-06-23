package credentials

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/entity"
	"github.com/google/uuid"
)

type CredentialsRepository struct {
	*postgres.Postgres
}

func NewCredentialsRepository(pg *postgres.Postgres) *CredentialsRepository {
	return &CredentialsRepository{pg}
}

func (r *CredentialsRepository) Add(ctx context.Context, credentials entity.Credentials) error {
	sql, args, err := r.Builder.
		Insert("credentials").
		Columns("id", "user_id", "email", "password_hash", "role", "created_at", "updated_at").
		Values(
			credentials.ID,
			credentials.UserID,
			credentials.Email,
			credentials.PasswordHash,
			credentials.Role,
			credentials.CreatedAt,
			credentials.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("credentialsRepository - Add() - failed to build query string: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("credentialsRepository - Add() - failed to add new entry: %w", err)
	}

	return nil
}

func (r *CredentialsRepository) GetOneByMail(ctx context.Context, email string) (entity.Credentials, error) {
	sql, args, err := r.Builder.
		Select("user_id", "password_hash", "role").
		From("credentials").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return entity.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByMail() - failed to build query string: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)

	var c entity.Credentials
	if err := row.Scan(&c.UserID, &c.PasswordHash, &c.Role); err != nil {
		return entity.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByMail() - failed to query: %w", err)
	}

	return c, nil
}

func (r *CredentialsRepository) GetOneByID(ctx context.Context, id uuid.UUID) (entity.Credentials, error) {
	sql, args, err := r.Builder.
		Select("email", "password_hash", "role").
		From("credentials").
		Where(sq.Eq{"user_id": id.String()}).
		ToSql()
	if err != nil {
		return entity.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByID() - failed to build query string: %w", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)

	var c entity.Credentials
	if err := row.Scan(&c.Email, &c.PasswordHash, &c.Role); err != nil {
		return entity.Credentials{}, fmt.Errorf("credentialsRepository - GetOneByID() - failed to query: %w", err)
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
