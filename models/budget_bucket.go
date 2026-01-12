package models

import (
	"time"

	"gorm.io/gorm"
)

type BudgetBucket struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	BudgetPlanID uint           `gorm:"not null" json:"budget_plan_id"`
	BucketTypeID uint           `gorm:"not null" json:"bucket_type_id"`
	Name         string         `gorm:"size:255;not null" json:"name"`
	Percentage   float64        `gorm:"type:decimal(5,2);not null" json:"percentage"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// RELATIONSHIPS
	BucketType BucketType `gorm:"foreignKey:BucketTypeID" json:"bucket_type"`
}

type BudgetBucketRequest struct {
	ID           *uint   `json:"id,omitempty"`
	Name         string  `json:"name" validate:"required"`
	BucketTypeID uint    `json:"bucket_type_id" validate:"required"`
	Percentage   float64 `json:"percentage" validate:"required,gt=0"`
}

type DeleteBudgetBucketResponse struct {
	ID           uint      `json:"id"`
	BudgetPlanID uint      `json:"budget_plan_id"`
	BucketTypeID uint      `json:"bucket_type_id"`
	Name         string    `json:"name"`
	Percentage   float64   `json:"percentage"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`

	// RELATIONSHIPS
	BucketType BucketType `gorm:"foreignKey:BucketTypeID" json:"bucket_type"`
}
