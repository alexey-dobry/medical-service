package app

import (
	"context"
	"testing"

	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var adminCfg = config.AdminConfig{Email: "admin@example.com", Password: "admin-password"}

func TestEnsureAdmin_CreatesMissingAdmin(t *testing.T) {
	repo := mocks.NewMockCredentialsRepository(gomock.NewController(t))
	repo.EXPECT().GetOneByMail(gomock.Any(), adminCfg.Email).Return(model.Credentials{}, repository.ErrNotFound)
	repo.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c model.Credentials) error {
		assert.Equal(t, pkgmodel.AdminRole, c.Role)
		assert.True(t, utils.CheckPasswordHash(adminCfg.Password, c.PasswordHash))
		return nil
	})

	created, err := ensureAdmin(context.Background(), repo, adminCfg)
	require.NoError(t, err)
	assert.True(t, created)
}

func TestEnsureAdmin_IsIdempotent(t *testing.T) {
	repo := mocks.NewMockCredentialsRepository(gomock.NewController(t))
	repo.EXPECT().GetOneByMail(gomock.Any(), adminCfg.Email).Return(model.Credentials{Role: pkgmodel.AdminRole}, nil)
	repo.EXPECT().Add(gomock.Any(), gomock.Any()).Times(0)

	created, err := ensureAdmin(context.Background(), repo, adminCfg)
	require.NoError(t, err)
	assert.False(t, created)
}

func TestEnsureAdmin_RefusesToPromoteExistingUser(t *testing.T) {
	repo := mocks.NewMockCredentialsRepository(gomock.NewController(t))
	repo.EXPECT().GetOneByMail(gomock.Any(), adminCfg.Email).Return(model.Credentials{Role: pkgmodel.PatientRole}, nil)

	_, err := ensureAdmin(context.Background(), repo, adminCfg)
	assert.Error(t, err)
}

func TestEnsureAdmin_SkippedWithoutEmail(t *testing.T) {
	repo := mocks.NewMockCredentialsRepository(gomock.NewController(t))

	created, err := ensureAdmin(context.Background(), repo, config.AdminConfig{})
	require.NoError(t, err)
	assert.False(t, created)
}
