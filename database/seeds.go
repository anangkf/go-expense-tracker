package database

import (
	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

func Seed() {
	seedBucketType()
	seedCategories()
	seedBudgetPlanTemplates()
	seedExpenseTemplates()
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
			ID:   t.ID,
			Name: t.Name,
		})
	}
}

func seedCategories() {
	bt1, bt2 := uint(1), uint(2)
	categories := []models.Category{
		// Expense
		{ID: 1, Name: "Food", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "beef"},
		{ID: 2, Name: "Transport", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "bus-front"},
		{ID: 3, Name: "Health", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "heart-pulse"},
		{ID: 4, Name: "Entertainment", Type: "expense", IsDefault: true, BucketTypeID: &bt2, IconName: "film"},
		{ID: 5, Name: "Bills", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "receipt"},
		{ID: 6, Name: "Groceries", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "shopping-bag"},
		{ID: 7, Name: "Savings", Type: "expense", IsDefault: true, BucketTypeID: &bt1, IconName: "hand-coins"},
		{ID: 8, Name: "Snack & Beverages", Type: "expense", IsDefault: true, BucketTypeID: &bt2, IconName: "hamburger"},
		// Income
		{ID: 9, Name: "Salary", Type: "income", IsDefault: true, IconName: "wallet"},
	}

	for _, c := range categories {
		DB.FirstOrCreate(&c, models.Category{
			ID:        c.ID,
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

func seedExpenseTemplates() {
	catFood, catTransport, catBills, catGroceries, catSavings, catSnackBeverages := uint(1), uint(2), uint(5), uint(6), uint(7), uint(8)
	bt1, bt2, bt3 := uint(1), uint(2), uint(3)

	templates := []models.ExpenseTemplate{
		{Name: "Makan", CategoryID: &catFood, BucketTypeID: &bt1, IconName: "ham"},
		{Name: "Parkir", CategoryID: &catTransport, BucketTypeID: &bt1, IconName: "square-parking"},
		{Name: "Internet", CategoryID: &catBills, BucketTypeID: &bt1, IconName: "wifi"},
		{Name: "Ojol", CategoryID: &catTransport, BucketTypeID: &bt1, IconName: "motorbike"},
		{Name: "Bayar Kos", CategoryID: &catBills, BucketTypeID: &bt1, IconName: "house"},
		{Name: "Kopi", CategoryID: &catSnackBeverages, BucketTypeID: &bt2, IconName: "coffee"},
		{Name: "Belanja", CategoryID: &catGroceries, BucketTypeID: &bt1, IconName: "shopping-cart"},
		{Name: "Menabung", CategoryID: &catSavings, BucketTypeID: &bt3, IconName: "hand-coins"},
	}

	for _, t := range templates {
		DB.FirstOrCreate(&t, models.ExpenseTemplate{
			Name:      t.Name,
			IsDefault: true,
		})
	}
}
