package grpc

import (
	"context"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
)

func (c *grpcClient) RegisterPatient(ctx context.Context, request client.RegisterRequest) (client.RegisterPatientResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	grpcRequest := pb.RegisterRequest{
		UserId:   request.UserID,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
	}

	result, err := c.Auth.RegisterPatient(ctx, &grpcRequest)
	if err != nil {
		return client.RegisterPatientResponse{}, err
	}

	return client.RegisterPatientResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (c *grpcClient) RegisterDoctor(ctx context.Context, request client.RegisterRequest) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	grpcRequest := pb.RegisterRequest{
		UserId:   request.UserID,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
	}

	_, err := c.Auth.RegisterDoctor(ctx, &grpcRequest)
	return err
}

func (c *grpcClient) DeleteUser(ctx context.Context, request client.DeleteRequest) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	grpcRequest := pb.DeleteUserRequest{
		UserId: request.UserID,
	}

	_, err := c.Auth.DeleteUser(ctx, &grpcRequest)
	return err
}

func (c *grpcClient) ValidateAccessToken(ctx context.Context, token string) (client.TokenInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.Auth.ValidateAccessToken(ctx, &pb.ValidateAccessTokenRequest{Token: token})
	if err != nil {
		return client.TokenInfo{}, err
	}

	return client.TokenInfo{UserID: resp.UserId, Role: resp.Role}, nil
}
