package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/service"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const minPasswordLength = 8

func (s *RESTServer) handleCreatePatientProfile() fiber.Handler {
	type registerDTO struct {
		FirstName  string `json:"name"`
		LastName   string `json:"surname"`
		MiddleName string `json:"patronymic"`
		Email      string `json:"email"`
		Phone      string `json:"phone_number"`
		Sex        string `json:"sex"`
		Password   string `json:"password"`
		BirthDate  string `json:"birth_date"`
	}

	return func(c *fiber.Ctx) error {
		var req registerDTO
		var photo *service.PhotoUpload

		// Принимаем два формата: обычный JSON или multipart/form-data,
		// где JSON лежит в поле "data", а файл — в необязательном "photo".
		if form, err := c.MultipartForm(); err == nil {
			data := form.Value["data"]
			if len(data) != 1 || json.Unmarshal([]byte(data[0]), &req) != nil {
				return errorJSON(c, fiber.StatusBadRequest, "'data' form field must contain valid JSON")
			}

			if files := form.File["photo"]; len(files) > 0 {
				file, err := files[0].Open()
				if err != nil {
					return errorJSON(c, fiber.StatusBadRequest, "failed to read 'photo'")
				}
				defer file.Close()

				photo = &service.PhotoUpload{
					Name:     files[0].Filename,
					MimeType: files[0].Header.Get("Content-Type"),
					Size:     files[0].Size,
					Content:  file,
				}
			}
		} else if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
		}

		if len(req.Password) < minPasswordLength {
			return errorJSON(c, fiber.StatusBadRequest, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
		}

		birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "birth_date must be in YYYY-MM-DD format")
		}

		userID := uuid.New()

		user := model.User{
			ID:         userID.String(),
			FirstName:  req.FirstName,
			MiddleName: req.MiddleName,
			LastName:   req.LastName,
			Email:      req.Email,
			Phone:      req.Phone,
			Sex:        req.Sex,
			BirthDate:  birthDate,
		}

		if err = user.Validate(); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		}

		resp, err := s.registration.RegisterPatient(c.UserContext(), service.PatientRegistration{
			User:     user,
			Password: req.Password,
			Photo:    photo,
		})
		if errors.Is(err, service.ErrInvalidPhoto) {
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		} else if errors.Is(err, service.ErrAlreadyExists) {
			return errorJSON(c, fiber.StatusConflict, "user with this email already exists")
		} else if err != nil {
			code := grpcErrorToHTTP(err)
			if code == fiber.StatusInternalServerError {
				s.logger.Errorf("Failed to register patient: %s", err)
			}
			return c.SendStatus(code)
		}

		s.setAuthCookies(c, resp.AccessToken, resp.RefreshToken)

		return c.SendStatus(fiber.StatusCreated)
	}
}

func (s *RESTServer) handleGetPatientProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identity, err := middleware.IdentityFrom(c)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		patient, err := s.store.User().GetPatient(c.UserContext(), identity.UserID)
		if errors.Is(err, store.ErrNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "profile not found")
		} else if err != nil {
			s.logger.Errorf("Failed to get patient: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		profilePictureURL, err := s.photoURL(c.UserContext(), patient.PhotoID)
		if err != nil {
			s.logger.Errorf("Failed to get profile picture: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":                patient.FirstName,
			"surname":             patient.LastName,
			"patronymic":          patient.MiddleName,
			"email":               patient.Email,
			"phone_number":        patient.Phone,
			"sex":                 patient.Sex,
			"birth_date":          patient.BirthDate.Format(time.DateOnly),
			"profile_picture_url": nullIfEmpty(profilePictureURL),
		})
	}
}

func (s *RESTServer) handleGetDoctorProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identity, err := middleware.IdentityFrom(c)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		doctor, doctorAdditionalData, err := s.store.User().GetDoctor(c.UserContext(), identity.UserID)
		if errors.Is(err, store.ErrNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "profile not found")
		} else if err != nil {
			s.logger.Errorf("Failed to get doctor: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		profilePictureURL, err := s.photoURL(c.UserContext(), doctor.PhotoID)
		if err != nil {
			s.logger.Errorf("Failed to get profile picture: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":                doctor.FirstName,
			"surname":             doctor.LastName,
			"patronymic":          doctor.MiddleName,
			"email":               doctor.Email,
			"phone_number":        doctor.Phone,
			"sex":                 doctor.Sex,
			"specialty":           doctorAdditionalData.Specialty,
			"profile_picture_url": nullIfEmpty(profilePictureURL),
		})
	}
}

func (s *RESTServer) handleUpdateProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identity, err := middleware.IdentityFrom(c)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// Объявляем на каждый запрос: map, захваченный замыканием хендлера,
		// был бы общим для всех конкурентных запросов и привёл бы к гонке.
		var updateDTO map[string]interface{}
		if err = c.BodyParser(&updateDTO); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
		}

		err = s.store.User().UpdateUser(c.UserContext(), identity.UserID, updateDTO)
		switch {
		case errors.Is(err, store.ErrInvalidField):
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, store.ErrNotFound):
			return errorJSON(c, fiber.StatusNotFound, "profile not found")
		case err != nil:
			s.logger.Errorf("Failed to update profile: %s", err)
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
		Specialty      string `json:"specialty"`
		WorkExperience string `json:"work_experience"`
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

		doctorIDs, err := s.store.Search().SearchDoctor(c.UserContext(), searchParams)
		if err != nil {
			s.logger.Errorf("Failed to search doctors: %s", err)
			return errorJSON(c, fiber.StatusInternalServerError, "failed to search doctors")
		}

		doctors := make([]doctorResponse, 0, len(doctorIDs))
		for _, id := range doctorIDs {
			user, additional, err := s.store.User().GetDoctor(c.UserContext(), id)
			if errors.Is(err, store.ErrNotFound) {
				// Поисковый индекс догоняет Postgres с задержкой.
				s.logger.Warnf("Doctor %s is indexed but missing in database", id)
				continue
			} else if err != nil {
				s.logger.Errorf("Failed to get doctor %s: %s", id, err)
				return errorJSON(c, fiber.StatusInternalServerError, "failed to search doctors")
			}

			photoURL, err := s.photoURL(c.UserContext(), user.PhotoID)
			if err != nil {
				// Отсутствие фото не должно скрывать врача из выдачи.
				s.logger.Errorf("Failed to get photo of doctor %s: %s", id, err)
			}

			doctors = append(doctors, doctorResponse{
				Name:           user.FirstName,
				Surname:        user.LastName,
				Patronymic:     user.MiddleName,
				Specialty:      additional.Specialty,
				WorkExperience: fmt.Sprintf("%d years", additional.WorkExperience),
				Photo:          photoURL,
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
		doctorID, err := uuid.Parse(c.Params("doctor_id"))
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "invalid doctor_id")
		}

		user, doctorData, err := s.store.User().GetDoctor(c.UserContext(), doctorID)
		if errors.Is(err, store.ErrNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "doctor not found")
		} else if err != nil {
			s.logger.Errorf("Failed to get doctor details: %s", err)
			return errorJSON(c, fiber.StatusInternalServerError, "failed to get doctor details")
		}

		photoURL, err := s.photoURL(c.UserContext(), user.PhotoID)
		if err != nil {
			s.logger.Errorf("Failed to get photo of doctor %s: %s", doctorID, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"doctor": fiber.Map{
				"name":            user.FirstName,
				"surname":         user.LastName,
				"patronymic":      user.MiddleName,
				"specialty":       doctorData.Specialty,
				"work_experience": fmt.Sprintf("%d years", doctorData.WorkExperience),
				"description":     doctorData.Description,
				"services":        doctorData.Services,
				"photo":           nullIfEmpty(photoURL),
			},
		})
	}
}

func (s *RESTServer) handleCreateDoctor() fiber.Handler {
	type createDoctorRequest struct {
		FirstName      string   `json:"first_name"`
		MiddleName     string   `json:"middle_name"`
		LastName       string   `json:"last_name"`
		Phone          string   `json:"phone"`
		Email          string   `json:"email"`
		Password       string   `json:"password"`
		Sex            string   `json:"sex"`
		BirthDate      string   `json:"birth_date"`
		Specialty      string   `json:"specialty"`
		WorkExperience string   `json:"work_experience"`
		Description    string   `json:"description"`
		Services       []string `json:"services"`
	}

	return func(c *fiber.Ctx) error {
		dataField := c.FormValue("data")
		if dataField == "" {
			return errorJSON(c, fiber.StatusBadRequest, "missing 'data' form field")
		}

		var req createDoctorRequest
		if err := json.Unmarshal([]byte(dataField), &req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "'data' must be valid JSON")
		}

		if len(req.Password) < minPasswordLength {
			return errorJSON(c, fiber.StatusBadRequest, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
		}

		birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "birth_date must be in YYYY-MM-DD format")
		}

		workExperience, err := strconv.Atoi(req.WorkExperience)
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "work_experience must be an integer")
		}

		fileHeader, err := c.FormFile("photo")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "missing 'photo' form file")
		}

		file, err := fileHeader.Open()
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "failed to read 'photo'")
		}
		defer file.Close()

		userID := uuid.New()
		user := model.User{
			ID:         userID.String(),
			FirstName:  req.FirstName,
			MiddleName: req.MiddleName,
			LastName:   req.LastName,
			Phone:      req.Phone,
			Email:      req.Email,
			Sex:        req.Sex,
			BirthDate:  birthDate,
		}

		doctorData := model.DoctorAdditionalData{
			UserID:         userID.String(),
			Specialty:      req.Specialty,
			WorkExperience: workExperience,
			Description:    req.Description,
			Services:       req.Services,
		}

		if err := user.Validate(); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		}
		if err := doctorData.Validate(); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		}

		err = s.registration.RegisterDoctor(c.UserContext(), service.DoctorRegistration{
			User:     user,
			Doctor:   doctorData,
			Password: req.Password,
			Photo: &service.PhotoUpload{
				Name:     fileHeader.Filename,
				MimeType: fileHeader.Header.Get("Content-Type"),
				Size:     fileHeader.Size,
				Content:  file,
			},
		})
		switch {
		case errors.Is(err, service.ErrInvalidPhoto):
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrAlreadyExists):
			return errorJSON(c, fiber.StatusConflict, "user with this email already exists")
		case err != nil:
			code := grpcErrorToHTTP(err)
			if code == fiber.StatusInternalServerError {
				s.logger.Errorf("Failed to create doctor: %s", err)
			}
			return c.SendStatus(code)
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"uuid": userID.String()})
	}
}

func (s *RESTServer) handleDeleteDoctor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		doctorID, err := uuid.Parse(c.Params("doctor_id"))
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "invalid doctor_id")
		}

		if _, _, err := s.store.User().GetDoctor(ctx, doctorID); errors.Is(err, store.ErrNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "doctor not found")
		} else if err != nil {
			s.logger.Errorf("Failed to get doctor: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Сначала удаляем учётные данные, чтобы врач не смог залогиниться,
		// даже если следующий шаг упадёт. Все шаги идемпотентны — админ может повторить.
		err = s.grpcClient.DeleteUser(ctx, client.DeleteRequest{UserID: doctorID.String()})
		if err != nil {
			s.logger.Errorf("Failed to delete user credentials: %s", err)
			return c.SendStatus(grpcErrorToHTTP(err))
		}

		photoStorageKey, err := s.store.Meta().DeleteByUserID(ctx, doctorID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			s.logger.Errorf("Failed to delete photo metadata: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err = s.store.User().DeleteUser(ctx, doctorID); err != nil {
			s.logger.Errorf("Failed to delete doctor profile: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err = s.store.Search().DeleteDoctor(ctx, doctorID); err != nil && !errors.Is(err, store.ErrNotFound) {
			s.logger.Errorf("Failed to delete doctor %s from search index: %s", doctorID, err)
		}

		if photoStorageKey != "" {
			if err = s.store.Photos().Delete(photoStorageKey); err != nil {
				s.logger.Errorf("Failed to delete photo %s: %s", photoStorageKey, err)
			}
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
