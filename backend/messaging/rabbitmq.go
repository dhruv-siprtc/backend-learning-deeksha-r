package messaging

import (
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

var (
	conn     *amqp091.Connection
	connOnce sync.Once
)

// GetConnection returns a singleton RabbitMQ connection.
// This ensures that the application can maintain a single connection as requested,
// even if multiple worker pools are used for separate queues.
func GetConnection() (*amqp091.Connection, error) {
	var err error
	connOnce.Do(func() {
		url := GetRabbitMQURL()
		conn, err = amqp091.Dial(url)
		if err != nil {
			log.Printf("❌ Failed to connect to RabbitMQ: %v", err)
		} else {
			log.Println("✅ RabbitMQ singleton connection established")
		}
	})
	return conn, err
}

// CloseConnection closes the singleton connection
func CloseConnection() {
	if conn != nil {
		conn.Close()
		log.Println("👋 RabbitMQ connection closed")
	}
}
