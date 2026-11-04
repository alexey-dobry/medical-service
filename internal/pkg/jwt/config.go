package jwt

import "time"

type Config struct {
	AccessSecret  string        `validate:"required" yaml:"-" env:"ACCESS_SECRET"`
	RefreshSecret string        `validate:"required" yaml:"-" env:"REFRESH_SECRET"`
	AccessTTL     time.Duration `validate:"required" yaml:"access-ttl"`
	RefreshTTL    time.Duration `validate:"required" yaml:"refresh-ttl"`
}
