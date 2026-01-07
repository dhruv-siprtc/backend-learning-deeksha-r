package config

import (
	"fmt"
	"os"

	"go-backend-learning/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	host := os.Getenv("PGHOST")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")

	if host == "" || user == "" || dbname == "" || port == "" {
		return fmt.Errorf("database configuration missing")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = db
	fmt.Println("✅ PostgreSQL connected successfully")

	// Auto Migration
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Printf("❌ Failed to migrate database: %v\n", err)
		return err
	}
	fmt.Println("✅ Database migration completed")

	return nil
}
