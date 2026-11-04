package minio

import "time"

type Config struct {
	Endpoint       string        `validate:"required" yaml:"endpoint"`
	PublicEndpoint string        `yaml:"public_endpoint"`
	AccessKey      string        `validate:"required" yaml:"access_key" env:"MINIO_ACCESS_KEY"`
	SecretKey      string        `validate:"required" yaml:"secret_key" env:"MINIO_SECRET_KEY"`
	Bucket         string        `validate:"required" yaml:"bucket"`
	Region         string        `yaml:"region" env-default:"eu-west-2"`
	UseSSL         bool          `yaml:"use_ssl"`
	ConnAttempts   int           `validate:"required,gte=1" yaml:"connection_attempts"`
	ConnTimeout    time.Duration `validate:"required" yaml:"connection_retry_timeout"`
	URLExpiry      time.Duration `yaml:"presigned_url_expiry" env-default:"1h"`
}
