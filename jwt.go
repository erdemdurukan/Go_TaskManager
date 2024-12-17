package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your_secret_key")

func GenerateToken(Id int) (string, error) {
	claims := jwt.MapClaims{
		"Id":  Id,
		"exp": time.Now().Add(time.Hour * 2).Unix(), // 2 hours
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)

}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return &claims, nil
	}
	return nil, errors.New("invalid token")
}

// ExtractClaims extracts claims from the token
func ExtractClaims(claims *jwt.MapClaims) (int, error) {
	Id, ok := (*claims)["Id"].(float64) // JSON'dan gelen int, float64 olarak döner
	if !ok {
		return 0, errors.New("userID not found in token claims")
	}
	return int(Id), nil
}
