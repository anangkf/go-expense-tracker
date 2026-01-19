package models

import (
	"time"

	"gorm.io/gorm"
)

type ExpenseTemplate struct {
	ID           uint    `json:"id" gorm:"primaryKey, autoIncrement"`
	Name         string  `json:"name"`
	Amount       float64 `json:"amount"`
	IsDefault    bool    `json:"is_default" gorm:"default:false"`
	UserID       *uint   `json:"-" gorm:"foreignKey:UserID;references:ID;index"`
	CategoryID   *uint   `json:"-" gorm:"foreignKey:CategoryID;references:ID"`
	BucketTypeID *uint   `json:"-" gorm:"index"`

	// RELATIONSHIPS
	Category Category `json:"category" gorm:"foreignKey:CategoryID;references:ID"`
	// BucketType BucketType `json:"bucket_type" gorm:"foreignKey:BucketTypeID"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type ExpenseTemplateReponse struct {
	ExpenseTemplate
	Total int64 `json:"total,omitempty"`
	Page  int   `json:"page,omitempty"`
	Limit int   `json:"limit,omitempty"`
}

type ExpenseTemplateRequest struct {
	Name       string  `json:"name" validate:"required"`
	Amount     float64 `json:"amount"`
	CategoryID uint    `json:"category_id,omitempty"`
}
