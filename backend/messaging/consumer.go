package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/surendratiwari3/paota/config"
	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

type ConsumerService struct {
	workerPools []*workerpool.Pool
}

var consumerInstance *ConsumerService

// Initialize and start consumers
func StartUserEventConsumer() error {
	consumerInstance = &ConsumerService{
		workerPools: make([]*workerpool.Pool, 0),
	}

	// Start consumer for USER_CREATED
	if err := consumerInstance.startConsumer(UserCreatedConfig); err != nil {
		return fmt.Errorf("failed to start USER_CREATED consumer: %w", err)
	}

	// Start consumer for USER_UPDATED
	if err := consumerInstance.startConsumer(UserUpdatedConfig); err != nil {
		return fmt.Errorf("failed to start USER_UPDATED consumer: %w", err)
	}

	log.Println("📥 Paota Consumers started successfully")
	return nil
}

func (cs *ConsumerService) startConsumer(rmqConfig RmqConfig) error {
	rmqConfig.RmQURL = GetRabbitMQURL()

	// Create Paota config
	paotaConfig := config.Config{
		Broker:        "amqp",
		TaskQueueName: rmqConfig.QueueName,
		AMQP: &config.AMQPConfig{
			Url:                rmqConfig.RmQURL,
			Exchange:           rmqConfig.ExchangeName,
			ExchangeType:       "topic",
			BindingKey:         rmqConfig.BindingKey,
			PrefetchCount:      rmqConfig.PrefetchCount,
			ConnectionPoolSize: rmqConfig.ConnectionPoolSize,
			DelayedQueue:       rmqConfig.DelayedQueue,
			TimeoutQueue:       rmqConfig.TimeoutQueue,
			FailedQueue:        rmqConfig.FailedQueue,
		},
	}

	// Set application config
	err := config.GetConfigProvider().SetApplicationConfig(paotaConfig)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	// Create worker pool
	wp, err := workerpool.NewWorkerPool(
		context.Background(),
		10, // number of workers
		rmqConfig.QueueName,
	)
	if err != nil {
		return fmt.Errorf("failed to create worker pool: %w", err)
	}

	// Register task handlers
	tasks := map[string]interface{}{}
	switch rmqConfig.QueueTaskName {
	case UserCreatedConfig.QueueTaskName:
		tasks[rmqConfig.QueueTaskName] = handleUserCreatedTask
	case UserUpdatedConfig.QueueTaskName:
		tasks[rmqConfig.QueueTaskName] = handleUserUpdatedTask
	}

	err = wp.RegisterTasks(tasks)
	if err != nil {
		return fmt.Errorf("failed to register tasks: %w", err)
	}

	// Start the worker pool
	err = wp.Start()
	if err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	cs.workerPools = append(cs.workerPools, &wp)
	log.Printf("✅ Consumer started for queue: %s", rmqConfig.QueueName)
	return nil
}

// Task handler for USER_CREATED events
func handleUserCreatedTask(sig *schema.Signature) error {
	if len(sig.Args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	eventJSON, ok := sig.Args[0].Value.(string)
	if !ok {
		return fmt.Errorf("invalid argument type")
	}

	var event UserEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		log.Printf("❌ Failed to parse USER_CREATED: %v", err)
		return err
	}

	// Business logic for USER_CREATED
	log.Printf("✅ [USER_CREATED] Welcome email sent to %s (User ID: %d)",
		event.Data.Email, event.Data.UserID)

	// Add your actual business logic here
	// e.g., send email, update cache, etc.

	return nil
}

// Task handler for USER_UPDATED events
func handleUserUpdatedTask(sig *schema.Signature) error {
	if len(sig.Args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	eventJSON, ok := sig.Args[0].Value.(string)
	if !ok {
		return fmt.Errorf("invalid argument type")
	}

	var event UserEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		log.Printf("❌ Failed to parse USER_UPDATED: %v", err)
		return err
	}

	// Business logic for USER_UPDATED
	log.Printf("✅ [USER_UPDATED] User %d profile updated (Name: %s, Email: %s)",
		event.Data.UserID, event.Data.Name, event.Data.Email)

	// Add your actual business logic here
	// e.g., invalidate cache, send notification, etc.

	return nil
}
