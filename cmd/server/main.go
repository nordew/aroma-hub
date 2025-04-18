package main

import (
	"aroma-hub/internal/application"
	"log"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

// @title						Aroma-Hub API
// @version					1.0
// @description				dAPI documentation.
// @BasePath					/api/v1
func main() {
	application.MustRun()
}
