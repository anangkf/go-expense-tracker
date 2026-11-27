package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go-expense-tracker-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// RUNS AGAINST A THROWAWAY SCHEMA THAT IS DROPPED AFTERWARDS
func TestBackfillEmailVerifiedAt(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is empty, skipping migration integration test")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("mig_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE") })

	// ONE CONNECTION SO search_path APPLIES TO EVERY STATEMENT
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	// PRE-FEATURE SCHEMA: users WITHOUT email_verified_at
	if err := db.Exec(`CREATE TABLE users (
		id bigserial primary key, email text not null, name text not null, password text not null,
		active_budget_plan_id bigint, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz)`).Error; err != nil {
		t.Fatal(err)
	}
	db.Exec("INSERT INTO users (email, name, password) VALUES ('a@t.local','A','x')")
	db.Exec("INSERT INTO users (email, name, password, deleted_at) VALUES ('b@t.local','B','x', now())")

	DB = db
	AutoMigrate()

	var nullCount int64
	db.Unscoped().Model(&models.User{}).Where("email_verified_at IS NULL").Count(&nullCount)
	if nullCount != 0 {
		t.Fatalf("expected all existing users verified, %d still NULL", nullCount)
	}

	// A USER REGISTERED AFTER THE ROLLOUT MUST STAY UNVERIFIED ON THE NEXT START
	db.Exec("INSERT INTO users (email, name, password) VALUES ('c@t.local','C','x')")
	AutoMigrate()
	db.Model(&models.User{}).Where("email = ?", "c@t.local").Count(&nullCount)
	var c models.User
	db.Where("email = ?", "c@t.local").First(&c)
	if c.EmailVerifiedAt != nil {
		t.Fatal("backfill must not run again once the column exists")
	}
}
