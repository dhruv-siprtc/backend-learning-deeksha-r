package main

import (
	"go-backend-learning/backend/config"
	"go-backend-learning/backend/messaging"
	"go-backend-learning/backend/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found")
	}
	// Connect to database
	if err := config.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Connect to RabbitMQ
	if err := messaging.ConnectRabbitMQ(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	if messaging.Instance != nil {
		defer messaging.Instance.Close()
	}

	// Start RabbitMQ Consumer
	if err := messaging.StartUserEventConsumer(messaging.Instance); err != nil {
		log.Printf("Failed to start RabbitMQ consumer: %v", err)
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register routes
	routes.RegisterRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
