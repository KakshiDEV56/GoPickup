package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"go_pickup/models"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func StartLocationConsumer() error {
	// Initialize Redis client with options
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Println("✅ Connected to Redis")
			return nil
		},
		MaxRetries:      3,
		MaxRetryBackoff: time.Second * 2,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %v", err)
	}
// convert the below written code into the a function that i can pass on value to change its value rather than always using the hardcode one
	// Initialize Kafka reader with configuration
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "location-updates",
		GroupID: "tracker-service",
		// Start consuming from the beginning if no offset is stored
		StartOffset: kafka.FirstOffset,
		// Minimum number of bytes to wait for
		MinBytes: 10e3, // 10KB
		// Maximum number of bytes to wait for
		MaxBytes: 10e6, // 10MB
		// Maximum wait time for new data
		MaxWait: 3 * time.Second,
		// Logger: log.New(os.Stdout, "kafka reader: ", 0),
	})
	defer reader.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go func() {
		for {
			// Create a context with timeout for each read operation
			readCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Read message from Kafka
			msg, err := reader.ReadMessage(readCtx)
			cancel() // Cancel context after read attempt

			if err != nil {
				log.Printf("❌ Error reading from Kafka: %v", err)
				// Check if we should exit
				select {
				case <-done:
					return
				default:
					continue
				}
			}

			// Parse the location update
			var location models.LocationUpdate

			if err := json.Unmarshal(msg.Value, &location); err != nil {
				log.Printf("❌ Error parsing message: %v", err)
				continue
			}
			// Validate and create Redis key for storing latest location
			var locationKey string
			if !location.ID.IsZero() && len(location.ID.Hex()) == 24 { // Check if ID is valid (ObjectID.Hex() is 24 chars)
				locationKey = fmt.Sprintf("location:%s:%s", location.ID.Hex(), location.ParcelID)
			} else {
				log.Printf("❌ Invalid or empty Agent ID received: %v", location.ID)
				continue // Skip this message
			}

			// Store latest position in Redis with 24h expiry
			if err := rdb.Set(context.Background(), locationKey, msg.Value, 24*time.Hour).Err(); err != nil {
				log.Printf("❌ Error storing in Redis: %v", err)
				continue
			}

			// Publish to Redis channel for real-time updates
			channel := fmt.Sprintf("realtime-locations:%s", location.ParcelID)
			if err := rdb.Publish(context.Background(), channel, msg.Value).Err(); err != nil {
				log.Printf("❌ Redis publish error: %v", err)
				continue
			}

			log.Printf("✅ Processed location update: Agent=%s, Parcel=%s, Status=%s",
				location.ID, location.ParcelID, location.Status)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutting down consumer...")
	done <- true

	return nil
}
