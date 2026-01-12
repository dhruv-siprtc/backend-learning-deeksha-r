package producer

import (
	"context"
	"fmt"
	"log"

	paotaconfig "github.com/surendratiwari3/paota/config"
	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

type RmqServices struct {
}

var (
	workerPoolMap = make(map[string]*workerpool.Pool)
)

func (rmqService *RmqServices) initRmqServices(rmqConfig RmqConfig) error {

	// Check if a worker pool for this queue already exists
	if _, ok := workerPoolMap[rmqConfig.QueueName]; ok {
		log.Printf("Worker pool already exists: %s", rmqConfig.QueueName)
		return nil
	}

	log.Printf("Initializing worker pool for queue: %s", rmqConfig.QueueName)

	rebitmq := paotaconfig.Config{
		Broker:        "amqp",
		TaskQueueName: rmqConfig.QueueName,
		AMQP: &paotaconfig.AMQPConfig{
			Url:                rmqConfig.RmQURL,
			Exchange:           rmqConfig.ExchangeName,
			ExchangeType:       "direct",
			BindingKey:         rmqConfig.BindingKey,
			PrefetchCount:      rmqConfig.PrefetchCount,
			ConnectionPoolSize: rmqConfig.ConnectionPoolSize,
			DelayedQueue:       rmqConfig.DelayedQueue,
			TimeoutQueue:       rmqConfig.TimeoutQueue,
			FailedQueue:        rmqConfig.FailedQueue,
		},
	}

	// Create worker pool
	wp, err := workerpool.NewWorkerPoolWithConfig(context.Background(), 10, rmqConfig.QueueName, rebitmq)
	if err != nil {
		log.Printf("Failed to initialize worker pool for queue %s: %v", rmqConfig.QueueName, err)
		return fmt.Errorf("failed to initialize worker pool: %w", err)
	}

	if wp == nil {
		return fmt.Errorf("worker pool is nil")
	}

	workerPoolMap[rmqConfig.QueueName] = &wp
	log.Printf("Worker pool successfully initialized: %s", rmqConfig.QueueName)
	return nil
}

func (rmqService *RmqServices) rmqPublish(message []byte, taskName, queueName string) error {

	pool, exists := workerPoolMap[queueName]
	if !exists {
		return fmt.Errorf("no worker pool found for queue: %s", queueName)
	}

	printJob := &schema.Signature{
		Name: taskName,
		Args: []schema.Arg{
			{
				Type:  "string",
				Value: string(message),
			},
		},
		RetryCount:                  50,
		IgnoreWhenTaskNotRegistered: true,
		Priority:                    2,
	}

	_, err := (*pool).SendTaskWithContext(context.Background(), printJob)
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}
	return nil
}
