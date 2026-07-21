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

func (r *BlacklistRepository) BlacklistToken(ctx context.Context, jti string, expiresIn time.Duration) error {
	status := r.Client.Set(ctx, fmt.Sprintf("blacklist:%s", jti), 1, expiresIn)
	return fmt.Errorf("blacklistRepository - BlacklistToken() - failed to blacklist token: %w", status.Err())
}

func (r *BlacklistRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	result, err := r.Client.Exists(ctx, fmt.Sprintf("blacklist:%s", jti)).Result()
	if err != nil {
		return false, fmt.Errorf("blacklistRepository - IsTokenBlacklisted() - failed to check if entry exists: %w", err)
	}

	return result == 1, nil
}

func (r *BlacklistRepository) Logout(
	ctx context.Context,
	refreshJTI string,
	accessJTI string,
	refreshExpiresIn time.Duration,
	accessExpiresIn time.Duration,
) error {
	tx := r.Client.TxPipeline()

	tx.Set(ctx, fmt.Sprintf("blacklist:%s", refreshJTI), 1, refreshExpiresIn)
	tx.Set(ctx, fmt.Sprintf("blacklist:%s", accessJTI), 1, accessExpiresIn)

	_, err := tx.Exec(ctx)
	if err != nil {
		return fmt.Errorf("blacklistRepository - Logout() - failed to logout: %w", err)
	}

	return nil
}
