package config

import (
	"fmt"

	"go-backend-learning/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	host := Config.Postgres.Host
	user := Config.Postgres.User
	password := Config.Postgres.Password
	dbname := Config.Postgres.DB
	port := Config.Postgres.Port

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
