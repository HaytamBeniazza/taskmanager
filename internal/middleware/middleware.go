package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs information about each request
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate request time
		duration := time.Since(startTime)

		// Log request details
		log.Printf(
			"[%s] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			duration,
		)
	}
}

// Recovery middleware handles panics and recovers
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique request ID (in a real app, use a proper UUID library)
		requestID := time.Now().UnixNano()

		// Set the request ID in the context
		c.Set("RequestID", requestID)

		// Add it as a response header
		c.Header("X-Request-ID", time.Now().Format("20060102150405.000"))

		c.Next()
	}
}

// RateLimiter middleware limits request rates (simple implementation)
func RateLimiter() gin.HandlerFunc {
	// In a real app, use a proper rate limiting library and store
	// This is just a simple placeholder implementation
	return func(c *gin.Context) {
		// For this example, we're allowing all requests
		c.Next()
	}
}

// SimpleAuth middleware provides basic authentication
// Note: This is a placeholder. For real applications, use proper auth middleware
func SimpleAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API token in header
		token := c.GetHeader("Authorization")
		if token == "" {
			// For development, allow unauthenticated access
			c.Next()
			return
		}

		// In a real app, validate the token against a database or JWT
		// For this example, allow all tokens
		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// APIHealthCheck middleware adds a health check endpoint
func APIHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status": "up",
				"time":   time.Now(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// APIVersioning middleware adds version info to response headers
func APIVersioning(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-API-Version", version)
		c.Next()
	}
}

// RequestSizeLimiter sets a maximum request size
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
