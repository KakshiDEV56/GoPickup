package main

import (
	"github.com/gin-gonic/gin"
	"go_pickup/config"
	"go_pickup/routes"
	"go_pickup/twilio" // ✅ Import your Twilio package
)

func main() {
	config.LoadEnv()
	config.ConnectMongo() // 🔗 Connect MongoDB

	twilio.Init()         // ✅ Initialize Twilio client here

	r := gin.Default()
	routes.PaymentRoutes(r)
	routes.DeliveryPartnerRoute(r)
	routes.AuthRoutes(r)

	port := config.GetEnv("PORT")
	r.Run(":" + port)
}
