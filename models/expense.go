package models

import "time"

type Expense struct {
	ID         uint    `json:"id" gorm:"primaryKey, autoIncrement"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	UserID     uint    `json:"-" gorm:"foreignKey:UserID;references:ID"`
	CategoryID uint    `json:"-" gorm:"foreignKey:CategoryID;references:ID"`

	// RELATIONSHIPS
	Category Category `json:"category" gorm:"foreignKey:CategoryID;references:ID"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt time.Time `json:"deleted_at" gorm:"index"`
}

type ExpenseRequest struct {
	Name       string  `json:"name" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,gt=0"`
	CategoryID uint    `json:"category_id" gorm:"foreignKey:CategoryID;references:ID"`
}
