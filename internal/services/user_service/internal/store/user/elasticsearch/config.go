package elasticsearch

import "github.com/alexey-dobry/medical-service/internal/pkg/elasticsearch"

type Config struct {
	elasticsearch.ElasticsearchConfig `yaml:",inline"`

	DoctorIndex string `validate:"required" yaml:"doctor_index"`
}
