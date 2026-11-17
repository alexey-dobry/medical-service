package elasticsearch

import "time"

type Option func(*Elasticsearch)

func Credentials(username, password string) Option {
	return func(e *Elasticsearch) {
		e.username = username
		e.password = password
	}
}

func MaxRetries(v int) Option {
	return func(e *Elasticsearch) {
		e.maxRetries = v
	}
}

func ResponseTimeout(v time.Duration) Option {
	return func(e *Elasticsearch) {
		e.responseTimeout = v
	}
}

func ConnAttempts(v int) Option {
	return func(e *Elasticsearch) {
		e.connAttempts = v
	}
}

func ConnTimeout(v time.Duration) Option {
	return func(e *Elasticsearch) {
		e.connTimeout = v
	}
}

func FromConfig(cfg ElasticsearchConfig) []Option {
	opts := []Option{
		Credentials(cfg.Username, cfg.Password),
		MaxRetries(cfg.MaxRetries),
		ConnAttempts(cfg.ConnAttempts),
		ConnTimeout(cfg.ConnTimeout),
	}
	if cfg.ResponseTimeout > 0 {
		opts = append(opts, ResponseTimeout(cfg.ResponseTimeout))
	}
	return opts
}
