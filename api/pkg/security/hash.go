package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytePass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error in hashing pass: %w", err)
	}

	return string(bytePass), nil
}

func ComparePassword(hashPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password)); err != nil {
		return fmt.Errorf("error in comparing password: %w", err)
	}
	return nil
}

func GenerateRefreshToken() (string, error) {
	bytesToken := make([]byte, 32)
	if _, err := rand.Read(bytesToken); err != nil {
		return "", fmt.Errorf("error in generating refresh tokens")
	}

	return hex.EncodeToString(bytesToken), nil
}

func HashRefreshToken(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashedToken[:])
}
