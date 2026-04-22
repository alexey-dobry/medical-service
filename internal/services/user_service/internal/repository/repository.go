package repository

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddPatient(userData model.User) error

	GetPatient(ID uuid.UUID) (model.User, error)

	AddDoctor(userData model.User, doctorData model.DoctorAdditionalData) error

	GetDoctor(ID uuid.UUID) (model.User, model.DoctorAdditionalData, error)

	Close() error
}
