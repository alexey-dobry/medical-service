package store

import (
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store/pg"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store/redis"
)

type Config struct {
	PgConfig    pg.Config    `yaml:"pg_config"`
	RedisConfig redis.Config `yaml:"redis_config"`
}
