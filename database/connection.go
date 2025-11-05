package database

import (
	"fmt"
	"log"
	"time"

	"go-expense-tracker-api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	// CONFIGURE CONNECTION POOL
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying SQL SB:", err)
	}

	// SET CONNECTION POOL SETTINGS
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	// AUTO MIGRATE DATABASE SCHEMAS
	AutoMigrate()

	// SEEDS DATA
	SeedCategories()

	log.Println("Database connection successfully!")
}