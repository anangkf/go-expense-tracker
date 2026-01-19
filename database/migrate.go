package database

import (
	"go-expense-tracker-api/models"
	"log"
)

func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.BudgetPlan{},
		&models.BucketType{},
		&models.BudgetBucket{},
		&models.Category{},
		&models.Expense{},
		&models.ExpenseTemplate{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed!")
}
