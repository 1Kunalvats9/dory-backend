package services

import (
	"context"
	"dory-backend/internal/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
)

// VerifyGoogleToken checks if the ID Token from the frontend is valid

func VerfiyGoogleToken(idTokenStr string) (*idtoken.Payload, error) {
	ctx := context.Background()
	payload, err := idtoken.Validate(context.Background(), idTokenStr, config.AppConfig.GoogleWebClientID)
	if err == nil {
		return payload, nil
	}

	payload, err = idtoken.Validate(ctx, idTokenStr, config.AppConfig.GoogleIOSClientID)
	if err == nil {
		return payload, nil
	}

	return nil, errors.New("identity verification failed for all platforms")
}
func GenerateJWTToken(userID string, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 72 * 3).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}
