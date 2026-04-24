package elasticsearch

import "time"

type Config struct {
	Addresses []string
	Username  string
	Password  string

	DoctorIndex string

	MaxRetries int
	RetryDelay time.Duration
}
