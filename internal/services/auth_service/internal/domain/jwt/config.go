package jwt

type Config struct {
	AccessSecret  string `validate:"required" yaml:"access-secret"`
	RefreshSecret string `validate:"required" yaml:"refresh-secret"`
	TTL           TTL    `validate:"required" yaml:"ttl"`
}

type TTL struct {
	AccessTTL  string `validate:"required,duration" yaml:"access-ttl"`
	RefreshTTL string `validate:"required,duration" yaml:"refresh-ttl"`
}
