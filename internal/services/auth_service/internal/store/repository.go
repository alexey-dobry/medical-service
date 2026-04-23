package store

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

type CredentialsRepository interface {
	Add(userCredentials model.Credentials) error

	GetOneByMail(email string) (model.Credentials, error)

	GetOneByID(userID uuid.UUID) (model.Credentials, error)

	Delete(ID uuid.UUID) error

	Close() error
}

type BlacklistRepository interface {
	BlacklistAccessToken(jti string, expiresIn time.Duration) error

	IsAccessTokenBlacklisted(jti string) (bool, error)

	StoreLogoutSession(jti string, expiresIn time.Duration) error

	IsSessionLoggedOut(jti string) (bool, error)

	Close() error
}
