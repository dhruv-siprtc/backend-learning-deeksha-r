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

	// Determine service mode
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "all" // Default to running both
	}

	log.Printf("🚀 Starting service in mode: %s", serviceName)

	switch serviceName {
	case "worker":
		// Start RabbitMQ Consumer Only
		if err := messaging.StartUserEventConsumer(messaging.Instance); err != nil {
			log.Fatalf("Failed to start consumers: %v", err)
		}
		// Block forever to keep the consumer running
		select {}

	case "api":
		startServer()

	case "all":
		// Start RabbitMQ Consumer
		if err := messaging.StartUserEventConsumer(messaging.Instance); err != nil {
			log.Fatalf("Failed to start consumers: %v", err)
		}
		// Start API Server
		startServer()

	default:
		log.Fatalf("Unknown SERVICE_NAME: %s. Valid values: 'worker', 'api', 'all'", serviceName)
	}
}

func startServer() {
	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routes.RegisterRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
