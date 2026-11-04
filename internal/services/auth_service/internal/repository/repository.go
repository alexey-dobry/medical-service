package repository

//go:generate mockgen -source=repository.go -destination=mocks/repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/google/uuid"
)

type CredentialsRepository interface {
	Add(ctx context.Context, userCredentials model.Credentials) error

	GetOneByMail(ctx context.Context, email string) (model.Credentials, error)

	GetOneByID(ctx context.Context, userID uuid.UUID) (model.Credentials, error)

	Delete(ctx context.Context, id uuid.UUID) error
}

type BlacklistRepository interface {
	BlacklistAccessToken(ctx context.Context, jti string, expiresIn time.Duration) error

	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)

	StoreLogoutSession(ctx context.Context, jti string, expiresIn time.Duration) error

	IsSessionLoggedOut(ctx context.Context, jti string) (bool, error)
}
