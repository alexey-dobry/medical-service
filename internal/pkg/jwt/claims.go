package jwt

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims описывает клеймы каждого токена, выпускаемого auth_service:
//   - sub: идентификатор пользователя
//   - jti: уникальный идентификатор токена, используется для отзыва
//   - exp, iat: время жизни
//   - role: роль пользователя
//   - typ: ACCESS или REFRESH, чтобы один тип нельзя было использовать
//     вместо другого, даже если секреты случайно совпадут.
type Claims struct {
	jwtlib.RegisteredClaims
	Role model.Role `json:"role"`
	Type TokenType  `json:"typ"`
}

// UserID возвращает идентификатор пользователя, хранящийся в "sub".
func (c Claims) UserID() string {
	return c.Subject
}
