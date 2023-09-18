package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type Image struct {
	ID   string `json:"id"`
	Data string `json:"data"`
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

	fmt.Printf("Server is running on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Server error: %s\n", err)
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
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		handleError(w, "Failed to read the image data", http.StatusInternalServerError)
		return
	}

	imageData := base64.StdEncoding.EncodeToString(data)

	imageID++
	id := fmt.Sprintf("image_%d", imageID)

	image := Image{
		ID:   id,
		Data: imageData,
	}
	imagesLock.Lock()
	images[id] = image
	imagesLock.Unlock()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": "%s"}`, id)
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
	w.Write(imageData)
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}