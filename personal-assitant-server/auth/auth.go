// auth.go

package auth

import (
	"crypto/rand"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func GenerateJWTToken(username string) (string, error) {
	key, err := generateRandomKey(32)
	if err != nil {
		return "", err
	}

	// Создание токена с настройками и подпись
	claims := jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Например, токен действителен 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateRandomKey(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
