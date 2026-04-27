package store

import (
	"io"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddPatient(userData model.User) error

	GetPatient(ID uuid.UUID) (model.User, error)

	AddDoctor(userData model.User, doctorData model.DoctorAdditionalData) error

	GetDoctor(ID uuid.UUID) (model.User, model.DoctorAdditionalData, error)

	DeleteUser(ID uuid.UUID) error

	Close() error
}

type MetaRepository interface {
	Create(photo model.Photo) error

	GetByID(ID uuid.UUID) (model.Photo, error)

	Delete(ID uuid.UUID) (string, error)

	Close() error
}

type SearchRepository interface {
	AddDoctor(doctorData model.DoctorSearchParams) error

	SearchDoctor(searchParams model.DoctorSearchParams) (uuid.UUID, error)

	DeleteDoctor(ID uuid.UUID) error

	Close() error
}

type PhotosRepository interface {
	Put(key string, reader io.Reader, size int64, contentType string) error

	Get(key string) (io.ReadCloser, error)

	Delete(key string) error

	Stat(key string) (model.StorageObjectInfo, error)
}
