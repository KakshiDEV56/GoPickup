package middleware

import (
	"bytes"
	"encoding/json"
	"go_pickup/email"
	"go_pickup/models"
	
	"go_pickup/utils"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        claims, err := utils.ValidateJWT(tokenStr)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        c.Set("user_id", claims.UserID.Hex()) // ✅ this is what the route will use
        c.Next()
    }

}
func NotifyUser(user models.User, message string) {
 //   go twilio.SendSMS(user.Phone, message)
    go email.SendEmail(user.Email, "GoPickup Notification", message)
}

func LicenseVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			c.Abort()
			return
		}

		// Step 2: Parse the license field
		var req struct {
			License string `json:"license"`
		}
		if err := json.Unmarshal(body, &req); err != nil || req.License == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "License is required"})
			c.Abort()
			return
		}

		license := req.License
		if len(license) != 9 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid license format"})
			c.Abort()
			return
		}

		for i := 0; i < 6; i++ {
			if license[i] < 'A' || license[i] > 'Z' {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid license format"})
				c.Abort()
				return
			}
		}

		for i := 6; i < 9; i++ {
			if license[i] < '0' || license[i] > '9' {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid license format"})
				c.Abort()
				return
			}
		}

		// Step 3: Replace the body so the handler can read it again
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		c.Next()
	}
}