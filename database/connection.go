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
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// SET GORM LOG LEVEL BY SERVER MODE
	logLevel := logger.Warn
	if cfg.Server.Mode == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
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
	Seed()

	log.Println("Database connection successfully!")
}
