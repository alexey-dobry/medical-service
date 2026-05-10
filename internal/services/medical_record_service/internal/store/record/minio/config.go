package minio

type Config struct {
	MaxRetries int    `validate:"required" yaml:"max_retries"`
	RetryDelay int    `validate:"required" yaml:"retry_delay"`
	Port       string `validate:"required" yaml:"port"`
	Host       string `validate:"required" yaml:"host"`
	AccessKey  string `validate:"required" yaml:"access_key"`
	SecretKey  string `validate:"required" yaml:"secret_key"`
	Bucket     string `validate:"required" yaml:"bucket"`
	UseSSL     bool   `validate:"required" yaml:"user_ssl"`
}
