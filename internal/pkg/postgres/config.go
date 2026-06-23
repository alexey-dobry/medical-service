package postgres

import "time"

type PostgresConfig struct {
	MaxPoolSize  int           `validate:"required,gte=1" yaml:"max_pool_size"`
	ConnAttempts int           `validate:"required,gte=1" yaml:"connection_attempts"`
	ConnTimeout  time.Duration `validate:"required" yaml:"connection_retry_timeout"`
}
