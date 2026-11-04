package auth

import (
	"context"
	"errors"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) RegisterPatient(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	credentials, err := s.registerUser(ctx, req, pkgmodel.PatientRole)
	if err != nil {
		return nil, err
	}

	pair, err := s.jwtHandler.GeneratePair(credentials.UserID, credentials.Role)
	if err != nil {
		s.logger.Errorf("Failed to generate token pair: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	response := pb.RegisterResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}

	return &response, nil
}

func (s *GRPCServer) RegisterDoctor(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	if _, err := s.registerUser(ctx, req, pkgmodel.DoctorRole); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// registerUser проверяет роль из запроса, хеширует пароль и сохраняет учётные данные.
func (s *GRPCServer) registerUser(ctx context.Context, req *pb.RegisterRequest, expected pkgmodel.Role) (model.Credentials, error) {
	role, ok := pkgmodel.ParseRole(req.Role)
	if !ok || role != expected {
		return model.Credentials{}, status.Errorf(codes.InvalidArgument, "Invalid role %q, expected %q", req.Role, expected)
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorf("Failed to hash password: %s", err)
		return model.Credentials{}, status.Error(codes.Internal, "Internal server error")
	}

	credentials := model.Credentials{
		UserID:       req.UserId,
		Email:        req.Email,
		Role:         role,
		PasswordHash: passwordHash,
	}

	if err = credentials.Validate(); err != nil {
		return model.Credentials{}, status.Error(codes.InvalidArgument, "Invalid user field value")
	}

	err = s.credentialsRepo.Add(ctx, credentials)
	if errors.Is(err, repository.ErrAlreadyExists) {
		return model.Credentials{}, status.Error(codes.AlreadyExists, "Account with specified email already exists")
	} else if err != nil {
		s.logger.Errorf("Failed to add new user to data: %s", err)
		return model.Credentials{}, status.Error(codes.Internal, "Internal server error")
	}

	return credentials, nil
}

func (s *GRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user id")
	}

	err = s.credentialsRepo.Delete(ctx, userID)
	if err != nil {
		s.logger.Errorf("Failed to delete user: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &emptypb.Empty{}, nil
}
