package blacklist

import (
	"context"
	"fmt"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/redis"
)

type BlacklistRepository struct {
	*redis.Redis
}

func NewBlacklistRepository(rd *redis.Redis) *BlacklistRepository {
	return &BlacklistRepository{rd}
}

func (r *BlacklistRepository) BlacklistAccessToken(jti string, expiresIn time.Duration) error {
	status := r.Client.Set(context.Background(), fmt.Sprintf("blacklist:%s", jti), 1, expiresIn)
	return fmt.Errorf("blacklistRepository - BlacklistAccessToken() - failed to blacklist token: %w", status.Err())
}

func (r *BlacklistRepository) IsAccessTokenBlacklisted(jti string) (bool, error) {
	result, err := r.Client.Exists(context.Background(), fmt.Sprintf("blacklist:%s", jti)).Result()
	if err != nil {
		return false, fmt.Errorf("blacklistRepository - IsAccessTokenBlacklisted() - failed to check if entry exists: %w", err)
	}

	return result == 1, nil
}

func (r *BlacklistRepository) StoreLogoutSession(jti string, expiresIn time.Duration) error {
	status := r.Client.Set(context.Background(), fmt.Sprintf("logout_session:%s", jti), 1, expiresIn)
	return fmt.Errorf("blacklistRepository - StoreLogoutSession() - failed to store logout session: %w", status.Err())
}

func (r *BlacklistRepository) IsSessionLoggedOut(jti string) (bool, error) {
	result, err := r.Client.Exists(context.Background(), fmt.Sprintf("logout_session:%s", jti)).Result()
	if err != nil {
		return false, fmt.Errorf("blacklistRepository - IsSessionLoggedOut() - failed to check if entry exists: %w", err)
	}

	return result == 1, nil
}
