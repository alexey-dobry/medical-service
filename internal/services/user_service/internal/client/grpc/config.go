package grpc

import "time"

type Config struct {
	// Timeout ограничивает каждый вызов auth_service, чтобы зависшая
	// зависимость не вешала запрос (проверка токена идёт на каждом запросе).
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`

	Host string `validate:"required" yaml:"host"`
	Port string `validate:"required" yaml:"port"`
	// CertFilePath включает TLS, если задан. В приватной docker-сети
	// соединение может оставаться без TLS — так же, как у auth gRPC-сервера.
	CertFilePath       string `yaml:"cert_file_path"`
	ServerNameOverride string `yaml:"server_override_name"`
}
