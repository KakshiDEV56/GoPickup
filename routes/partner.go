package routes

import (
	"go_pickup/handlers"
	//"go_pickup/kafka"
	"go_pickup/middleware"

	"github.com/gin-gonic/gin"
)

        func DeliveryPartnerRoute(c *gin.Engine){
AgentAuth:=c.Group("/partner")
AgentAuth.POST("/send-location",)
AgentAuth.POST("/register",middleware.PasswordStrength(),middleware.CreateRateLimiterMiddleware("5-M"), middleware.LicenseVerification(), handlers.DeliveryPartnerRegistration)
AgentAuth.POST("/Parcel/status",middleware.JWTAuthMiddleware(),handlers.UpdateParcelStatusByAgent)
AgentAuth.POST("/login",middleware.CreateRateLimiterMiddleware("5-M"),handlers.DeliveryPartnerLogin)
AgentAuth.GET("/ParcelDetails",middleware.JWTAuthMiddleware(),handlers.ParcelDetails)
//AgentAuth.POST("/location/update",handler.PostLocation)
AgentAuth.POST(":id/profile-image",middleware.JWTAuthMiddleware(),handlers.UploadAgentProfile)
AgentAuth.POST("/forgot-password",middleware.CreateRateLimiterMiddleware("5-M"),handlers.ForgotAgentPasswordRequest)
AgentAuth.POST("/verify-reset-otp",handlers.VerifyAgentPasswordResetOTP)
AgentAuth.POST("/reset-password",middleware.PasswordStrength(),handlers.ResetAgentPassword)
}