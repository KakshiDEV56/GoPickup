package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress/snappy"
	"go_pickup/models"
	"log"
	"time"
)

// Producer encapsulates the Kafka producer functionality
type Producer struct {
	// The underlying Kafka writer
	writer *kafka.Writer
	// Topic to write messages to
	topic string
	// Configuration options
	config ProducerConfig
}

// ProducerConfig holds configuration for the Kafka producer
type ProducerConfig struct {
	// List of broker addresses
	Brokers []string
	// Kafka topic for location updates
	Topic string
	// Number of messages to batch before sending
	BatchSize int
	// Maximum time to wait before sending a batch
	BatchTimeout time.Duration
	// Number of required acknowledgments
	RequiredAcks kafka.RequiredAcks
}

// DefaultConfig returns default configuration for the producer
func DefaultConfig() ProducerConfig {
	return ProducerConfig{
		Brokers:      []string{"localhost:9092"},
		Topic:        "location-updates",
		BatchSize:    100,
		BatchTimeout: time.Millisecond * 100,
		RequiredAcks: kafka.RequireOne,
	}
}

// NewProducer creates a new Kafka producer with the given configuration
func NewProducer(config ProducerConfig) (*Producer, error) {
	// Validate configuration
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker address is required")
	}
	if config.Topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	// Create Kafka writer with optimized settings
	writer := kafka.NewWriter(kafka.WriterConfig{

		Brokers: config.Brokers,
		Topic:   config.Topic,
		// Use Hash balancer to ensure messages from same agent go to same partition
		Balancer: &kafka.Hash{},
		// Enable async writes for better performance
		Async:            true,
		BatchSize:        config.BatchSize,
		BatchTimeout:     config.BatchTimeout,
		RequiredAcks:     int(kafka.RequireOne),
		CompressionCodec: &snappy.Codec{},
		// Compression is not supported in kafka.WriterConfig; field removed
	})

	log.Printf("✅ Created Kafka producer for topic: %s", config.Topic)
	return &Producer{
		writer: writer,
		topic:  config.Topic,
		config: config,
	}, nil
}

// SendLocationUpdate sends a location update to Kafka
func (p *Producer) SendLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
	if update.Timestamp.IsZero() {
		update.Timestamp = time.Now()
	}

	// Validate coordinates
	if !isValidCoordinate(update.Latitude, update.Longitude) {
		return fmt.Errorf("invalid coordinates: lat=%v, lng=%v",
			update.Latitude, update.Longitude)
	}

	// Serialize the update
	value, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %v", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   update.ID[:], // Use AgentID for consistent partitioning
		Value: value,
		Time:  update.Timestamp,
		// Add headers for message metadata
		Headers: []kafka.Header{
			{Key: "version", Value: []byte("1.0")},
			{Key: "type", Value: []byte("location-update")},
		},
	}

	// Write message with context for timeout/cancellation
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	log.Printf("✅ Location update sent: Agent=%s, Parcel=%s",
		update.ID, update.ParcelID)
	return nil
}

// Close gracefully shuts down the producer
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}
	log.Printf("✅ Kafka producer closed for topic: %s", p.topic)
	return nil
}

// isValidCoordinate checks if coordinates are valid
func isValidCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// GetStats returns producer statistics
func (p *Producer) GetStats() kafka.WriterStats {
	return p.writer.Stats()
}
