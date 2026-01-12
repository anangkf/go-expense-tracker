package models

import (
	"time"

	"gorm.io/gorm"
)

type BudgetPlan struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      *uint          `json:"user_id" gorm:"index;references:ID"` // Nullable for templates
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	IsTemplate  bool           `gorm:"default:false" json:"is_template"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// RELATIONSHIPS
	Buckets []BudgetBucket `json:"buckets" gorm:"foreignKey:BudgetPlanID;constraint:OnDelete:CASCADE"`
}

type BudgetPlanRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty"`
	IsTemplate  bool   `json:"is_template" validate:"omitempty"`
	SetAsActive bool   `json:"set_as_active" validate:"omitempty"`

	// RELATIONSHIPS
	Buckets []BudgetBucketRequest `json:"buckets"`
}

type DeleteBudgetPlanResponse struct {
	ID          uint      `json:"id"`
	UserID      *uint     `json:"user_id"` // Nullable for templates
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsTemplate  bool      `json:"is_template"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`

	// RELATIONSHIPS
	Buckets []DeleteBudgetBucketResponse `json:"buckets" gorm:"foreignKey:BudgetPlanID"`
}

type BudgetBucketResponse struct {
	BudgetBucket
	MaxAllocation      float64 `json:"max_allocation"`
	TotalSpending      float64 `json:"total_spending"`
	SpendingPercentage float64 `json:"spending_percentage"`
}

type ActiveBudgetPlanResponse struct {
	BudgetPlan
	TotalIncome float64                `json:"total_income"`
	Buckets     []BudgetBucketResponse `json:"buckets"`
}
