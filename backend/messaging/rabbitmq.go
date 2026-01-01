package messaging

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Paota struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

var Instance *Paota

// ConnectRabbitMQ creates connection & channel
func ConnectRabbitMQ() error {
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

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("✅ RabbitMQ connected successfully")

	Instance = &Paota{
		Conn:    conn,
		Channel: ch,
	}

	return nil
}

func (p *Paota) Close() {
	if p.Channel != nil {
		p.Channel.Close()
	}
	if p.Conn != nil {
		p.Conn.Close()
	}
}

// Publish generic message
func (p *Paota) Publish(exchange, routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.Channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}

// Alias for RabbitMQ delivery type
type Delivery = amqp091.Delivery
