package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	_defaultMaxCons      = 0
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Redis struct {
	maxCons      int
	connAttempts int
	connTimeout  time.Duration

	Client *redis.Client
}

func (r *Redis) Ping(ctx context.Context) error { return r.Client.Ping(ctx).Err() }

func New(url string, opts ...Option) (*Redis, error) {
	rd := &Redis{
		maxCons:      _defaultMaxCons,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(rd)
	}

	clientConfig, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis - New() - failed to parse conn url: %w", err)
	}

	clientConfig.DialerRetryTimeout = rd.connTimeout
	clientConfig.DialerRetries = rd.connAttempts
	clientConfig.MaxActiveConns = rd.maxCons

	rd.Client = redis.NewClient(clientConfig)

	return rd, nil
}

func (r *Redis) Close() {
	if r.Client != nil {
		_ = r.Client.Close()
	}
}
