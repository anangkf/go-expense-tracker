package models

import (
	"time"

	"gorm.io/gorm"
)

type BucketType struct {
	ID        uint           `gorm:"primarykey;autoIncrement:false" json:"id"`
	Name      string         `gorm:"size:255;not null;unique" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
