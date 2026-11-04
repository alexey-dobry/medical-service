package blacklist

import "github.com/alexey-dobry/medical-service/internal/pkg/redis"

type Config struct {
	URL               string `validate:"required" yaml:"url" env:"AUTH_REDIS_URL"`
	redis.RedisConfig `yaml:",inline"`
}
