# Go Image Upload Web Application

This is a simple Go web application that allows users to upload images along with associated properties (in JSON format). Users can also view the uploaded images.

## Prerequisites

Before running this application, you need to have the following installed:

- Go (v1.16 or higher)

## Getting Started

1. Clone this repository:
```bash
   git clone https://github.com/your-username/your-repo.git
   cd your-repo
```
2. Install dependencies: 
```bash
go mod tidy
```
3. Access the web application:
Open a web browser and navigate to `http://localhost:8080` (or the configured port) to access the image upload interface.

### Usage
Upload Images: On the homepage, you can select an image file and provide associated properties in JSON format.
View Images: Navigate to /images/{imageID} to view a specific image by its ID. Replace {imageID} with the actual ID of the image.

## Project Structure

    main.go: The main entry point of the Go application.
    templates/: Contains the HTML templates for the web pages.
    README.md: This file.

License

This project is licensed under the MIT License. See the LICENSE file for details.
