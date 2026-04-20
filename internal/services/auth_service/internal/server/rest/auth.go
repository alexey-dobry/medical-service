package rest

import (
	"errors"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/gofiber/fiber/v2"
)

// CHANGE RETURN ERRORS

func (s *RESTServer) handleLogin() fiber.Handler {
	type loginDTO struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	return func(c *fiber.Ctx) error {
		var req loginDTO

		err := c.BodyParser(&req)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = validator.V.Struct(req)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		credentials, err := s.store.Credentials().GetOneByMail(req.Email)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if !utils.CheckPasswordHash(req.Password, credentials.PasswordHash) {
			return c.SendStatus(fiber.StatusForbidden)
		}

		refreshToken, accessToken, err := s.jwtHandler.GenerateJWTPair(jwt.Claims{
			ID:   credentials.UserID,
			Role: credentials.Role,
		})

		if err != nil {
			s.logger.Errorf("Failed to generate token pair: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "refreshToken",
			Value:    refreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/auth/refresh",
		})

		c.Cookie(&fiber.Cookie{
			Name:     "accessToken",
			Value:    accessToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/",
		})

		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *RESTServer) handleLogout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		refreshToken := c.Cookies("refreshToken")
		accessToken := c.Cookies("accessToken")

		accessClaims, err := s.jwtHandler.ValidateJWT(accessToken, jwt.AccessToken)
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed validate access token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		refreshClaims, err := s.jwtHandler.ValidateJWT(refreshToken, jwt.RefreshToken)
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed validate refresh token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = s.store.Blacklist().BlacklistAccessToken(accessClaims.ID, accessClaims.ExpiresAt.Sub(time.Now()))
		if err != nil {
			s.logger.Errorf("Failed to blacklist access token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = s.store.Blacklist().StoreLogoutSession(refreshClaims.ID, refreshClaims.ExpiresAt.Sub(time.Now()))
		if err != nil {
			s.logger.Errorf("Failed to blacklist refresh token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Cookie(&fiber.Cookie{
			Name:   "accessToken",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})
		c.Cookie(&fiber.Cookie{
			Name:   "refreshToken",
			Value:  "",
			MaxAge: -1,
			Path:   "/auth/refresh",
		})

		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *RESTServer) handleRefresh() fiber.Handler {
	return func(c *fiber.Ctx) error {
		refreshToken := c.Cookies("refreshToken")
		accessToken := c.Cookies("accessToken")

		accessClaims, err := s.jwtHandler.ValidateJWT(accessToken, jwt.AccessToken)
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed validate access token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		refreshClaims, err := s.jwtHandler.ValidateJWT(refreshToken, jwt.RefreshToken)
		if errors.Is(err, jwt.ErrJWTTokenExpired) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if errors.Is(err, jwt.ErrSignatureInvalid) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed validate refresh token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = s.store.Blacklist().BlacklistAccessToken(accessClaims.ID, accessClaims.ExpiresAt.Sub(time.Now()))
		if err != nil {
			s.logger.Errorf("Failed to blacklist access token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		_, accessToken, err = s.jwtHandler.GenerateJWTPair(jwt.Claims{
			ID:   refreshClaims.ID,
			Role: refreshClaims.Role,
		})

		if err != nil {
			s.logger.Errorf("Failed to generate token pair: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "accessToken",
			Value:    accessToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/",
		})

		return c.SendStatus(fiber.StatusOK)
	}
}
