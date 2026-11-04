package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client/mocks"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newApp поднимает защищённый роут и admin-only роут, которые возвращают
// identity, полученную в хендлере. Валидация идёт через настоящий
// AuthServiceValidator с замоканным gRPC-клиентом.
func newApp(t *testing.T) (*fiber.App, *mocks.MockClient) {
	t.Helper()
	authClient := mocks.NewMockClient(gomock.NewController(t))

	auth := middleware.ValidateJWT(middleware.Config{}, middleware.NewAuthServiceValidator(authClient))
	echo := func(c *fiber.Ctx) error {
		identity, err := middleware.IdentityFrom(c)
		if err != nil {
			return c.SendStatus(fiber.StatusTeapot)
		}
		return c.JSON(fiber.Map{"user_id": identity.UserID.String(), "role": identity.Role})
	}

	app := fiber.New()
	app.Get("/protected", auth, echo)
	app.Get("/admin", auth, middleware.RequireRole(model.AdminRole), echo)
	app.Get("/unprotected", echo)
	return app, authClient
}

func request(t *testing.T, app *fiber.App, path, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "accessToken", Value: token})
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func TestValidateJWT_NoCookie(t *testing.T) {
	app, authClient := newApp(t)
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), gomock.Any()).Times(0)

	assert.Equal(t, http.StatusUnauthorized, request(t, app, "/protected", "").StatusCode)
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	app, authClient := newApp(t)
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), "garbage").
		Return(client.TokenInfo{}, status.Error(codes.Unauthenticated, "invalid or expired token"))

	assert.Equal(t, http.StatusUnauthorized, request(t, app, "/protected", "garbage").StatusCode)
}

func TestValidateJWT_RevokedToken(t *testing.T) {
	app, authClient := newApp(t)
	// после логаута auth_service находит jti в чёрном списке
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), "logged-out").
		Return(client.TokenInfo{}, status.Error(codes.Unauthenticated, "token revoked"))

	assert.Equal(t, http.StatusUnauthorized, request(t, app, "/protected", "logged-out").StatusCode)
}

func TestValidateJWT_AuthServiceDownIs500(t *testing.T) {
	app, authClient := newApp(t)
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), gomock.Any()).
		Return(client.TokenInfo{}, status.Error(codes.Unavailable, "connection refused"))

	assert.Equal(t, http.StatusInternalServerError, request(t, app, "/protected", "token").StatusCode)
}

func TestValidateJWT_ValidTokenExposesIdentity(t *testing.T) {
	app, authClient := newApp(t)
	userID := uuid.New()
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), "good").
		Return(client.TokenInfo{UserID: userID.String(), Role: string(model.PatientRole)}, nil)

	resp := request(t, app, "/protected", "good")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, userID.String(), body["user_id"])
	assert.Equal(t, string(model.PatientRole), body["role"])
}

func TestRequireRole(t *testing.T) {
	for role, want := range map[model.Role]int{
		model.PatientRole: http.StatusForbidden,
		model.DoctorRole:  http.StatusForbidden,
		model.AdminRole:   http.StatusOK,
	} {
		t.Run(string(role), func(t *testing.T) {
			app, authClient := newApp(t)
			authClient.EXPECT().ValidateAccessToken(gomock.Any(), gomock.Any()).
				Return(client.TokenInfo{UserID: uuid.NewString(), Role: string(role)}, nil)

			assert.Equal(t, want, request(t, app, "/admin", "token").StatusCode)
		})
	}
}

func TestIdentityFromWithoutMiddlewareDoesNotPanic(t *testing.T) {
	app, _ := newApp(t)
	assert.Equal(t, http.StatusTeapot, request(t, app, "/unprotected", "").StatusCode)
}

func TestValidateJWT_UnknownRoleFromAuthIsRejected(t *testing.T) {
	app, authClient := newApp(t)
	authClient.EXPECT().ValidateAccessToken(gomock.Any(), gomock.Any()).
		Return(client.TokenInfo{UserID: uuid.NewString(), Role: "ROOT"}, nil)

	assert.Equal(t, http.StatusInternalServerError, request(t, app, "/protected", "token").StatusCode)
}
