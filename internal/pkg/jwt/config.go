package jwt

import "time"

type Config struct {
	AccessSecret  string        `validate:"required"`
	RefreshSecret string        `validate:"required"`
	AccessTTL     time.Duration `validate:"required" yaml:"access-ttl"`
	RefreshTTL    time.Duration `validate:"required" yaml:"refresh-ttl"`
}
