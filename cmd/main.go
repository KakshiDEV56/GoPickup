package main

import (
	"go_pickup/config"
	"go_pickup/kafka"
	"go_pickup/routes"
	"go_pickup/twilio" // ✅ Import your Twilio package
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("App Starting - Version: %s", time.Now().Format("2006-01-02 15:04:05")) // <-- ADD THIS LINE

	config.LoadEnv()
	config.ConnectMongo() // 🔗 Connect MongoDB

	twilio.Init()         // ✅ Initialize Twilio client here
	go func() {
        err := kafka.StartLocationConsumer(kafka.ConsumerConfig{
            Brokers:   []string{"localhost:9092"},
            Topic:     "location-updates",
            GroupID:   "tracker-service",
            RedisAddr: "localhost:6379",
        })
        if err != nil {
            log.Fatalf("Kafka consumer error: %v", err)
        }
    }()
	r := gin.Default()
	routes.PaymentRoutes(r)
	routes.DeliveryPartnerRoute(r)
	routes.AuthRoutes(r)

	port := config.GetEnv("PORT")
	r.Run(":" + port)
}
