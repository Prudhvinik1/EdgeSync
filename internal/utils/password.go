package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost     = 12
	PasswordLength = 12
)

func HashPassword(password string) (string, error) {
	if len(password) < PasswordLength {
		return "", fmt.Errorf("password must be at least %d characters long", PasswordLength)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
