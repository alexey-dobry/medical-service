package credentials

import "github.com/alexey-dobry/medical-service/internal/pkg/postgres"

type Config struct {
	URL                     string `validate:"required" yaml:"url" env:"AUTH_POSTGRES_URL"`
	postgres.PostgresConfig `yaml:",inline"`
}
