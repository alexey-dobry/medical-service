package rest

type Config struct {
	Port int `validate:"required" yaml:"port"`
}
