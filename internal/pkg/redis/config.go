package redis

import "time"

type RedisConfig struct {
	MaxCons      int           `validate:"required,gte=0" yaml:"max_connections"`
	ConnAttempts int           `validate:"required,gte=1" yaml:"connection_attempts"`
	ConnTimeout  time.Duration `validate:"required" yaml:"connection_retry_timeout"`
}
