package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// loads environment variables from .env to ENV
func LoadEnvironmentVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
}

// retrieves the environment variable from .env
func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("%s not found in environment variables", key)
	}
	return value
}

// checks if debugging output is enabled
func DebugEnabled() bool {
	if debug, exists := os.LookupEnv("DEBUG"); exists && debug == "true" {
		return true
	}
	return false
}
