package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

var (
	CreatedCount uint64
	UpdatedCount uint64
)

type ConsumerService struct {
	workerPools []workerpool.Pool
	mu          sync.Mutex
}

func (cs *ConsumerService) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for _, wp := range cs.workerPools {
		log.Printf("🛑 Stopping worker pool...")
		wp.Stop()
	}
	cs.workerPools = nil
}

var consumerInstance *ConsumerService

// StartUserEventConsumer initializes and starts consumers for different events
func StartUserEventConsumer() error {
	consumerInstance = &ConsumerService{
		workerPools: make([]workerpool.Pool, 0),
	}

	appCtx := struct{}{}

	// 1. Initialize first pool (USER_CREATED)
	createdCfg := GetPaotaConfig(UserCreatedQueue, UserCreatedKey)
	wp1, err := workerpool.NewWorkerPoolWithConfig(appCtx, 10, UserCreatedQueue, createdCfg)
	if err != nil {
		return fmt.Errorf("failed to create created_pool: %w", err)
	}

	// 2. Initialize second pool (USER_UPDATED)
	updatedCfg := GetPaotaConfig(UserUpdatedQueue, UserUpdatedKey)
	wp2, err := workerpool.NewWorkerPoolWithConfig(appCtx, 10, UserUpdatedQueue, updatedCfg)
	if err != nil {
		return fmt.Errorf("failed to create updated_pool: %w", err)
	}

	// MANDATORY: Share broker to maintain single connection
	sharedBroker := wp1.GetBroker()
	wp2.SetBroker(sharedBroker)
	log.Printf("🔌 Broker shared between %s and %s", UserCreatedQueue, UserUpdatedQueue)

	// Register all tasks to both pools (since they share the same broker/registry now)
	allTasks := map[string]interface{}{
		UserCreatedTask: handleUserCreatedTask,
		UserUpdatedTask: handleUserUpdatedTask,
	}

	if err := wp1.RegisterTasks(allTasks); err != nil {
		return fmt.Errorf("failed to register tasks on wp1: %w", err)
	}
	if err := wp2.RegisterTasks(allTasks); err != nil {
		return fmt.Errorf("failed to register tasks on wp2: %w", err)
	}

	// Track pools
	consumerInstance.mu.Lock()
	consumerInstance.workerPools = append(consumerInstance.workerPools, wp1, wp2)
	consumerInstance.mu.Unlock()

	// Start consumers in background
	go func() {
		log.Printf("📥 Starting USER_CREATED consumer loop")
		if err := wp1.Start(); err != nil {
			log.Printf("❌ USER_CREATED consumer error: %v", err)
		}
	}()

	go func() {
		log.Printf("📥 Starting USER_UPDATED consumer loop")
		if err := wp2.Start(); err != nil {
			log.Printf("❌ USER_UPDATED consumer error: %v", err)
		}
	}()

	log.Println("📥 Paota Consumers aligned and started")
	return nil
}

// StopUserEventConsumer stops all running consumers
func StopUserEventConsumer() {
	if consumerInstance != nil {
		consumerInstance.Stop()
	}
}

// handleUserCreatedTask processes USER_CREATED events.
func handleUserCreatedTask(sig *schema.Signature) error {
	log.Printf("🔍 DISPATCHED: %s", UserCreatedTask)

	if len(sig.Args) == 0 {
		return fmt.Errorf("no arguments in signature")
	}

	eventJSON, ok := sig.Args[0].Value.(string)
	if !ok {
		return fmt.Errorf("invalid argument type: expected string")
	}

	var event UserEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return fmt.Errorf("failed to unmarshal USER_CREATED: %w", err)
	}

	log.Printf("👤 [EVENT] USER_CREATED: %d (%s)", event.Data.UserID, event.Data.Email)
	atomic.AddUint64(&CreatedCount, 1)
	return nil
}

// handleUserUpdatedTask processes USER_UPDATED events.
func handleUserUpdatedTask(sig *schema.Signature) error {
	log.Printf("🔍 DISPATCHED: %s", UserUpdatedTask)

	if len(sig.Args) == 0 {
		return fmt.Errorf("no arguments in signature")
	}

	eventJSON, ok := sig.Args[0].Value.(string)
	if !ok {
		return fmt.Errorf("invalid argument type: expected string")
	}

	var event UserEvent
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return fmt.Errorf("failed to unmarshal USER_UPDATED: %w", err)
	}

	log.Printf("📝 [EVENT] USER_UPDATED: %d (%s)", event.Data.UserID, event.Data.Email)
	atomic.AddUint64(&UpdatedCount, 1)
	return nil
}
