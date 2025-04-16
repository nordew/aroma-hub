package main

import (
	"aroma-hub/internal/application"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	application.MustRun()
}
