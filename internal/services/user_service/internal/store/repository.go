package store

//go:generate mockgen -source=repository.go -destination=mocks/repository_mock.go -package=mocks
//go:generate mockgen -source=store.go -destination=mocks/store_mock.go -package=mocks

import (
	"context"
	"io"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddPatient(ctx context.Context, userData model.User) error

	GetPatient(ctx context.Context, ID uuid.UUID) (model.User, error)

	AddDoctor(ctx context.Context, userData model.User, doctorData model.DoctorAdditionalData) error

	GetDoctor(ctx context.Context, ID uuid.UUID) (model.User, model.DoctorAdditionalData, error)

	DeleteUser(ctx context.Context, ID uuid.UUID) error

	UpdateUser(ctx context.Context, ID uuid.UUID, data map[string]interface{}) error
}

type MetaRepository interface {
	Create(ctx context.Context, photo model.Photo) error

	GetByID(ctx context.Context, ID uuid.UUID) (model.Photo, error)

	DeleteByUserID(ctx context.Context, userID uuid.UUID) (string, error)
}

type SearchRepository interface {
	AddDoctor(ctx context.Context, doctorData model.DoctorSearchParams) error

	SearchDoctor(ctx context.Context, searchParams model.DoctorSearchParams) ([]uuid.UUID, error)

	DeleteDoctor(ctx context.Context, ID uuid.UUID) error

	Close() error
}

type PhotosRepository interface {
	Put(key string, reader io.Reader, size int64, contentType string) error

	Get(key string) (string, error)

	Delete(key string) error

	Stat(key string) (model.StorageObjectInfo, error)
}
