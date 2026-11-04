package elasticsearch

type Config struct {
	Addresses []string `validate:"required,min=1" yaml:"addresses"`
	Username  string   `yaml:"username" env:"ELASTIC_USERNAME"`
	Password  string   `yaml:"password" env:"ELASTIC_PASSWORD"`

	DoctorIndex string `validate:"required" yaml:"doctor_index"`

	MaxRetries int `yaml:"max_retries" env-default:"3"`
}
