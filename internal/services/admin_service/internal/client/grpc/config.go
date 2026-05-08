package grpc

type Config struct {
	CertFilePath       string `validate:"required" yaml:"cert_file_path"`
	ServerNameOverride string `validate:"required" yaml:"server_override_name"`
	Host               string `validate:"required" yaml:"host"`
	Port               string `validate:"required" yaml:"port"`
}
