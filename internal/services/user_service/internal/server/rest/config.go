package rest

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
)

type Config struct {
	Port             int               `validate:"required" yaml:"port"`
	ReadTimeout      time.Duration     `validate:"required" yaml:"read_timeout"`
	WriteTimeout     time.Duration     `validate:"required" yaml:"write_timeout"`
	IdleTimeout      time.Duration     `validate:"required" yaml:"idle_timeout"`
	MiddlewareConfig middleware.Config `validate:"required" yaml:"middleware_config"`
}
