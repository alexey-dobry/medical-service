package app

import (
	"context"
	"errors"
	"fmt"

	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/config"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/google/uuid"
)

const minAdminPasswordLength = 8

// ensureAdmin создаёт администратора, если его ещё нет. Через публичный API
// админов не зарегистрировать, поэтому без этого никто не создаст врачей.
// Идемпотентна: безопасно вызывать при каждом старте и из нескольких реплик.
func ensureAdmin(ctx context.Context, repo repository.CredentialsRepository, cfg config.AdminConfig) (created bool, err error) {
	if cfg.Email == "" {
		return false, nil
	}
	if len(cfg.Password) < minAdminPasswordLength {
		return false, fmt.Errorf("ADMIN_PASSWORD must be at least %d characters", minAdminPasswordLength)
	}

	existing, err := repo.GetOneByMail(ctx, cfg.Email)
	switch {
	case err == nil:
		if existing.Role != pkgmodel.AdminRole {
			return false, fmt.Errorf("account %s exists with role %s, not %s", cfg.Email, existing.Role, pkgmodel.AdminRole)
		}
		return false, nil
	case !errors.Is(err, repository.ErrNotFound):
		return false, fmt.Errorf("failed to look up admin: %w", err)
	}

	hash, err := utils.HashPassword(cfg.Password)
	if err != nil {
		return false, fmt.Errorf("failed to hash admin password: %w", err)
	}

	credentials := model.Credentials{
		UserID:       uuid.NewString(),
		Email:        cfg.Email,
		Role:         pkgmodel.AdminRole,
		PasswordHash: hash,
	}
	if err := credentials.Validate(); err != nil {
		return false, fmt.Errorf("invalid admin credentials: %w", err)
	}

	err = repo.Add(ctx, credentials)
	if errors.Is(err, repository.ErrAlreadyExists) {
		// Админа успела создать другая реплика.
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to create admin: %w", err)
	}

	return true, nil
}
