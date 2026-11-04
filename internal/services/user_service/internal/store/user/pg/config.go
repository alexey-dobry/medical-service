package pg

import "github.com/alexey-dobry/medical-service/internal/pkg/postgres"

type Config struct {
	URL                     string `validate:"required" yaml:"url" env:"USER_POSTGRES_URL"`
	postgres.PostgresConfig `yaml:",inline"`
}
