package models

import (
	"time"
)

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	JTI       string    `gorm:"not null;uniqueIndex" json:"jti"` // JWT ID
	Token     string    `gorm:"not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"not null;default:false" json:"is_revoked"`
}
