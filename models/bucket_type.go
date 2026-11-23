package models

import (
	"time"

	"gorm.io/gorm"
)

// VALID BUCKET TYPE IDS (SEEDED WITH FIXED IDS: 1 NEEDS, 2 WANTS, 3 SAVINGS, 4 INVESTMENTS)
var ValidBucketTypeIDs = []uint{1, 2, 3, 4}

// IsValidBucketTypeID REPORTS WHETHER id IS ONE OF THE SEEDED BUCKET TYPE IDS
func IsValidBucketTypeID(id uint) bool {
	for _, valid := range ValidBucketTypeIDs {
		if id == valid {
			return true
		}
	}
	return false
}

type BucketType struct {
	ID        uint           `gorm:"primarykey;autoIncrement:false" json:"id"`
	Name      string         `gorm:"size:255;not null;unique" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
