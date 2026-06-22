package redis

import "time"

type Option func(*Redis)

func MaxCons(v int) Option {
	return func(r *Redis) {
		r.maxCons = v
	}
}

func ConnAttempts(v int) Option {
	return func(r *Redis) {
		r.connAttempts = v
	}
}

func ConnTimeout(v time.Duration) Option {
	return func(r *Redis) {
		r.connTimeout = v
	}
}
