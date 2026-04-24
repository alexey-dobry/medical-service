package store

import (
	"io"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	intModel "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddPatient(userData model.User) error

	GetPatient(ID uuid.UUID) (model.User, error)

	AddDoctor(userData model.User, doctorData model.DoctorAdditionalData) error

	GetDoctor(ID uuid.UUID) (model.User, model.DoctorAdditionalData, error)

	Close() error
}

type MetaRepository interface {
	Create(photo intModel.Photo) error

	GetByID(ID string) (intModel.Photo, error)

	Delete(ID string) error

	Close() error
}

type SearchRepository interface {
	AddDoctor(doctorData model.DoctorSearchParams) error

	SearchDoctor(searchParams model.DoctorSearchParams) (uuid.UUID, error)

	Close() error
}

type PhotosRepository interface {
	Put(key string, reader io.Reader, size int64, contentType string) error

	Get(key string) (io.ReadCloser, error)

	Delete(key string) error

	Stat(key string) (*intModel.StorageObjectInfo, error)
}
