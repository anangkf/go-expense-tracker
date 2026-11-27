package database

import (
	"go-expense-tracker-api/models"
	"log"

	"gorm.io/gorm"
)

func AutoMigrate() {
	// ADD email_verified_at AND BACKFILL EXISTING USERS (MUST RUN BEFORE AUTOMIGRATE)
	backfillEmailVerifiedAt()

	err := DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.EmailVerification{},
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

// MARK EVERY PRE-EXISTING USER AS VERIFIED, ONLY WHEN THE COLUMN IS INTRODUCED.
// COLUMN CREATION AND BACKFILL SHARE ONE TRANSACTION, SO A FAILED BACKFILL ROLLS BACK THE COLUMN AND IS RETRIED ON THE NEXT START.
func backfillEmailVerifiedAt() {
	m := DB.Migrator()
	if !m.HasTable(&models.User{}) || m.HasColumn(&models.User{}, "email_verified_at") {
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AddColumn(&models.User{}, "EmailVerifiedAt"); err != nil {
			return err
		}
		// RAW SQL ALSO COVERS SOFT-DELETED USERS
		return tx.Exec("UPDATE users SET email_verified_at = NOW() WHERE email_verified_at IS NULL").Error
	})
	if err != nil {
		log.Fatal("Failed to backfill email_verified_at:", err)
	}
	log.Println("Backfilled email_verified_at for existing users")
}
