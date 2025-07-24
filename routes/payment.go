package routes

import (
    "go_pickup/handlers"
    "github.com/gin-gonic/gin"
)

func PaymentRoutes(r *gin.Engine) {
    pay := r.Group("/payment")
	pay.POST("/create-payment", handlers.CreatePaymentOrder)
    pay.POST("/verify", handlers.VerifyParcelOtp)
}