package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string         `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	UserID    *uint          `json:"-" gorm:"index"`
	Type      string         `json:"type" gorm:"not null" validate:"required,oneof=expense income"`
	IsDefault bool           `json:"is_default" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CategoryRequest struct {
	Name string `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Type string `json:"type" gorm:"not null" validate:"required,oneof=expense income"`
}

type DeleteCategoryResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	UserID    *uint     `json:"-" gorm:"index"`
	Type      string    `json:"type" gorm:"not null" validate:"required,oneof=expense income"`
	IsDefault bool      `json:"is_default" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
