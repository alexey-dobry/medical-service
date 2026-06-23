package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func (c Claims) GetRole() (string, error) {
	return c.Role, nil
}
