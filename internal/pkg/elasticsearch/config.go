package elasticsearch

import "time"

// конфиг elasticsearch встраивается в конфиг сервиса
type ElasticsearchConfig struct {
	Addresses       []string      `validate:"required,min=1" yaml:"addresses"`
	Username        string        `yaml:"username" env:"ELASTIC_USERNAME"`
	Password        string        `yaml:"password" env:"ELASTIC_PASSWORD"`
	MaxRetries      int           `validate:"gte=0" yaml:"max_retries"`
	ResponseTimeout time.Duration `yaml:"response_timeout"`
	ConnAttempts    int           `validate:"required,gte=1" yaml:"connection_attempts"`
	ConnTimeout     time.Duration `validate:"required" yaml:"connection_retry_timeout"`
}
