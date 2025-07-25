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

type ConsumerConfig struct {
    Brokers   []string
    Topic     string
    GroupID   string
    RedisAddr string
}

func StartLocationConsumer(cfg ConsumerConfig) error {
    log.Printf("DEBUG: Kafka consumer received config: Brokers=%v, Topic=%s, GroupID=%s, RedisAddr=%s", cfg.Brokers, cfg.Topic, cfg.GroupID, cfg.RedisAddr) // <-- ADD THIS LINE
    rdb := redis.NewClient(&redis.Options{
        Addr:       cfg.RedisAddr,
        MaxRetries: 3,
    })
    defer rdb.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := rdb.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis connection failed: %v", err)
    }

    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:      cfg.Brokers,
        Topic:        cfg.Topic,
        GroupID:      cfg.GroupID,
        StartOffset:  kafka.FirstOffset,
        MinBytes:     10e3,
        MaxBytes:     10e6,
        MaxWait:      3 * time.Second,
    })
    defer reader.Close()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    done := make(chan bool)

    go func() {
        for {
            readCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            msg, err := reader.ReadMessage(readCtx)
            cancel()

            if err != nil {
                log.Printf("❌ Kafka read error: %v", err)
                select {
                case <-done:
                    return
                default:
                    continue
                }
            }

            var location models.LocationUpdate
            if err := json.Unmarshal(msg.Value, &location); err != nil {
                log.Printf("❌ JSON unmarshal error: %v", err)
                continue
            }

            if location.ID.IsZero() || len(location.ID.Hex()) != 24 {
                log.Printf("❌ Invalid Agent ID: %v", location.ID)
                continue
            }

            key := fmt.Sprintf("location:%s:%s", location.ID.Hex(), location.ParcelID)
            if err := rdb.Set(context.Background(), key, msg.Value, 24*time.Hour).Err(); err != nil {
                log.Printf("❌ Redis set error: %v", err)
                continue
            }

            channel := fmt.Sprintf("realtime-locations:%s", location.ParcelID)
            if err := rdb.Publish(context.Background(), channel, msg.Value).Err(); err != nil {
                log.Printf("❌ Redis publish error: %v", err)
                continue
            }

            log.Printf("✅ Location stored & published: Agent=%s, Parcel=%s", location.ID, location.ParcelID)
        }
    }()

    <-sigChan
    done <- true
    return nil
}
