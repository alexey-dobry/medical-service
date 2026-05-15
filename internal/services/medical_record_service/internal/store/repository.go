package store

import (
	"io"

	model "github.com/alexey-dobry/medical-service/internal/services/medical_record_service/internal/domain/model"
	"github.com/google/uuid"
)

type RecordRepository interface {
	Add(medicalRecord model.MedicalRecord) error

	GetMany(patientID uuid.UUID, limit, offset int) ([]model.MedicalRecord, error)

	GetOne(id uuid.UUID) (model.MedicalRecord, error)

	Delete(id uuid.UUID) error

	Close() error
}

type MetaRepository interface {
	Add(documentMeta model.DocumentMeta) error

	Get(id uuid.UUID) (model.DocumentMeta, error)

	Delete(id uuid.UUID) (string, error)

	Close() error
}

type DocumentRepository interface {
	Put(key string, reader io.Reader, size int64, contentType string) error

	Get(key string) (io.ReadCloser, int64, string, error)

	Delete(key string) error
}
