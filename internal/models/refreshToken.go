package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	IPAddress string
}
