package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/surendratiwari3/paota/config"
	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

type ProducerService struct {
	workerPoolMap map[string]*workerpool.Pool
}

var producerInstance *ProducerService

// Initialize the producer service
func InitProducer() error {
	producerInstance = &ProducerService{
		workerPoolMap: make(map[string]*workerpool.Pool),
	}

	// Initialize worker pools for both queues
	configs := []RmqConfig{UserCreatedConfig, UserUpdatedConfig}

	for _, rmqConfig := range configs {
		rmqConfig.RmQURL = GetRabbitMQURL()
		if err := producerInstance.initWorkerPool(rmqConfig); err != nil {
			return fmt.Errorf("failed to init producer for %s: %w", rmqConfig.QueueName, err)
		}
	}

	log.Println("✅ Paota Producer initialized successfully")
	return nil
}

func (ps *ProducerService) initWorkerPool(rmqConfig RmqConfig) error {
	// Check if worker pool already exists
	if _, ok := ps.workerPoolMap[rmqConfig.QueueName]; ok {
		log.Printf("Worker pool already exists for queue: %s", rmqConfig.QueueName)
		return nil
	}

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

	// Create worker pool
	wp, err := workerpool.NewWorkerPoolWithConfig(
		context.Background(),
		10, // number of workers
		rmqConfig.QueueName,
		paotaConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create worker pool: %w", err)
	}

	ps.workerPoolMap[rmqConfig.QueueName] = &wp
	log.Printf("✅ Worker pool created for queue: %s", rmqConfig.QueueName)
	return nil
}

// Publish user event using Paota
func PublishUserEvent(
	eventName string,
	userID int,
	name string,
	email string,
) error {
	if producerInstance == nil {
		return fmt.Errorf("producer not initialized")
	}

	// Create event payload
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

	// Determine which queue to use
	var queueName, taskName string
	switch eventName {
	case "USER_CREATED":
		queueName = UserCreatedConfig.QueueName
		taskName = UserCreatedConfig.QueueTaskName
	case "USER_UPDATED":
		queueName = UserUpdatedConfig.QueueName
		taskName = UserUpdatedConfig.QueueTaskName
	default:
		return fmt.Errorf("unknown event type: %s", eventName)
	}

	return producerInstance.publish(string(eventJSON), taskName, queueName)
}

func (ps *ProducerService) publish(message string, taskName string, queueName string) error {
	pool, exists := ps.workerPoolMap[queueName]
	if !exists {
		return fmt.Errorf("no worker pool found for queue: %s", queueName)
	}

	// Create Paota task signature
	taskSignature := &schema.Signature{
		Name: taskName,
		Args: []schema.Arg{
			{
				Type:  "string",
				Value: message,
			},
		},
		RetryCount:                  3,
		IgnoreWhenTaskNotRegistered: false,
		Priority:                    2,
	}

	// Send task
	_, err := (*pool).SendTaskWithContext(context.Background(), taskSignature)
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}

	log.Printf("📤 Event published via Paota: %s (queue: %s)", taskName, queueName)
	return nil
}
