package repository

import (
	"context"
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/entity"
)

type CredentialsRepository interface {
	Add(ctx context.Context, userCredentials entity.Credentials) error
	GetOneByMail(ctx context.Context, email string) (entity.Credentials, error)
	GetOneByID(ctx context.Context, userID string) (entity.Credentials, error)
	Delete(ctx context.Context, userID string) error
}

type BlacklistRepository interface {
	BlacklistToken(ctx context.Context, jti string, expiresIn time.Duration) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	Logout(ctx context.Context, refreshJTI, accessJTI string, refreshExpiresIn, accessExpiresIn time.Duration) error
}
