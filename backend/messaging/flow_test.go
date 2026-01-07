package messaging

import (
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestFullFlow(t *testing.T) {
	// Load env if exists
	_ = godotenv.Load("../../.env")

	log.Println("🚀 Starting Full Flow Refactor Test")

	// Reset counters
	atomic.StoreUint64(&CreatedCount, 0)
	atomic.StoreUint64(&UpdatedCount, 0)

	// 1. Initialize Producer
	if err := InitProducer(); err != nil {
		t.Fatalf("❌ Failed to init producer: %v", err)
	}

	// 2. Initialize Consumers
	if err := StartUserEventConsumer(); err != nil {
		t.Fatalf("❌ Failed to start consumers: %v", err)
	}
	defer StopUserEventConsumer()

	// Give consumers time to start and connect
	log.Println("⏳ Waiting for consumers to connect...")
	time.Sleep(3 * time.Second)

	// 3. Publish USER_CREATED
	log.Println("📤 Publishing USER_CREATED...")
	err := PublishUserEvent("USER_CREATED", 101, "John Refactor", "john@refactor.com")
	if err != nil {
		t.Errorf("❌ Publish USER_CREATED failed: %v", err)
	}

	// 4. Publish USER_UPDATED
	log.Println("📤 Publishing USER_UPDATED...")
	err = PublishUserEvent("USER_UPDATED", 101, "John Updated", "john@refactor.com")
	if err != nil {
		t.Errorf("❌ Publish USER_UPDATED failed: %v", err)
	}

	// 5. Wait for processing
	log.Println("⏳ Waiting for consumers to process events...")
	time.Sleep(5 * time.Second)

	cCount := atomic.LoadUint64(&CreatedCount)
	uCount := atomic.LoadUint64(&UpdatedCount)

	log.Printf("📊 Verification -> Created: %d, Updated: %d", cCount, uCount)

	if cCount == 0 {
		t.Errorf("❌ USER_CREATED was not received")
	}
	if uCount == 0 {
		t.Errorf("❌ USER_UPDATED was not received")
	}

	log.Println("✅ Full Flow Test Completed")
}
