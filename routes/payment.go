package routes

import (
    "go_pickup/Payment"
    "github.com/gin-gonic/gin"
)

func PaymentRoutes(r *gin.Engine) {
    pay := r.Group("/payment")
	pay.POST("/create-payment", payment.CreatePaymentOrder)
    pay.POST("/verify", payment.VerifyParcelOtp)
}