package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func SetUpCors() []string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

	if allowedOrigins == "" {
		log.Fatal("CORS_ALLOWED_ORIGINS is not set")
	}

	return strings.Split(allowedOrigins, ",")
}
