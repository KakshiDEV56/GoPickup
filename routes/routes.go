package routes

import (
	"go_pickup/handlers"
	"go_pickup/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	auth.POST("/forgot-password",middleware.CreateRateLimiterMiddleware("5-M"),handlers.ForgotPasswordRequest)
	auth.POST("/verify-reset-otp",handlers.VerifyPasswordResetOTP)
	auth.POST("/reset-password",middleware.PasswordStrength(),handlers.ResetPassword)
	auth.POST("/parcel/:id/send-otp",middleware.CreateRateLimiterMiddleware("10-M"),handlers.GenerateAndSendOTP)
	auth.POST("/parcel",middleware.CreateRateLimiterMiddleware("10-M"),middleware.JWTAuthMiddleware(),handlers.CreateParcel)
	auth.GET("parcels",middleware.JWTAuthMiddleware(),handlers.ParcelDetails)
	auth.POST("/register",middleware.PasswordStrength(),middleware.CreateRateLimiterMiddleware("20-M"), handlers.Register)
	auth.POST("/login",middleware.CreateRateLimiterMiddleware("5-M"), handlers.Login)
	auth.POST(":id/profile-image",middleware.JWTAuthMiddleware(),handlers.UploadUserProfile)

}
