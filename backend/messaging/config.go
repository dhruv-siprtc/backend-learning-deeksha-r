package messaging

import (
	"fmt"
	"os"
)

type RmqConfig struct {
	QueueName          string
	ExchangeName       string
	BindingKey         string
	PrefetchCount      int
	ConnectionPoolSize int
	DelayedQueue       string
	RmQURL             string
	FailedQueue        string
	TimeoutQueue       string
	QueueTaskName      string
}

func GetRabbitMQURL() string {
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("RABBITMQ_PORT")
	if port == "" {
		port = "5672"
	}
	user := os.Getenv("RABBITMQ_USER")
	if user == "" {
		user = "guest"
	}
	password := os.Getenv("RABBITMQ_PASSWORD")
	if password == "" {
		password = "guest"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
}

// Configuration for USER_CREATED events
var UserCreatedConfig = RmqConfig{
	QueueName:          "user.created.queue",
	ExchangeName:       "user.events",
	BindingKey:         "user.created",
	PrefetchCount:      10,
	ConnectionPoolSize: 5,
	DelayedQueue:       "user.created.delayed",
	FailedQueue:        "user.created.failed",
	TimeoutQueue:       "user.created.timeout",
	QueueTaskName:      "handle_user_created",
}

// Configuration for USER_UPDATED events
var UserUpdatedConfig = RmqConfig{
	QueueName:          "user.updated.queue",
	ExchangeName:       "user.events",
	BindingKey:         "user.updated",
	PrefetchCount:      10,
	ConnectionPoolSize: 5,
	DelayedQueue:       "user.updated.delayed",
	FailedQueue:        "user.updated.failed",
	TimeoutQueue:       "user.updated.timeout",
	QueueTaskName:      "handle_user_updated",
}
