package user

import (
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg"
)

type Config struct {
	Postgres      pg.Config            `yaml:"postgres"`
	Minio         minio.Config         `yaml:"minio"`
	Elasticsearch elasticsearch.Config `yaml:"elasticsearch"`
}
