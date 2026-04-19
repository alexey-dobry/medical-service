package store

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store/pg"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store/redis"
)

type Store interface {
	Credentials() CredentialsRepository

	Blacklist() BlacklistRepository

	Close() error
}

type authStore struct {
	credentialsRepository CredentialsRepository
	blacklistRepository   BlacklistRepository
}

func New(logger logger.Logger, cfg Config) (Store, error) {
	credentialsRepository, err := pg.New(logger, cfg.PgConfig)
	if err != nil {
		return nil, err
	}

	blacklistRepository, err := redis.New(logger, cfg.RedisConfig)
	if err != nil {
		return nil, err
	}

	return &authStore{
		credentialsRepository: credentialsRepository,
		blacklistRepository:   blacklistRepository,
	}, nil
}

func (as *authStore) Credentials() CredentialsRepository {
	return as.credentialsRepository
}

func (as *authStore) Blacklist() BlacklistRepository {
	return as.blacklistRepository
}

func (as *authStore) Close() error {
	err := as.credentialsRepository.Close()
	if err != nil {
		return err
	}

	err = as.blacklistRepository.Close()
	if err != nil {
		return err
	}

	return nil
}
