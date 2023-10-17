package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type Image struct {
	ID         string `json:"id"`
	Data       string `json:"data"`
	Properties string `json:"properties"`
}

var (
	images     = make(map[string]Image)
	imagesLock sync.RWMutex
	imageID    int
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./templates")))
	http.HandleFunc("/upload", uploadImage)
	http.HandleFunc("/images/", getImageByID)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s...\n", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("Server error: %s\n", err)
	}
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		handleError(w, "Failed to read the image", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Server Error: %s", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		handleError(w, "Failed to read the image data", http.StatusInternalServerError)
		return
	}

	imageData := base64.StdEncoding.EncodeToString(data)

	imageID++
	id := fmt.Sprint(imageID)

	jsonStr := r.FormValue("properties")
	image := Image{
		ID:         id,
		Data:       imageData,
		Properties: jsonStr,
	}

	imagesLock.Lock()
	images[id] = image
	imagesLock.Unlock()

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(image); err != nil {
		log.Printf("Server Error: %s", err)
	}
}

func getImageByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/images/"):]

	imagesLock.RLock()
	image, ok := images[id]
	imagesLock.RUnlock()

	if !ok {
		handleError(w, "Image not found", http.StatusNotFound)
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(image.Data)
	if err != nil {
		handleError(w, "Failed to decode image data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(imageData); err != nil {
		log.Printf("Server Error: %s", err)
	}
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
