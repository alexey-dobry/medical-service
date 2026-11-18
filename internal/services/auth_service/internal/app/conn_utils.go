package app

import (
	"context"
	"fmt"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/pkg/redis"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/blacklist"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/credentials"
)

func connectToPostgres(cfg credentials.Config, logger logger.Logger) (*postgres.Postgres, error) {
	logger = logger.WithFields("layer", "pg_store")

	db, err := postgres.New(cfg.URL,
		postgres.MaxPoolSize(cfg.MaxPoolSize),
		postgres.ConnAttempts(cfg.ConnAttempts),
		postgres.ConnTimeout(cfg.ConnTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	logger.Info("connection established")

	return db, nil
}

func connectToRedis(cfg blacklist.Config, logger logger.Logger) (*redis.Redis, error) {
	logger = logger.WithFields("layer", "redis_store")

	rd, err := redis.New(cfg.URL, redis.MaxCons(cfg.MaxCons))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
		err = rd.Ping(ctx)
		cancel()
		if err == nil {
			break
		}
		if attempt == cfg.ConnAttempts {
			rd.Close()
			return nil, fmt.Errorf("redis not reachable after %d attempts: %w", attempt, err)
		}
		logger.Warnf("redis ping failed (attempt %d/%d): %s", attempt, cfg.ConnAttempts, err)
		time.Sleep(cfg.ConnTimeout)
	}

	logger.Info("connection established")
	return rd, nil
}
