package redis

import (
	"fmt"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const maxRetries = 10
const delay = 2 * time.Second

type BlacklistRepository struct {
	db     *redis.Client
	logger logger.Logger
}

func New(logger logger.Logger, cfg Config) (*BlacklistRepository, error) {
	redisDSN := fmt.Sprintf("redis://%s:%s@%s:%s/%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DatabaseName,
	)

	var db *redis.Client
	opt, err := redis.ParseURL(redisDSN)
	if err != nil {
		return nil, err
	}

	for range cfg.MaxRetries {
		db = redis.NewClient(opt)
		if err == nil {
			break
		}

		time.Sleep(time.Second * time.Duration(cfg.RetryDelay))
	}

	if err != nil {
		return nil, err
	}

	return &BlacklistRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (br *BlacklistRepository) Close() error {
	return br.db.Close()
}
