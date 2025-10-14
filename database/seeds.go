package database

import (
	"go-expense-tracker-api/models"
)

func SeedCategories() {
	categories := []models.Category{
		{Name: "Food", Type: "expense", IsDefault: true},
		{Name: "Transport", Type: "expense", IsDefault: true},
		{Name: "Health", Type: "expense", IsDefault: true},
		{Name: "Entertainment", Type: "expense", IsDefault: true},
		{Name: "Bills", Type: "expense", IsDefault: true},
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
