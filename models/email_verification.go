package models

import "time"

// EMAIL VERIFICATION OTP (ONE ROW PER USER, REPLACED ON EVERY RESEND)
type EmailVerification struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"not null;uniqueIndex"`
	CodeHash        string    `gorm:"not null"`
	ExpiresAt       time.Time `gorm:"not null"`
	Attempts        int       `gorm:"not null;default:0"` // WRONG GUESSES FOR THE CURRENT CODE
	LastSentAt      time.Time `gorm:"not null"`
	SendCount       int       `gorm:"not null;default:1"` // CODES SENT INSIDE THE CURRENT WINDOW
	SendWindowStart time.Time `gorm:"not null"`
}
