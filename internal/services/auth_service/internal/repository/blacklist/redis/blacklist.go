package redis

import (
	"context"
	"fmt"
	"time"
)

func (br *RedisBlacklistRepository) BlacklistAccessToken(jti string, expiresIn time.Duration) error {
	return br.db.Set(context.Background(), fmt.Sprintf("blacklist:%s", jti), 1, expiresIn).Err()
}

func (br *RedisBlacklistRepository) IsAccessTokenBlacklisted(jti string) (bool, error) {
	result, err := br.db.Exists(context.Background(), fmt.Sprintf("blacklist:%s", jti)).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (br *RedisBlacklistRepository) StoreLogoutSession(jti string, expiresIn time.Duration) error {
	return br.db.Set(context.Background(), fmt.Sprintf("logout_session:%s", jti), 1, expiresIn).Err()
}

func (br *RedisBlacklistRepository) IsSessionLoggedOut(jti string) (bool, error) {
	result, err := br.db.Exists(context.Background(), fmt.Sprintf("logout_session:%s", jti)).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
