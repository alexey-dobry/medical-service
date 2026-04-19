package redis

import (
	"context"
	"fmt"
	"time"
)

func (br *BlacklistRepository) BlacklistAccessToken(jti string, expiresIn time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", jti)
	err := br.db.Set(context.Background(), key, 1, expiresIn).Err()
	return err
}

func (br *BlacklistRepository) IsAccessTokenBlacklisted(jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", jti)
	result, err := br.db.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (br *BlacklistRepository) StoreLogoutSession(jti string, expiresIn time.Duration) error {
	key := fmt.Sprintf("logout_session:%s", jti)
	err := br.db.Set(context.Background(), key, 1, expiresIn).Err()
	return err
}

func (br *BlacklistRepository) IsSessionLoggedOut(jti string) (bool, error) {
	key := fmt.Sprintf("logout_session:%s", jti)
	result, err := br.db.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
