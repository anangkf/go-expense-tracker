package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Name      string         `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Password  string         `json:"-" gorm:"not null" validate:"required,min=6"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// RELATIONSHIPS
	Categories []Category `json:"categories" gorm:"foreignKey:UserID"`
}

// LOGIN REQUEST PAYLOAD
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LOGIN RESPONSE
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// REGISTER REQUEST PAYLOAD
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// REGISTER RESPONSE
type RegisterResponse struct {
	User         UserResponse `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
}

// REFRESH TOKEN REQUEST PAYLOAD
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// REFRESH TOKEN RESPONSE
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// USER RESPONSE
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`

	// RELATIONSHIPS
	Categories []*Category `json:"categories"`
}
