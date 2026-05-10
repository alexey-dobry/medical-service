package record

import (
	"github.com/alexey-dobry/medical-service/internal/services/medical_record_service/internal/store/record/minio"
	"github.com/alexey-dobry/medical-service/internal/services/medical_record_service/internal/store/record/pg"
)

type Config struct {
	pgConfig    pg.Config    `yaml:"pg_config"`
	minioConfig minio.Config `yaml:"minio_config"`
}
