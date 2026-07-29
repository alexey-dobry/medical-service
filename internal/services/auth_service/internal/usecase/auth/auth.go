package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/entity"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordDoesNotMatch = errors.New("password does not match")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

type UseCase struct {
	credentialsRepo repository.CredentialsRepository
	blacklistRepo   repository.BlacklistRepository
	jwt             *jwt.Handler
}

func New(
	c repository.CredentialsRepository,
	b repository.BlacklistRepository,
	j *jwt.Handler,
) *UseCase {
	return &UseCase{
		credentialsRepo: c,
		blacklistRepo:   b,
		jwt:             j,
	}
}

func (uc *UseCase) Register(ctx context.Context, email, password string, role entity.Role) (string, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("auth useCase - Register - failed to hash password: %w", err)
	}

	now := time.Now().UTC()

	userId := uuid.New().String()

	credentials := entity.Credentials{
		ID:           uuid.New().String(),
		UserID:       userId,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = uc.credentialsRepo.Add(ctx, credentials)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := uc.jwt.GenerateToken(userId, string(role), jwt.RefreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := uc.jwt.GenerateToken(userId, string(role), jwt.AccessToken)
	if err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	creds, err := uc.credentialsRepo.GetOneByMail(ctx, email)
	if err != nil {
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password))
	if err != nil {
		return "", "", ErrPasswordDoesNotMatch
	}

	refreshToken, err := uc.jwt.GenerateToken(creds.UserID, string(creds.Role), jwt.RefreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := uc.jwt.GenerateToken(creds.UserID, string(creds.Role), jwt.AccessToken)
	if err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

func (uc *UseCase) Logout(ctx context.Context, refreshClaims, accessClaims jwt.Claims) error {
	refreshExpiresIn := refreshClaims.ExpiresAt.Sub(time.Now())
	accessExpiresIn := accessClaims.ExpiresAt.Sub(time.Now())

	return uc.blacklistRepo.Logout(ctx, refreshClaims.ID, accessClaims.ID, refreshExpiresIn, accessExpiresIn)
}

func (uc *UseCase) Refresh(ctx context.Context, refreshClaims, accessClaims jwt.Claims) (string, error) {
	if !refreshClaims.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("auth usecase - Refresh() - %w", ErrRefreshTokenExpired)
	}

	err := uc.blacklistRepo.BlacklistToken(ctx, accessClaims.ID, accessClaims.ExpiresAt.Sub(time.Now()))
	if err != nil {
		return "", err
	}

	accessToken, err := uc.jwt.GenerateToken(refreshClaims.Subject, string(refreshClaims.Role), jwt.AccessToken)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (uc *UseCase) Delete(ctx context.Context, userID string) error {
	return uc.credentialsRepo.Delete(ctx, userID)
}
