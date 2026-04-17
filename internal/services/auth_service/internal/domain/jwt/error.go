package jwt

import "errors"

var (
	ErrIncorrectJWTSecret  = errors.New("incorrect jwt secret")
	ErrFailedToGenerateJWT = errors.New("failed to generate jwt")
	ErrJWTTokenExpired     = errors.New("jwt token expired")
	ErrSignatureInvalid    = errors.New("invalid token signature")
)
