package config

import (
	"fmt"
	"log"
	"os"

	"go-backend-learning/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitTestDB initializes database connection for tests
func InitTestDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to test database:", err)
	}

	DB = db

	// Auto-create tables needed for tests
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("❌ Failed to migrate test database:", err)
	}
}
