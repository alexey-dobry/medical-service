package postgres

import "time"

type Option func(*Postgres)

func MaxPoolSize(v int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = v
	}
}

func ConnAttempts(v int) Option {
	return func(p *Postgres) {
		p.connAttempts = v
	}
}

func ConnTimeout(v time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = v
	}
}
