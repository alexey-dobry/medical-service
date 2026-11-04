package middleware

import (
	"context"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceValidator проверяет токен через auth_service по gRPC.
// Секрет JWT и чёрный список в Redis живут в auth_service, поэтому
// отозванный при логауте токен здесь тоже отклоняется, а user_service
// не нужно держать их у себя.
type AuthServiceValidator struct {
	client client.Client
}

func NewAuthServiceValidator(c client.Client) *AuthServiceValidator {
	return &AuthServiceValidator{client: c}
}

func (v *AuthServiceValidator) ValidateAccessToken(ctx context.Context, token string) (Identity, error) {
	info, err := v.client.ValidateAccessToken(ctx, token)
	if status.Code(err) == codes.Unauthenticated {
		return Identity{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	} else if err != nil {
		return Identity{}, fmt.Errorf("auth service: %w", err)
	}

	userID, err := uuid.Parse(info.UserID)
	if err != nil {
		return Identity{}, fmt.Errorf("auth service returned invalid user id %q: %w", info.UserID, err)
	}

	role, ok := model.ParseRole(info.Role)
	if !ok {
		return Identity{}, fmt.Errorf("auth service returned unknown role %q", info.Role)
	}

	return Identity{UserID: userID, Role: role}, nil
}
