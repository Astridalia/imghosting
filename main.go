package main

import (
	"encoding/base64"
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

var cassandraClient *database.MyCassandraClient

func main() {
	// Initialize Cassandra client
	cassandraClient = database.InitCassandra()

	router := gin.Default()
	router.StaticFile("/", "./templates")
	// Define routes
	router.POST("/upload", handleUpload)

	// Define the /images route for viewing images by ID
	router.GET("/images/:id", handleGetImage)

	// Define a wildcard route to serve static files, but make sure it comes after specific routes

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

	image, err := cassandraClient.GetImageFromCassandra(id)

	if err != nil {
		// Check if the error is due to "not found"
		if err == database.ErrImageNotFound {
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

	c.Data(http.StatusOK, "image/jpeg", imageData)
}
