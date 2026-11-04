package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var dummyHash = mustHash("dummy-password-for-timing-equalization")

func SimulatePasswordCheck(password string) {
	CheckPasswordHash(password, dummyHash)
}

func mustHash(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}
