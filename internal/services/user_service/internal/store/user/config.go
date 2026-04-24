package user

import (
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg"
)

type Config struct {
	pgConfig            pg.Config            `yaml:"pg_config"`
	minioConfig         minio.Config         `yaml:"minio_config"`
	elasticsearchConfig elasticsearch.Config `yaml:"elasticsearch_config"`
}
