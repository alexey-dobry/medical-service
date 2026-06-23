package repository

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/entity"
	"github.com/google/uuid"
)

type CredentialsRepository interface {
	Add(userCredentials entity.Credentials) error

	GetOneByMail(email string) (entity.Credentials, error)

	GetOneByID(userID uuid.UUID) (entity.Credentials, error)

	Delete(id uuid.UUID) error
}

type BlacklistRepository interface {
	BlacklistAccessToken(jti string, expiresIn time.Duration) error

	IsAccessTokenBlacklisted(jti string) (bool, error)

	StoreLogoutSession(jti string, expiresIn time.Duration) error

	IsSessionLoggedOut(jti string) (bool, error)
}
