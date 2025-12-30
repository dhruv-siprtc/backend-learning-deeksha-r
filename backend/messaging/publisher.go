package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type UserEvent struct {
	Event     string    `json:"event"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      UserData  `json:"data"`
}

type UserData struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func PublishUserEvent(
	rmq *RabbitMQ,
	routingKey string,
	eventName string,
	userID int,
	name string,
	email string,
) error {

	event := UserEvent{
		Event:     eventName,
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		Data: UserData{
			UserID: userID,
			Name:   name,
			Email:  email,
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rmq.Channel.PublishWithContext(
		ctx,
		"user.events", // exchange
		routingKey,    // routing key
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)

	if err != nil {
		return err
	}

	log.Printf("📤 Event published: %s (%s)", eventName, routingKey)
	return nil
}
