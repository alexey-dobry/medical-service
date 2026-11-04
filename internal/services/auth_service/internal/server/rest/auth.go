package rest

import (
	"errors"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/gofiber/fiber/v2"
)

func (s *RESTServer) handleLogin() fiber.Handler {
	type loginDTO struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	return func(c *fiber.Ctx) error {
		var req loginDTO

		err := c.BodyParser(&req)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = validator.V.Struct(req)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// Несуществующий email и неверный пароль дают одинаковый ответ и
		// примерно одинаковую задержку, чтобы нельзя было перебрать почты.
		credentials, err := s.credentialsRepo.GetOneByMail(c.UserContext(), req.Email)
		if errors.Is(err, repository.ErrNotFound) {
			utils.SimulatePasswordCheck(req.Password)
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed to get credentials: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if !utils.CheckPasswordHash(req.Password, credentials.PasswordHash) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		pair, err := s.jwtHandler.GeneratePair(credentials.UserID, credentials.Role)
		if err != nil {
			s.logger.Errorf("Failed to generate token pair: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		s.setTokenCookies(c, pair.AccessToken, pair.RefreshToken)

		return c.SendStatus(fiber.StatusOK)
	}
}

// handleLogout отзывает все предъявленные токены: jti access-токена уходит
// в чёрный список, jti refresh-токена — в завершённые сессии. Просроченный
// access-токен не считается ошибкой, но хотя бы один валидный токен нужен.
func (s *RESTServer) handleLogout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		revoked := false

		if accessToken := c.Cookies("accessToken"); accessToken != "" {
			accessClaims, err := s.jwtHandler.ParseToken(accessToken, jwt.AccessToken)
			if err == nil {
				err = s.blacklistRepo.BlacklistAccessToken(ctx, accessClaims.ID, time.Until(accessClaims.ExpiresAt.Time))
				if err != nil {
					s.logger.Errorf("Failed to blacklist access token: %s", err)
					return c.SendStatus(fiber.StatusInternalServerError)
				}
				revoked = true
			} else if !isTokenError(err) {
				s.logger.Errorf("Failed validate access token: %s", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		if refreshToken := c.Cookies("refreshToken"); refreshToken != "" {
			refreshClaims, err := s.jwtHandler.ParseToken(refreshToken, jwt.RefreshToken)
			if err == nil {
				err = s.blacklistRepo.StoreLogoutSession(ctx, refreshClaims.ID, time.Until(refreshClaims.ExpiresAt.Time))
				if err != nil {
					s.logger.Errorf("Failed to store logout session: %s", err)
					return c.SendStatus(fiber.StatusInternalServerError)
				}
				revoked = true
			} else if !isTokenError(err) {
				s.logger.Errorf("Failed validate refresh token: %s", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		if !revoked {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		clearTokenCookies(c)

		return c.SendStatus(fiber.StatusOK)
	}
}

// handleRefresh работает только с refresh-токеном: отсутствующий или
// просроченный access-токен — обычная причина для вызова. Refresh-токен
// ротируется: его jti сохраняется как завершённая сессия, а клиенту
// выдаётся новая пара.
func (s *RESTServer) handleRefresh() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		refreshToken := c.Cookies("refreshToken")
		if refreshToken == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		refreshClaims, err := s.jwtHandler.ParseToken(refreshToken, jwt.RefreshToken)
		if isTokenError(err) {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if err != nil {
			s.logger.Errorf("Failed validate refresh token: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		loggedOut, err := s.blacklistRepo.IsSessionLoggedOut(ctx, refreshClaims.ID)
		if err != nil {
			s.logger.Errorf("Failed to check refresh session: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if loggedOut {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if err = s.blacklistRepo.StoreLogoutSession(ctx, refreshClaims.ID, time.Until(refreshClaims.ExpiresAt.Time)); err != nil {
			s.logger.Errorf("Failed to close rotated refresh session: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Старый access-токен отзывается, если ещё валиден; отсутствие или
		// просрочка — не ошибка.
		if accessToken := c.Cookies("accessToken"); accessToken != "" {
			if accessClaims, err := s.jwtHandler.ParseToken(accessToken, jwt.AccessToken); err == nil {
				err = s.blacklistRepo.BlacklistAccessToken(ctx, accessClaims.ID, time.Until(accessClaims.ExpiresAt.Time))
				if err != nil {
					s.logger.Errorf("Failed to blacklist access token: %s", err)
					return c.SendStatus(fiber.StatusInternalServerError)
				}
			}
		}

		pair, err := s.jwtHandler.GeneratePair(refreshClaims.UserID(), refreshClaims.Role)
		if err != nil {
			s.logger.Errorf("Failed to generate token pair: %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		s.setTokenCookies(c, pair.AccessToken, pair.RefreshToken)

		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *RESTServer) setTokenCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   s.secureCookies,
		SameSite: "Strict",
		Path:     "/auth",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   s.secureCookies,
		SameSite: "Strict",
		Path:     "/",
	})
}

func clearTokenCookies(c *fiber.Ctx) {
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
		Path:   "/auth",
	})
}

// isTokenError показывает, что клиент прислал плохой токен (просроченный,
// битый, с неверной подписью), а не внутреннюю ошибку.
func isTokenError(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenInvalid)
}
