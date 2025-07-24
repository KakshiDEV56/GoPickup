package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go_pickup/models"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Generates a random float64 between min and max
func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func main() {
	//rand.Seed(time.Now().UnixNano())

	agentID := "AGENT_SIM_" + strconv.Itoa(rand.Intn(1000))
	parcelID := "PARCEL_SIM_" + strconv.Itoa(rand.Intn(1000))

	// Mumbai region bounds
	latMin, latMax := 19.00, 19.30
	lngMin, lngMax := 72.80, 73.00

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("🚚 Simulator started for agent %s, parcel %s", agentID, parcelID)

	client := &http.Client{Timeout: 5 * time.Second}
	apiURL := "http://localhost:8080/api/location" // Change if your API runs elsewhere

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lat := randomFloat(latMin, latMax)
				lng := randomFloat(lngMin, lngMax)
				update := models.LocationUpdate{
					ParcelID: primitive.ObjectID{},
					Latitude:  lat,
					Longitude: lng,
					Timestamp: time.Now(),
					Status:    "IN_TRANSIT",
				}
					
					//Speed:     randomFloat(20, 40),

				body, _ := json.Marshal(update)
				req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("❌ Failed to send: %v", err)
				} else {
					resp.Body.Close()
					log.Printf("✅ Sent location: lat=%.5f, lng=%.5f", lat, lng)
				}

				time.Sleep(2 * time.Second)
			}
		}
	}()

	<-sigChan
	log.Println("🛑 Simulator shutting down...")
	cancel()
}
