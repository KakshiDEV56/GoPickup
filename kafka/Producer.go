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

type Producer struct {
    writer *kafka.Writer
}

type ProducerConfig struct {
    Brokers      []string
    Topic        string
    BatchSize    int
    BatchTimeout time.Duration
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
    if len(cfg.Brokers) == 0 {
        return nil, fmt.Errorf("no brokers provided")
    }
    if cfg.Topic == "" {
        return nil, fmt.Errorf("topic cannot be empty")
    }

    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers:          cfg.Brokers,
        Topic:            cfg.Topic,
        Balancer:         &kafka.Hash{},
        Async:            true,
        BatchSize:        cfg.BatchSize,
        BatchTimeout:     cfg.BatchTimeout,
        RequiredAcks:     int(kafka.RequireOne),
        CompressionCodec: &snappy.Codec{},
    })

    return &Producer{writer: writer}, nil
}

func (p *Producer) SendLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
    if update.Timestamp.IsZero() {
        update.Timestamp = time.Now()
    }

    value, err := json.Marshal(update)
    if err != nil {
        return fmt.Errorf("marshal error: %v", err)
    }

    msg := kafka.Message{
        Key:   []byte(update.ID.Hex()),
        Value: value,
        Time:  update.Timestamp,
        Headers: []kafka.Header{
            {Key: "version", Value: []byte("1.0")},
            {Key: "type", Value: []byte("location-update")},
        },
    }

    if err := p.writer.WriteMessages(ctx, msg); err != nil {
        return fmt.Errorf("write error: %v", err)
    }

    log.Printf("✅ Location update sent: Agent=%s, Parcel=%s", update.ID, update.ParcelID)
    return nil
}

func (p *Producer) Close() error {
    return p.writer.Close()
}
