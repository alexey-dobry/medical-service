package jwt

type Config struct {
	AccessSecret  string `validate:"required"`
	RefreshSecret string `validate:"required"`
	TTL           TTL    `validate:"required" yaml:"ttl"`
}

type TTL struct {
	AccessTTL  string `validate:"required,duration" yaml:"access-ttl"`
	RefreshTTL string `validate:"required,duration" yaml:"refresh-ttl"`
}
