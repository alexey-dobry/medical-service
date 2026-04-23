package grpc

import (
	"context"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
)

func (c *grpcClient) RegisterPatient(request client.RegisterRequest) (client.RegisterPatientResponse, error) {
	ctx := context.Background()

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

func (c *grpcClient) RegisterDoctor(request client.RegisterRequest) error {
	ctx := context.Background()

	grpcRequest := pb.RegisterRequest{
		UserId:   request.UserID,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
	}

	_, err := c.Auth.RegisterDoctor(ctx, &grpcRequest)
	if err != nil {
		return err
	}

	return nil
}

func (c *grpcClient) DeleteUser(request client.DeleteRequest) error {
	ctx := context.Background()

	grpcRequest := pb.DeleteUserRequest{
		UserId: request.UserID,
	}

	_, err := c.Auth.DeleteUser(ctx, &grpcRequest)
	if err != nil {
		return err
	}

	return nil
}
