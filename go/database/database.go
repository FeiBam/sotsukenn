package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(logLevel logger.LogLevel) (*gorm.DB, error) {
	dbPath := getEnv("DB_PATH", "./sotsukenn.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	log.Printf("Database connection established: %s", dbPath)
	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}

func GetDBWithLogger(logLevel logger.LogLevel) (*gorm.DB, error) {
	if DB != nil {
		return DB, nil
	}
	return InitDB(logLevel)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
