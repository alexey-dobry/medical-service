package grpc

type Config struct {
	Port int `validate:"required" yaml:"port"`
}
