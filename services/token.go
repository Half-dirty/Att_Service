package services

import (
	"att_service/config"
	"att_service/database"
	"att_service/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var JwtSecret = []byte("your_secret_key")

// GenerateVerificationLink генерирует JWT-токен и формирует ссылку подтверждения.
func GenerateVerificationLink(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtSecret)
	if err != nil {
		return "", err
	}
	return "http://localhost:3000/verify?token=" + tokenString, nil
}

// HashPassword возвращает хеш пароля.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func GenerateTokens(userID uint, userEmail, userRole string) (string, string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"user":   userEmail,
		"role":   userRole,
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"user":   userEmail,
		"role":   userRole,
		"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	accessTokenStr, err := accessToken.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", "", err
	}
	refreshTokenStr, err := refreshToken.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", "", err
	}
	return accessTokenStr, refreshTokenStr, nil
}

func SaveRefreshToken(email, refreshToken string) error {
	return database.DB.Model(&models.User{}).
		Where("email = ?", email).
		Update("refresh_token", refreshToken).
		Error
}

func VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неизвестный метод подписи")
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}
	return nil, fmt.Errorf("недействительный токен")
}
