package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

type ProducerService struct {
	createdPool workerpool.Pool
	updatedPool workerpool.Pool
}

var producerInstance *ProducerService

// InitProducer initializes the producer worker pools with a shared broker for single connection
func InitProducer() error {
	producerInstance = &ProducerService{}

	appCtx := struct{}{}

	// 1. Setup USER_CREATED pool (this will create the initial broker and connection)
	createdCfg := GetPaotaConfig(UserCreatedQueue, UserCreatedKey)
	cp1, err := workerpool.NewWorkerPoolWithConfig(appCtx, 1, UserCreatedQueue, createdCfg)
	if err != nil {
		return fmt.Errorf("failed to init created pool: %w", err)
	}
	producerInstance.createdPool = cp1

	// 2. Setup USER_UPDATED pool but share the broker from the first pool
	// This ensures we use a single RabbitMQ connection but separate logical pools
	updatedCfg := GetPaotaConfig(UserUpdatedQueue, UserUpdatedKey)
	cp2, err := workerpool.NewWorkerPoolWithConfig(appCtx, 1, UserUpdatedQueue, updatedCfg)
	if err != nil {
		return fmt.Errorf("failed to init updated pool: %w", err)
	}

	// Share the broker to maintain a single connection
	cp2.SetBroker(cp1.GetBroker())
	producerInstance.updatedPool = cp2

	log.Println("✅ Paota Producers initialized (Shared Connection)")
	return nil
}

// PublishUserEvent routes the event to the appropriate queue via Paota
func PublishUserEvent(eventName string, userID int, name string, email string) error {
	if producerInstance == nil {
		return fmt.Errorf("producer not initialized")
	}

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

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	var pool workerpool.Pool
	var routingKey string
	var taskName string

	switch eventName {
	case "USER_CREATED":
		pool = producerInstance.createdPool
		routingKey = UserCreatedKey
		taskName = UserCreatedTask
	case "USER_UPDATED":
		pool = producerInstance.updatedPool
		routingKey = UserUpdatedKey
		taskName = UserUpdatedTask
	default:
		return fmt.Errorf("unknown event type: %s", eventName)
	}

	// Create signature with explicit RoutingKey to ensure it reaches the queue reliably
	taskSignature, err := schema.NewSignature(taskName, []schema.Arg{
		{
			Type:  "string",
			Value: string(eventJSON),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	// CRITICAL: Set RoutingKey explicitly so the topic exchange routes it correctly
	taskSignature.RoutingKey = routingKey
	taskSignature.RetryCount = 3
	taskSignature.Priority = 1

	_, err = pool.SendTaskWithContext(context.Background(), taskSignature)
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}

	log.Printf("📤 Event published: %s (RoutingKey: %s)", taskName, routingKey)
	return nil
}
