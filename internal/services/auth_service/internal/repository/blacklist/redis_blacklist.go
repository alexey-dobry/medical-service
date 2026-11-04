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

func (r *BlacklistRepository) BlacklistAccessToken(ctx context.Context, jti string, expiresIn time.Duration) error {
	if err := r.Client.Set(ctx, "blacklist:"+jti, 1, expiresIn).Err(); err != nil {
		return fmt.Errorf("blacklistRepository - BlacklistAccessToken() - failed to blacklist token: %w", err)
	}

	return nil
}

func (r *BlacklistRepository) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	result, err := r.Client.Exists(ctx, "blacklist:"+jti).Result()
	if err != nil {
		return false, fmt.Errorf("blacklistRepository - IsAccessTokenBlacklisted() - failed to check if entry exists: %w", err)
	}

	return result == 1, nil
}

func (r *BlacklistRepository) StoreLogoutSession(ctx context.Context, jti string, expiresIn time.Duration) error {
	if err := r.Client.Set(ctx, "logout_session:"+jti, 1, expiresIn).Err(); err != nil {
		return fmt.Errorf("blacklistRepository - StoreLogoutSession() - failed to store logout session: %w", err)
	}

	return nil
}

func (r *BlacklistRepository) IsSessionLoggedOut(ctx context.Context, jti string) (bool, error) {
	result, err := r.Client.Exists(ctx, "logout_session:"+jti).Result()
	if err != nil {
		return false, fmt.Errorf("blacklistRepository - IsSessionLoggedOut() - failed to check if entry exists: %w", err)
	}

	return result == 1, nil
}
