package auth

import (
	"context"
	"testing"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type env struct {
	server      *GRPCServer
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
	e.server = New(logger.Nop(), e.credentials, e.blacklist, e.jwt)
	return e
}

func registerRequest(role string) *pb.RegisterRequest {
	return &pb.RegisterRequest{
		UserId:   uuid.NewString(),
		Email:    "patient@example.com",
		Password: "secret-password",
		Role:     role,
	}
}

func TestRegisterPatient_Success(t *testing.T) {
	e := newEnv(t)
	req := registerRequest(string(pkgmodel.PatientRole))

	e.credentials.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c model.Credentials) error {
		assert.Equal(t, req.UserId, c.UserID)
		assert.Equal(t, req.Email, c.Email)
		assert.Equal(t, pkgmodel.PatientRole, c.Role)
		assert.True(t, utils.CheckPasswordHash(req.Password, c.PasswordHash), "пароль должен храниться как bcrypt-хеш")
		return nil
	})

	resp, err := e.server.RegisterPatient(context.Background(), req)
	require.NoError(t, err)

	access, err := e.jwt.ParseToken(resp.AccessToken, jwt.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, req.UserId, access.Subject)
	assert.Equal(t, pkgmodel.PatientRole, access.Role)
	assert.Equal(t, jwt.AccessToken, access.Type)
	assert.NotEmpty(t, access.ID)

	refresh, err := e.jwt.ParseToken(resp.RefreshToken, jwt.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, req.UserId, refresh.Subject)
	assert.Equal(t, pkgmodel.PatientRole, refresh.Role)
	assert.Equal(t, jwt.RefreshToken, refresh.Type)
	assert.NotEqual(t, access.ID, refresh.ID)
}

func TestRegisterPatient_DuplicateEmail(t *testing.T) {
	e := newEnv(t)
	e.credentials.EXPECT().Add(gomock.Any(), gomock.Any()).Return(repository.ErrAlreadyExists)

	_, err := e.server.RegisterPatient(context.Background(), registerRequest(string(pkgmodel.PatientRole)))
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestRegister_InvalidRole(t *testing.T) {
	for name, tc := range map[string]struct {
		role   string
		doctor bool
	}{
		"unknown role":                  {role: "ROOT"},
		"empty role":                    {role: ""},
		"admin through patient sign-up": {role: string(pkgmodel.AdminRole)},
		"doctor through patient flow":   {role: string(pkgmodel.DoctorRole)},
		"patient through doctor flow":   {role: string(pkgmodel.PatientRole), doctor: true},
	} {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t)
			e.credentials.EXPECT().Add(gomock.Any(), gomock.Any()).Times(0)

			var err error
			if tc.doctor {
				_, err = e.server.RegisterDoctor(context.Background(), registerRequest(tc.role))
			} else {
				_, err = e.server.RegisterPatient(context.Background(), registerRequest(tc.role))
			}
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestRegisterDoctor_StoresDoctorRole(t *testing.T) {
	e := newEnv(t)
	e.credentials.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c model.Credentials) error {
		assert.Equal(t, pkgmodel.DoctorRole, c.Role)
		return nil
	})

	_, err := e.server.RegisterDoctor(context.Background(), registerRequest(string(pkgmodel.DoctorRole)))
	require.NoError(t, err)
}

func TestDeleteUser_InvalidUUID(t *testing.T) {
	e := newEnv(t)
	e.credentials.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)

	_, err := e.server.DeleteUser(context.Background(), &pb.DeleteUserRequest{UserId: "not-a-uuid"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestValidateAccessToken(t *testing.T) {
	userID := uuid.NewString()

	t.Run("valid token", func(t *testing.T) {
		e := newEnv(t)
		token, claims, err := e.jwt.GenerateToken(userID, pkgmodel.DoctorRole, jwt.AccessToken)
		require.NoError(t, err)
		e.blacklist.EXPECT().IsAccessTokenBlacklisted(gomock.Any(), claims.ID).Return(false, nil)

		resp, err := e.server.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{Token: token})
		require.NoError(t, err)
		assert.Equal(t, userID, resp.UserId)
		assert.Equal(t, string(pkgmodel.DoctorRole), resp.Role)
	})

	t.Run("revoked token", func(t *testing.T) {
		e := newEnv(t)
		token, claims, err := e.jwt.GenerateToken(userID, pkgmodel.DoctorRole, jwt.AccessToken)
		require.NoError(t, err)
		e.blacklist.EXPECT().IsAccessTokenBlacklisted(gomock.Any(), claims.ID).Return(true, nil)

		_, err = e.server.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{Token: token})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("expired token", func(t *testing.T) {
		e := newEnv(t)
		token, _, err := testutil.ExpiredJWT(t).GenerateToken(userID, pkgmodel.DoctorRole, jwt.AccessToken)
		require.NoError(t, err)

		_, err = e.server.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{Token: token})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("refresh token", func(t *testing.T) {
		e := newEnv(t)
		token, _, err := e.jwt.GenerateToken(userID, pkgmodel.DoctorRole, jwt.RefreshToken)
		require.NoError(t, err)

		_, err = e.server.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{Token: token})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("redis failure is internal", func(t *testing.T) {
		e := newEnv(t)
		token, _, err := e.jwt.GenerateToken(userID, pkgmodel.DoctorRole, jwt.AccessToken)
		require.NoError(t, err)
		e.blacklist.EXPECT().IsAccessTokenBlacklisted(gomock.Any(), gomock.Any()).Return(false, context.DeadlineExceeded)

		_, err = e.server.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{Token: token})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
