package database

import (
	"go-expense-tracker-api/models"
	"log"
)

func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Expense{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed!")
}
