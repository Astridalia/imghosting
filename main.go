package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/astridalia/tinyrpg/database"
	"github.com/astridalia/tinyrpg/models"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	cassandraClient *database.MyCassandraClient
	redisClient     *database.MyRedisClient
)

func main() {
	// Initialize Cassandra client
	cassandraClient = database.InitCassandra()

	// Initialize Redis client
	redisClient = database.InitRedisClient()

	router := gin.Default()

	// Use Redis caching middleware for all routes
	router.Use(redisClient.CacheMiddleware())

	// Serve static files
	router.StaticFile("/", "./templates")

	// Define routes
	router.POST("/upload", handleUpload)
	router.GET("/images/:id", handleGetImage)

	// Get the port from the environment or use the default port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	address := ":" + port
	err := router.Run(address)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleUpload(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the image"})
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the image data"})
		return
	}

	id, err := gocql.RandomUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate a UUID"})
		return
	}

	imageData := base64.StdEncoding.EncodeToString(data)
	jsonStr := c.PostForm("properties")

	image := &models.Image{ID: id.String(), Data: imageData, Properties: jsonStr}
	if err := cassandraClient.InsertImage(image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the image to Cassandra"})
		return
	}

	imageURL := fmt.Sprintf("/images/%s", id)
	c.Header("Location", imageURL)
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func handleGetImage(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	// Try to get the image data from Redis cache
	cachedImage, err := redisClient.Get(c, id)
	if err == nil {
		imageData, err := base64.StdEncoding.DecodeString(cachedImage)
		if err == nil {
			c.Data(http.StatusOK, "image/jpeg", imageData)
			return
		}
	}

	// Data not found in cache, proceed with Cassandra
	image, err := cassandraClient.GetImageFromCassandra(id)

	if err != nil {
		// Check if the error is due to "not found"
		if errors.Is(err, database.ErrImageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the image"})
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(image.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image data"})
		return
	}

	// Cache the image data in Redis, but only if successfully retrieved from Cassandra
	err = redisClient.Set(c, id, image.Data)
	if err != nil {
		// Handle the error (e.g., log it)
	}

	c.Data(http.StatusOK, "image/jpeg", imageData)
}
