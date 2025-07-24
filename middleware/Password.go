package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

func PasswordStrength() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the original request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			return
		}

		// Restore the body for future handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse into a map to extract password
		var requestBody map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		password, ok := requestBody["password"].(string)
		if !ok || !isStrongPassword(password) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 8 characters and include at least 1 uppercase, 1 lowercase, 1 digit, and 1 special character",
			})
			return
		}

		// Re-assign the request body again before proceeding to the next handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString
	hasLower := regexp.MustCompile(`[a-z]`).MatchString
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString
	hasSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/{}\\[\]-]`).MatchString

	return hasUpper(password) &&
		hasLower(password) &&
		hasDigit(password) &&
		hasSpecial(password)
}
