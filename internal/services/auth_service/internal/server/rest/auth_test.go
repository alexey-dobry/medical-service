package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/mocks"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type env struct {
	server      *RESTServer
	credentials *mocks.MockCredentialsRepository
	blacklist   *mocks.MockBlacklistRepository
	jwt         *jwt.Handler
}

func newEnv(t *testing.T) *env {
	t.Helper()
	ctrl := gomock.NewController(t)
	e := &env{
		credentials: mocks.NewMockCredentialsRepository(ctrl),
		blacklist:   mocks.NewMockBlacklistRepository(ctrl),
		jwt:         testutil.JWT(t),
	}
	e.server = New(logger.Nop(), e.credentials, e.blacklist, e.jwt, Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	})
	return e
}

func (e *env) do(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := e.server.fiberApp.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func loginRequest(email, password string) *http.Request {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func cookie(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func withCookies(req *http.Request, cookies map[string]string) *http.Request {
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return req
}

func TestLogin_Success(t *testing.T) {
	e := newEnv(t)
	userID := uuid.NewString()
	hash, err := utils.HashPassword("correct-password")
	require.NoError(t, err)

	e.credentials.EXPECT().GetOneByMail(gomock.Any(), "doc@example.com").Return(model.Credentials{
		UserID: userID, Email: "doc@example.com", Role: pkgmodel.DoctorRole, PasswordHash: hash,
	}, nil)

	resp := e.do(t, loginRequest("doc@example.com", "correct-password"))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	accessCookie := cookie(resp, "accessToken")
	refreshCookie := cookie(resp, "refreshToken")
	require.NotNil(t, accessCookie)
	require.NotNil(t, refreshCookie)
	assert.True(t, accessCookie.HttpOnly)
	assert.Equal(t, "/auth", refreshCookie.Path, "refresh-cookie должна доходить до /auth/refresh и /auth/logout")

	access, err := e.jwt.ParseToken(accessCookie.Value, jwt.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, access.Subject)
	assert.Equal(t, pkgmodel.DoctorRole, access.Role)
	assert.Equal(t, jwt.AccessToken, access.Type)
	assert.NotEmpty(t, access.ID)

	refresh, err := e.jwt.ParseToken(refreshCookie.Value, jwt.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refresh.Subject)
	assert.Equal(t, pkgmodel.DoctorRole, refresh.Role)
	assert.Equal(t, jwt.RefreshToken, refresh.Type)
	assert.NotEqual(t, access.ID, refresh.ID)
}

func TestLogin_UnknownEmailAndWrongPasswordLookTheSame(t *testing.T) {
	e := newEnv(t)
	hash, err := utils.HashPassword("correct-password")
	require.NoError(t, err)

	e.credentials.EXPECT().GetOneByMail(gomock.Any(), "known@example.com").Return(model.Credentials{
		UserID: uuid.NewString(), Email: "known@example.com", Role: pkgmodel.PatientRole, PasswordHash: hash,
	}, nil)
	e.credentials.EXPECT().GetOneByMail(gomock.Any(), "unknown@example.com").Return(model.Credentials{}, repository.ErrNotFound)

	wrongPassword := e.do(t, loginRequest("known@example.com", "wrong-password"))
	unknownEmail := e.do(t, loginRequest("unknown@example.com", "whatever-password"))

	assert.Equal(t, http.StatusUnauthorized, wrongPassword.StatusCode)
	assert.Equal(t, http.StatusUnauthorized, unknownEmail.StatusCode)

	wrongBody, _ := io.ReadAll(wrongPassword.Body)
	unknownBody, _ := io.ReadAll(unknownEmail.Body)
	assert.Equal(t, wrongBody, unknownBody, "ответы не должны раскрывать, существует ли email")
	assert.Nil(t, cookie(wrongPassword, "accessToken"))
	assert.Nil(t, cookie(unknownEmail, "accessToken"))
}

func TestLogin_BadBody(t *testing.T) {
	e := newEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")

	assert.Equal(t, http.StatusBadRequest, e.do(t, req).StatusCode)
}

func TestRefresh_SucceedsWithExpiredAccessToken(t *testing.T) {
	e := newEnv(t)
	userID := uuid.NewString()

	refresh, refreshClaims, err := e.jwt.GenerateToken(userID, pkgmodel.PatientRole, jwt.RefreshToken)
	require.NoError(t, err)
	expiredAccess, _, err := testutil.ExpiredJWT(t).GenerateToken(userID, pkgmodel.PatientRole, jwt.AccessToken)
	require.NoError(t, err)

	gomock.InOrder(
		e.blacklist.EXPECT().IsSessionLoggedOut(gomock.Any(), refreshClaims.ID).Return(false, nil),
		// ротация: использованный refresh-токен больше нельзя применить
		e.blacklist.EXPECT().StoreLogoutSession(gomock.Any(), refreshClaims.ID, testutil.DurationAround(testutil.JWTConfig.RefreshTTL, 5*time.Second)).Return(nil),
	)
	// просроченный access-токен — не ошибка и в чёрный список не идёт
	e.blacklist.EXPECT().BlacklistAccessToken(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/refresh", nil), map[string]string{
		"refreshToken": refresh,
		"accessToken":  expiredAccess,
	})
	resp := e.do(t, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	newAccess, err := e.jwt.ParseToken(cookie(resp, "accessToken").Value, jwt.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, newAccess.Subject)
	assert.Equal(t, pkgmodel.PatientRole, newAccess.Role, "роль должна сохраняться при refresh")
	assert.Equal(t, jwt.AccessToken, newAccess.Type)

	newRefresh, err := e.jwt.ParseToken(cookie(resp, "refreshToken").Value, jwt.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, refreshClaims.ID, newRefresh.ID, "refresh-токен должен ротироваться")
}

func TestRefresh_RejectedAfterLogout(t *testing.T) {
	e := newEnv(t)
	refresh, refreshClaims, err := e.jwt.GenerateToken(uuid.NewString(), pkgmodel.PatientRole, jwt.RefreshToken)
	require.NoError(t, err)

	e.blacklist.EXPECT().IsSessionLoggedOut(gomock.Any(), refreshClaims.ID).Return(true, nil)

	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/refresh", nil), map[string]string{"refreshToken": refresh})
	resp := e.do(t, req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, cookie(resp, "accessToken"))
}

func TestRefresh_RejectsAccessTokenInsteadOfRefresh(t *testing.T) {
	e := newEnv(t)
	// Одинаковые секреты: подмену ловит только claim typ.
	shared, err := jwt.New(jwt.Config{AccessSecret: "same", RefreshSecret: "same", AccessTTL: time.Minute, RefreshTTL: time.Hour})
	require.NoError(t, err)
	e.server.jwtHandler = shared

	access, _, err := shared.GenerateToken(uuid.NewString(), pkgmodel.PatientRole, jwt.AccessToken)
	require.NoError(t, err)

	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/refresh", nil), map[string]string{"refreshToken": access})
	assert.Equal(t, http.StatusUnauthorized, e.do(t, req).StatusCode)
}

func TestRefresh_WithoutCookie(t *testing.T) {
	e := newEnv(t)
	assert.Equal(t, http.StatusUnauthorized, e.do(t, httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)).StatusCode)
}

func TestLogout_RevokesBothTokensByJTIWithRemainingTTL(t *testing.T) {
	e := newEnv(t)
	pair, err := e.jwt.GeneratePair(uuid.NewString(), pkgmodel.PatientRole)
	require.NoError(t, err)

	e.blacklist.EXPECT().
		BlacklistAccessToken(gomock.Any(), pair.AccessClaims.ID, testutil.DurationAround(testutil.JWTConfig.AccessTTL, 5*time.Second)).
		Return(nil)
	e.blacklist.EXPECT().
		StoreLogoutSession(gomock.Any(), pair.RefreshClaims.ID, testutil.DurationAround(testutil.JWTConfig.RefreshTTL, 5*time.Second)).
		Return(nil)

	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/logout", nil), map[string]string{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
	})
	resp := e.do(t, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	cleared := cookie(resp, "accessToken")
	require.NotNil(t, cleared)
	assert.Empty(t, cleared.Value)
}

func TestLogout_WithoutValidTokens(t *testing.T) {
	e := newEnv(t)
	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/logout", nil), map[string]string{"accessToken": "garbage"})
	assert.Equal(t, http.StatusUnauthorized, e.do(t, req).StatusCode)
}

func TestLogout_RedisFailureIs500(t *testing.T) {
	e := newEnv(t)
	pair, err := e.jwt.GeneratePair(uuid.NewString(), pkgmodel.PatientRole)
	require.NoError(t, err)

	e.blacklist.EXPECT().BlacklistAccessToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)

	req := withCookies(httptest.NewRequest(http.MethodPost, "/auth/logout", nil), map[string]string{"accessToken": pair.AccessToken})
	assert.Equal(t, http.StatusInternalServerError, e.do(t, req).StatusCode)
}
