package jwt

import (
	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwtlib.RegisteredClaims
	Role string `json:"role"`
}

func (c Claims) GetRole() (string, error) {
	return c.Role, nil
}
