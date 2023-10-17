package database

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"time"
)

func InitRedisClient() *MyRedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	return &MyRedisClient{client: rdb}
}

type MyRedisClient struct {
	client *redis.Client
}

func (r *MyRedisClient) Get(c *gin.Context, key string) (string, error) {
	// Wrap the Redis Get method if you want to use it in your middleware
	val, err := r.client.Get(c, key).Result()
	if err == redis.Nil {
		return "", nil // Redis cache miss
	}
	return val, err
}

func (r *MyRedisClient) Set(c *gin.Context, key, data string) error {
	err := r.client.Set(c, key, data, 1*time.Hour).Err()
	return err
}

func (r *MyRedisClient) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.URL.Path

		// Try to get the data from Redis cache
		cachedData, err := r.Get(c, key)
		if err == nil {
			// Data found in cache, return it
			c.String(http.StatusOK, cachedData)
			return
		}

		// Create a custom response writer to capture the response
		writer := NewResponseWriter(c.Writer)

		// Continue with the request
		c.Writer = writer
		c.Next()

		// Cache the response only if the status code is OK
		if c.Writer.Status() == http.StatusOK {
			responseBody := writer.Body()
			err := r.Set(c, key, responseBody)
			if err != nil {
				// Handle the error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache the response"})
			}
		}
	}
}

type CustomResponseWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func NewResponseWriter(w gin.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: w,
		buffer:         bytes.NewBuffer(nil),
	}
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *CustomResponseWriter) Body() string {
	return w.buffer.String()
}
