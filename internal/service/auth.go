package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthSvc отвечает за создание и проверку JWT
type AuthSvc struct {
	secretKey []byte // просто пока что так оставил
	ttl       time.Duration
}

func NewAuthSvc(secretKey string, ttl time.Duration) *AuthSvc {
	return &AuthSvc{
		secretKey: []byte(secretKey),
		ttl:       ttl,
	}
}

func (s *AuthSvc) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                       // Subject (кому выдан токен)
		"exp": time.Now().Add(s.ttl).Unix(), // Expires at (когда истекает)
		"iat": time.Now().Unix(),            // Issued at (когда выдан)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthSvc) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Если токен валиден, извлекаем из него claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDFloat, ok := claims["sub"].(float64); ok {
			return int(userIDFloat), nil
		}
	}

	return 0, fmt.Errorf("invalid token claims")
}
