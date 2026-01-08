package database

import (
	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

func Seed() {
	seedBucketType()
	seedCategories()
	seedBudgetPlanTemplates()
}

func seedBucketType() {
	bucketTypes := []models.BucketType{
		{ID: 1, Name: "Needs"},
		{ID: 2, Name: "Wants"},
		{ID: 3, Name: "Savings"},
		{ID: 4, Name: "Investments"},
	}

	for _, t := range bucketTypes {
		DB.FirstOrCreate(&t, models.BucketType{
			Name: t.Name,
		})
	}
}

func seedCategories() {
	categories := []models.Category{
		{Name: "Food", Type: "expense", IsDefault: true, BucketTypeID: 1},
		{Name: "Transport", Type: "expense", IsDefault: true, BucketTypeID: 1},
		{Name: "Health", Type: "expense", IsDefault: true, BucketTypeID: 1},
		{Name: "Entertainment", Type: "expense", IsDefault: true, BucketTypeID: 2},
		{Name: "Bills", Type: "expense", IsDefault: true, BucketTypeID: 1},
		{Name: "Salary", Type: "income", IsDefault: true},
	}

	for _, c := range categories {
		DB.FirstOrCreate(&c, models.Category{
			Name:      c.Name,
			IsDefault: true,
			UserID:    nil,
		})
	}
}

func seedBudgetPlanTemplates() {
	templates := []models.BudgetPlan{
		{
			Name:        "50/30/20 Rule",
			Description: "A simple budgeting rule: 50% for needs, 30% for wants, and 20% for savings.",
			IsTemplate:  true,
			Buckets: []models.BudgetBucket{
				{Name: "Needs", Percentage: 50, BucketTypeID: 1},
				{Name: "Wants", Percentage: 30, BucketTypeID: 2},
				{Name: "Savings", Percentage: 20, BucketTypeID: 3},
			},
		},
		{
			Name:        "4/3/2/1 Rule",
			Description: "A variation for prioritizing investment: 40% for needs, 30% for wants, 20% for savings, and 10% for investments.",
			IsTemplate:  true,
			Buckets: []models.BudgetBucket{
				{Name: "Needs", Percentage: 40, BucketTypeID: 1},
				{Name: "Wants", Percentage: 30, BucketTypeID: 2},
				{Name: "Savings", Percentage: 20, BucketTypeID: 3},
				{Name: "Investments", Percentage: 10, BucketTypeID: 4},
			},
		},
	}

	for _, t := range templates {
		// Using transaction to ensure atomicity
		DB.Transaction(func(tx *gorm.DB) error {
			var existing models.BudgetPlan
			if err := tx.Where("name = ? AND is_template = ?", t.Name, true).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Create the plan
					if err := tx.Create(&t).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
			return nil
		})
	}
}
