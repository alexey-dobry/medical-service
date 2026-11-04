package rest

import "time"

type Config struct {
	Port         int           `validate:"required" yaml:"port"`
	ReadTimeout  time.Duration `validate:"required" yaml:"read_timeout"`
	WriteTimeout time.Duration `validate:"required" yaml:"write_timeout"`
	IdleTimeout  time.Duration `validate:"required" yaml:"idle_timeout"`
	// InsecureCookies снимает флаг Secure, чтобы cookie работали по обычному
	// HTTP при локальной разработке. За HTTPS должно оставаться false.
	InsecureCookies bool `yaml:"insecure_cookies"`
}
