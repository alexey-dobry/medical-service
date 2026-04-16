package model

// Credentials is a model to store users data used for authorization
type Credentials struct {
	UserID       string `validate:"required,uuid"`
	Email        string `validate:"required,email"`
	PasswordHash string `validate:"required"`
}
