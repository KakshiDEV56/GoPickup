package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

type LocationUpdate struct {
    ID        string    `json:"ID"`
    ParcelID  string    `json:"ParcelID"`
    Latitude  float64   `json:"Latitude"`
    Longitude float64   `json:"Longitude"`
    Status    string    `json:"Status"`
    Timestamp time.Time `json:"Timestamp"`
}

func main() {
    apiURL := "http://localhost:8070/partner/update-location"

    agentID := "64b8a1f0a3d2f1e2b4567890"
    parcelID := "parcel123"

    // Starting point (example: New Delhi)
    lat := 28.6139
    lng := 77.2090

    for {
        // Simulate slight movement
        lat += rand.Float64()*0.001 - 0.0005
        lng += rand.Float64()*0.001 - 0.0005

        update := LocationUpdate{
            ID:        agentID,
            ParcelID:  parcelID,
            Latitude:  lat,
            Longitude: lng,
            Status:    "en route",
            Timestamp: time.Now().UTC(),
        }

        body, _ := json.Marshal(update)
        resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
        if err != nil {
            fmt.Println("Failed to send update:", err)
        } else {
            fmt.Printf("Sent: %+v, Status: %d\n", update, resp.StatusCode)
            resp.Body.Close()
        }

        time.Sleep(5 * time.Second)
    }
}
