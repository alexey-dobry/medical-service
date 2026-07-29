package rest

import "time"

type Config struct {
	Port         int           `validate:"required" yaml:"port"`
	ReadTimeout  time.Duration `validate:"required" yaml:"read_timeout"`
	WriteTimeout time.Duration `validate:"required" yaml:"write_timeout"`
	IdleTimeout  time.Duration `validate:"required" yaml:"idle_timeout"`
}
