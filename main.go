package main

import (
	"go-backend-learning/backend/config"
	"go-backend-learning/backend/consumer"
	"go-backend-learning/backend/producer"
	"go-backend-learning/backend/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found")
	}

	// 2. Initialize Config
	config.InitConfig()

	// 3. Connect to database
	if err := config.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 4. Initialize Producer anyway (if we are in API mode or ALL mode)
	// Even if it's worker mode, it doesn't hurt to have it ready if needed,
	// but following reference we should initialize it.
	prodConfig := producer.RmqConfig{
		QueueName:          config.Config.UserTaskProducer.QueueName,
		ExchangeName:       config.Config.UserTaskProducer.ExchangeName,
		BindingKey:         config.Config.UserTaskProducer.BindingKeyName,
		PrefetchCount:      1,
		ConnectionPoolSize: 1,
		DelayedQueue:       config.Config.UserTaskProducer.DelayQueueName,
		RmQURL:             config.Config.UserTaskProducer.RabbitMQUrl,
		FailedQueue:        config.Config.UserTaskProducer.FailedQueue,
		TimeoutQueue:       config.Config.UserTaskProducer.TimeoutQueue,
	}
	if err := producer.UserProducer.Initialize(prodConfig); err != nil {
		log.Fatalf("Failed to initialize producer: %v", err)
	}

	// 5. Determine service mode
	serviceLauncher := os.Getenv("SERVICE_LAUNCHER")
	if serviceLauncher == "" {
		serviceLauncher = config.Config.Server.ServiceLauncher
	}

	log.Printf("🚀 Starting service launcher: %s", serviceLauncher)

	switch serviceLauncher {
	case "CONSUMER":
		startConsumer()
		// Block forever
		select {}

	case "PRODUCER", "API":
		startServer()

	case "ALL":
		go startConsumer()
		startServer()

	default:
		log.Fatalf("Unknown SERVICE_LAUNCHER: %s. Valid values: 'CONSUMER', 'PRODUCER', 'API', 'ALL'", serviceLauncher)
	}
}

func startConsumer() {
	consConfig := consumer.RmqConfig{
		QueueName:          config.Config.ConsumerConf.QueueName,
		ExchangeName:       config.Config.ConsumerConf.ExchangeName,
		BindingKey:         config.Config.ConsumerConf.BindingKeyName,
		PrefetchCount:      10,
		ConnectionPoolSize: 1,
		DelayedQueue:       config.Config.ConsumerConf.DelayQueueName,
		RmQURL:             config.Config.ConsumerConf.RabbitMQUrl,
		FailedQueue:        config.Config.ConsumerConf.FailedQueue,
		TimeoutQueue:       config.Config.ConsumerConf.TimeoutQueue,
	}
	mgr := &consumer.ConsumerManager{}
	if err := mgr.Initialize(consConfig); err != nil {
		log.Fatalf("Failed to initialize consumer: %v", err)
	}
}

func startServer() {
	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	routes.RegisterRoutes(e)

	// Start server
	port := config.Config.Server.Port
	if port == "" {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
}
