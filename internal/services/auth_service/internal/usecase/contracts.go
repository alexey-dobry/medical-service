package usecase

import (
	"context"

	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/entity"
)

type Auth interface {
	Register(ctx context.Context, email, password string, role entity.Role) (string, string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	Logout(ctx context.Context, refreshToken, accessToken string) error
	Refresh(ctx context.Context, refreshToken, accessToken string) error
	Delete(ctx context.Context, userId string) error
}
