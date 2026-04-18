package repository

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

type CredentialsRepository interface {
	Add(userCredentials model.Credentials) error

	GetOneByMail(email string) (model.Credentials, error)

	GetOneByID(userID uuid.UUID) (model.Credentials, error)

	Close() error
}
