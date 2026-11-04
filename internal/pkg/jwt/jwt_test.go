package jwt_test

import (
	"testing"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = jwt.Config{
	AccessSecret:  "access-secret",
	RefreshSecret: "refresh-secret",
	AccessTTL:     15 * time.Minute,
	RefreshTTL:    24 * time.Hour,
}

func newHandler(t *testing.T, opts ...jwt.Option) *jwt.Handler {
	t.Helper()
	h, err := jwt.New(cfg, opts...)
	require.NoError(t, err)
	return h
}

func TestGenerateAndParse(t *testing.T) {
	h := newHandler(t)
	userID := uuid.NewString()

	pair, err := h.GeneratePair(userID, model.DoctorRole)
	require.NoError(t, err)

	access, err := h.ParseToken(pair.AccessToken, jwt.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, access.Subject)
	assert.Equal(t, model.DoctorRole, access.Role)
	assert.Equal(t, jwt.AccessToken, access.Type)
	assert.NotEmpty(t, access.ID)
	assert.WithinDuration(t, time.Now().Add(cfg.AccessTTL), access.ExpiresAt.Time, 2*time.Second)
	assert.NotNil(t, access.IssuedAt)

	refresh, err := h.ParseToken(pair.RefreshToken, jwt.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refresh.Subject)
	assert.Equal(t, model.DoctorRole, refresh.Role, "refresh token must keep the role")
	assert.Equal(t, jwt.RefreshToken, refresh.Type)
	assert.NotEqual(t, access.ID, refresh.ID, "every token needs its own jti")
}

func TestEveryTokenHasUniqueJTI(t *testing.T) {
	h := newHandler(t)
	seen := map[string]bool{}
	for range 100 {
		_, claims, err := h.GenerateToken("user", model.PatientRole, jwt.AccessToken)
		require.NoError(t, err)
		require.False(t, seen[claims.ID], "duplicate jti %s", claims.ID)
		seen[claims.ID] = true
	}
}

func TestParseExpiredToken(t *testing.T) {
	past := func() time.Time { return time.Now().Add(-time.Hour) }
	token, _, err := newHandler(t, jwt.WithClock(past)).GenerateToken("user", model.PatientRole, jwt.AccessToken)
	require.NoError(t, err)

	_, err = newHandler(t).ParseToken(token, jwt.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestParseForeignSignature(t *testing.T) {
	other, err := jwt.New(jwt.Config{
		AccessSecret:  "attacker-secret",
		RefreshSecret: "attacker-secret",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Minute,
	})
	require.NoError(t, err)

	token, _, err := other.GenerateToken("user", model.AdminRole, jwt.AccessToken)
	require.NoError(t, err)

	_, err = newHandler(t).ParseToken(token, jwt.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestParseRejectsOtherAlgorithms(t *testing.T) {
	claims := jwt.Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   "user",
			ID:        uuid.NewString(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: model.AdminRole,
		Type: jwt.AccessToken,
	}

	t.Run("HS512 with the right secret", func(t *testing.T) {
		token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, claims).SignedString([]byte(cfg.AccessSecret))
		require.NoError(t, err)

		_, err = newHandler(t).ParseToken(token, jwt.AccessToken)
		assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
	})

	t.Run("alg none", func(t *testing.T) {
		token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, claims).SignedString(jwtlib.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = newHandler(t).ParseToken(token, jwt.AccessToken)
		assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
	})
}

func TestParseRejectsWrongTokenType(t *testing.T) {
	h, err := jwt.New(jwt.Config{
		AccessSecret:  "shared",
		RefreshSecret: "shared",
		AccessTTL:     time.Minute,
		RefreshTTL:    time.Hour,
	})
	require.NoError(t, err)

	pair, err := h.GeneratePair("user", model.PatientRole)
	require.NoError(t, err)

	_, err = h.ParseToken(pair.RefreshToken, jwt.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid, "refresh token must not work as access token")

	_, err = h.ParseToken(pair.AccessToken, jwt.RefreshToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid, "access token must not work as refresh token")
}

func TestParseGarbage(t *testing.T) {
	for _, token := range []string{"", "not-a-jwt", "a.b.c"} {
		_, err := newHandler(t).ParseToken(token, jwt.AccessToken)
		assert.ErrorIs(t, err, jwt.ErrTokenInvalid, "token %q", token)
	}
}

func TestNewValidatesConfig(t *testing.T) {
	for name, c := range map[string]jwt.Config{
		"empty access secret":  {RefreshSecret: "r", AccessTTL: time.Minute, RefreshTTL: time.Minute},
		"empty refresh secret": {AccessSecret: "a", AccessTTL: time.Minute, RefreshTTL: time.Minute},
		"zero ttl":             {AccessSecret: "a", RefreshSecret: "r", RefreshTTL: time.Minute},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := jwt.New(c)
			assert.ErrorIs(t, err, jwt.ErrInvalidConfig)
		})
	}
}
