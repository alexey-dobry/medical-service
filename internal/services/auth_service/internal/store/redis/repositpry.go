package redis

import (
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

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

	db = redis.NewClient(opt)

	return &BlacklistRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (br *BlacklistRepository) Close() error {
	return br.db.Close()
}
