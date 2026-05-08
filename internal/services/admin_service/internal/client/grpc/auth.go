package grpc

import (
	"context"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/user"
	"github.com/alexey-dobry/medical-service/internal/services/admin_service/internal/client"
	"github.com/google/uuid"
)

func (c *grpcClient) CreateDoctor(request client.CreateDoctorRequest) error {
	ctx := context.Background()

	profilePicture := pb.ProfilePicture{
		Filename: request.ProfilePicture.Filename,
		Content:  request.ProfilePicture.Content,
		MimeType: request.ProfilePicture.MimeType,
	}

	req := pb.CreateDoctorRequest{
		FirstName:      request.FirstName,
		MiddleName:     request.MiddleName,
		LastName:       request.LastName,
		Phone:          request.Phone,
		Email:          request.Email,
		Sex:            request.Sex,
		BirthDate:      request.BirthDate,
		Specialty:      request.Specialty,
		WorkExperience: request.WorkExperience,
		Description:    request.Description,
		Services:       request.Services,
		ProfilePicture: &profilePicture,
	}

	_, err := c.User.CreateDoctor(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}

func (c *grpcClient) DeleteDoctor(id uuid.UUID) error {
	ctx := context.Background()

	req := pb.DeleteDoctorRequest{
		UserId: id.String(),
	}

	_, err := c.User.DeleteDoctor(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}
