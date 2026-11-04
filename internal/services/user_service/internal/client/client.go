package client

//go:generate mockgen -source=client.go -destination=mocks/client_mock.go -package=mocks

import "context"

type Client interface {
	RegisterPatient(ctx context.Context, request RegisterRequest) (RegisterPatientResponse, error)
	RegisterDoctor(ctx context.Context, request RegisterRequest) error
	DeleteUser(ctx context.Context, request DeleteRequest) error
	ValidateAccessToken(ctx context.Context, token string) (TokenInfo, error)
	Close() error
}

// TokenInfo — данные пользователя из валидного access-токена.
type TokenInfo struct {
	UserID string
	Role   string
}

type RegisterRequest struct {
	UserID   string
	Email    string
	Role     string
	Password string
}

type RegisterPatientResponse struct {
	AccessToken  string
	RefreshToken string
}

type DeleteRequest struct {
	UserID string
}
