package jwt

import (
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	ID   string     `json:"id"`
	Role model.Role `json:"role"`
}
