package messaging

import (
	"fmt"
	"os"

	"github.com/surendratiwari3/paota/config"
)

// RmqConfig holds RabbitMQ and Paota specific configuration for a queue
type RmqConfig struct {
	QueueName          string
	ExchangeName       string
	BindingKey         string
	PrefetchCount      int
	ConnectionPoolSize int
	DelayedQueue       string
	FailedQueue        string
	TimeoutQueue       string
	QueueTaskName      string
}

// Event configuration constants
const (
	UserEventsExchange = "user.events.exchange"

	UserCreatedQueue = "user.created.queue"
	UserCreatedKey   = "user.created.key"
	UserCreatedTask  = "task.user.created"

	UserUpdatedQueue = "user.updated.queue"
	UserUpdatedKey   = "user.updated.key"
	UserUpdatedTask  = "task.user.updated"
)

// GetRabbitMQURL returns the connection string from environment
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

// GetPaotaConfig generates a config for a specific queue
func GetPaotaConfig(queueName, bindingKey string) config.Config {
	// If no specific binding key is provided, use a wildcard for user events
	// to allow broker sharing across different user queues.
	if bindingKey == "" {
		bindingKey = "user.#"
	}

	return config.Config{
		Broker:        "amqp",
		TaskQueueName: queueName,
		AMQP: &config.AMQPConfig{
			Url:                GetRabbitMQURL(),
			Exchange:           UserEventsExchange,
			ExchangeType:       "topic",
			BindingKey:         bindingKey,
			PrefetchCount:      10,
			ConnectionPoolSize: 1, // CONCEPT: Maintain 1 connection (shared via Broker in service layer)
			DelayedQueue:       queueName + ".delayed",
			FailedQueue:        queueName + ".failed",
			TimeoutQueue:       queueName + ".timeout",
			// Match existing queue arguments to avoid PRECONDITION_FAILED
			QueueDeclareArgs: config.QueueDeclareArgs{
				"x-dead-letter-exchange": "user.events.dlx",
			},
		},
	}
}
