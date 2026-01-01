package messaging

import (
	"encoding/json"
	"log"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

func StartUserEventConsumer(rmq *Paota) error {
	// 1️⃣ Declare Dead Letter Exchange (DLX)
	dlxName := "user.events.dlx"
	err := rmq.Channel.ExchangeDeclare(
		dlxName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 2️⃣ Declare Dead Letter Queue (DLQ)
	dlqName := "user.events.dlq"
	_, err = rmq.Channel.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 3️⃣ Bind DLQ to DLX
	err = rmq.Channel.QueueBind(
		dlqName,
		"#", // bind all keys to DLQ for now
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 4️⃣ Declare Main Exchange
	exchangeName := "user.events"
	err = rmq.Channel.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 5️⃣ Declare queues with DLX configuration
	args := amqp091.Table{
		"x-dead-letter-exchange": dlxName,
	}

	createdQueue, err := rmq.Channel.QueueDeclare(
		"user.created.queue",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	updatedQueue, err := rmq.Channel.QueueDeclare(
		"user.updated.queue",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	// 6️⃣ Bind queues to main exchange
	err = rmq.Channel.QueueBind(
		createdQueue.Name,
		"user.created",
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = rmq.Channel.QueueBind(
		updatedQueue.Name,
		"user.updated",
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 7️⃣ Start consuming
	createdMsgs, err := rmq.Channel.Consume(
		createdQueue.Name,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	updatedMsgs, err := rmq.Channel.Consume(
		updatedQueue.Name,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go handleCreatedUsers(createdMsgs)
	go handleUpdatedUsers(updatedMsgs)

	log.Println("📥 RabbitMQ consumers started with DLQ support")
	return nil
}

// ---------------- HANDLERS ----------------

func handleCreatedUsers(messages <-chan Delivery) {
	for msg := range messages {
		var event UserEvent

		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("❌ Failed to parse USER_CREATED: %v", err)
			msg.Nack(false, false) // Requeue: false (send to DLQ)
			continue
		}

		log.Printf("[USER_CREATED] Welcome email sent to %s", event.Data.Email)

		if err := msg.Ack(false); err != nil {
			log.Printf("❌ Failed to ack message: %v", err)
		}
	}
}

func handleUpdatedUsers(messages <-chan Delivery) {
	for msg := range messages {
		var event UserEvent

		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("❌ Failed to parse USER_UPDATED: %v", err)
			msg.Nack(false, false) // Requeue: false (send to DLQ)
			continue
		}

		log.Printf("[USER_UPDATED] User %d profile updated", event.Data.UserID)

		if err := msg.Ack(false); err != nil {
			log.Printf("❌ Failed to ack message: %v", err)
		}
	}
}
