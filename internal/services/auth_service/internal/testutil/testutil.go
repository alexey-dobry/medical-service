// Пакет testutil содержит хелперы для тестов auth_service.
package testutil

import (
	"testing"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var JWTConfig = jwt.Config{
	AccessSecret:  "test-access-secret",
	RefreshSecret: "test-refresh-secret",
	AccessTTL:     15 * time.Minute,
	RefreshTTL:    24 * time.Hour,
}

// JWT возвращает хендлер с тестовым конфигом.
func JWT(t *testing.T, opts ...jwt.Option) *jwt.Handler {
	t.Helper()
	h, err := jwt.New(JWTConfig, opts...)
	require.NoError(t, err)
	return h
}

// ExpiredJWT выпускает просроченные токены (истекли час назад),
// подписанные тестовыми секретами.
func ExpiredJWT(t *testing.T) *jwt.Handler {
	t.Helper()
	past := func() time.Time { return time.Now().Add(-JWTConfig.RefreshTTL - time.Hour) }
	return JWT(t, jwt.WithClock(past))
}

// DurationAround проверяет, что длительность отличается от want
// не более чем на tolerance. Нужен для TTL в чёрном списке.
func DurationAround(want, tolerance time.Duration) gomock.Matcher {
	return gomock.Cond(func(x any) bool {
		d, ok := x.(time.Duration)
		return ok && d <= want && d >= want-tolerance
	})
}
