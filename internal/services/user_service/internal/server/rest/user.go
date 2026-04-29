package rest

import (
	"fmt"
	"strings"
	"time"

	pkgModel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (s *RESTServer) handleCreatePatientProfile() fiber.Handler {
	type registerDTO struct {
		FirstName  string `json:"name"`
		MiddleName string `json:"surname"`
		LastName   string `json:"patronymic"`
		Email      string `json:"email"`
		Phone      string `json:"phone_number"`
		Sex        string `json:"sex"`
		Password   string `json:"password"`
		BirthDate  string `json:"birth_date"`
	}

	return func(c *fiber.Ctx) error {
		var req registerDTO

		err := c.BodyParser(&req)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		birthDate, _ := time.Parse("2006-01-02", req.BirthDate)

		user := model.User{
			FirstName:  req.FirstName,
			MiddleName: req.MiddleName,
			LastName:   req.LastName,
			Email:      req.Email,
			Phone:      req.Phone,
			Sex:        req.Sex,
			BirthDate:  birthDate,
		}

		err = user.Validate()
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		userID := uuid.New()

		authRequest := client.RegisterRequest{
			UserID:   userID.String(),
			Email:    req.Email,
			Role:     pkgModel.PatientRole,
			Password: req.Password,
		}

		resp, err := s.grpcClient.RegisterPatient(authRequest)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = s.store.User().AddPatient(user)
		if err != nil {
			for range 5 {
				err := s.grpcClient.DeleteUser(client.DeleteRequest{
					UserID: userID.String(),
				})
				if err == nil {
					break
				}

				time.Sleep(time.Second * 1)
			}
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "refreshToken",
			Value:    resp.RefreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/auth/refresh",
		})

		c.Cookie(&fiber.Cookie{
			Name:     "accessToken",
			Value:    resp.AccessToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/",
		})

		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *RESTServer) handleGetPatientProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Locals("access_token").(*jwt.Token)
		claims := accessToken.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		id, _ := uuid.Parse(userID)

		patient, err := s.store.User().GetPatient(id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		photoMeta, err := s.store.Meta().GetByID(id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		profilePictureURL := fmt.Sprintf("https://photo_storage:8000/files/%s", photoMeta.ID)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":                patient.FirstName,
			"surname":             patient.MiddleName,
			"patronymic":          patient.LastName,
			"email":               patient.Email,
			"phone_number":        patient.Phone,
			"sex":                 patient.Sex,
			"profile_picture_url": profilePictureURL,
		})
	}
}

func (s *RESTServer) handleGetDoctorProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Locals("access_token").(*jwt.Token)
		claims := accessToken.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		id, _ := uuid.Parse(userID)

		doctor, doctorAdditionalData, err := s.store.User().GetDoctor(id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		photoMeta, err := s.store.Meta().GetByID(id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		profilePictureURL := fmt.Sprintf("https://photo_storage:8000/files/%s", photoMeta.ID)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":                doctor.FirstName,
			"surname":             doctor.MiddleName,
			"patronymic":          doctor.LastName,
			"email":               doctor.Email,
			"phone_number":        doctor.Phone,
			"sex":                 doctor.Sex,
			"specialty":           doctorAdditionalData.Specialty,
			"profile_picture_url": profilePictureURL,
		})
	}
}

func (s *RESTServer) handleUpdateProfile() fiber.Handler {

	var updateDTO map[string]interface{}

	return func(c *fiber.Ctx) error {
		accessToken := c.Locals("access_token").(*jwt.Token)
		claims := accessToken.Claims.(jwt.MapClaims)
		userID, _ := uuid.Parse(claims["user_id"].(string))

		err := c.BodyParser(&updateDTO)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = s.store.User().UpdateUser(userID, updateDTO)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *RESTServer) handleSearchDoctors() fiber.Handler {
	type doctorResponse struct {
		Name           string `json:"name"`
		Surname        string `json:"surname"`
		Patronymic     string `json:"patronymic"`
		WorkExperience string `json:"work_expirence"`
		Photo          string `json:"photo"`
		UUID           string `json:"uuid"`
	}

	type response struct {
		Doctors []doctorResponse `json:"doctors"`
	}

	return func(c *fiber.Ctx) error {
		searchParams := model.DoctorSearchParams{
			ID:         c.Query("id"),
			FirstName:  c.Query("first_name"),
			MiddleName: c.Query("middle_name"),
			LastName:   c.Query("last_name"),
			Sex:        c.Query("sex"),
			Specialty:  c.Query("specialty"),
		}

		if services := c.Query("services"); services != "" {
			searchParams.Services = strings.Split(services, ",")
		}

		doctorIDs, err := s.store.Search().SearchDoctor(searchParams)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to search doctors",
			})
		}

		doctors := make([]doctorResponse, 0, len(doctorIDs))
		for _, id := range doctorIDs {
			user, additional, err := s.store.User().GetDoctor(id)
			if err != nil {
				continue
			}

			doctors = append(doctors, doctorResponse{
				Name:           user.FirstName,
				Surname:        user.LastName,
				Patronymic:     user.MiddleName,
				WorkExperience: fmt.Sprintf("%d years", additional.WorkExperience),
				Photo:          user.PhotoID,
				UUID:           id.String(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(response{
			Doctors: doctors,
		})
	}
}

func (s *RESTServer) handleGetDoctorDetails() fiber.Handler {
	return func(c *fiber.Ctx) error {
		doctorIDParam := c.Params("doctor_id")
		doctorID, err := uuid.Parse(doctorIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid doctor_id",
			})
		}

		user, doctorData, err := s.store.User().GetDoctor(doctorID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get doctor details",
			})
		}

		photoURL, err := s.store.Photos().Get(user.PhotoID)
		if err != nil {
			photoURL = ""
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"doctor": fiber.Map{
				"name":           user.FirstName,
				"surname":        user.LastName,
				"patronymic":     user.MiddleName,
				"specialty":      doctorData.Specialty,
				"work_expirence": fmt.Sprintf("%d years", doctorData.WorkExperience),
				"description":    doctorData.Description,
				"photo":          photoURL,
			},
		})
	}
}
