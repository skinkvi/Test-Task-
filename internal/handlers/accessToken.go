package handlers

import (
	"time"

	"github.com/skinkvi/TestTaskBackDev/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(userID, ipAddress string, cfg config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user_id": userID,
		"ip":      ipAddress,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString([]byte(cfg.JWT.Secret))
}
