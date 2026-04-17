package jwt

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	ID   uuid.UUID  `json:"id"`
	Role model.Role `json:"role"`
}
