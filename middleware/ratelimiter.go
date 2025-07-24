package middleware
// i will try to implement the leaky bucket algorithim after some time so that in order to ehance the user experience this method that i am currently applying is good for understanding but will not work as it is permananetly banning the ip 
import (
	"context"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

func CreateRateLimiterMiddleware(rate string) gin.HandlerFunc {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Println("✅ New Redis connection established")
			return nil
		},
		MaxRetries: 3,
		MaxRetryBackoff: time.Second,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Cannot connect to Redis: %v", err)
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limiter service unavailable"})
		}
	}
	log.Println("✅ Redis ping successful")

	parsedRate, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		log.Printf("❌ Invalid rate format: %v", err)
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter configuration error"})
		}
	}

	store, err := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:   "rate_limiter",
		MaxRetry: 3,
	})
	if err != nil {
		log.Printf("❌ Failed to create Redis store for rate limiter: %v", err)
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter store error"})
		}
	}

	rateLimiter := limiter.New(store, parsedRate)
	middleware := mgin.NewMiddleware(rateLimiter)

	return func(c *gin.Context) {
		// Check Redis connection before each request
		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			log.Printf("❌ Redis connection lost: %v", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limiter temporarily unavailable"})
			return
		}

		clientIP := c.ClientIP()
		log.Printf("🔥 RateLimiter triggered for IP: %s", clientIP)
		middleware(c)
	}
}
