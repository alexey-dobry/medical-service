package redis

type Config struct {
	MaxRetries   int    `validate:"required" yaml:"max_retries"`
	RetryDelay   int    `validate:"required" yaml:"retry_delay"`
	Port         int    `validate:"required" yaml:"port"`
	Host         string `validate:"required" yaml:"host"`
	User         string `validate:"required" yaml:"user"`
	Password     string `validate:"required" yaml:"password"`
	DatabaseName string `validate:"required" yaml:"database"`
}
