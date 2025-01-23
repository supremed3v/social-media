package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func GetString(key, fallback string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	val := os.Getenv(key)

	if val == "" {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	val := os.Getenv(key)

	valAsInt, err := strconv.Atoi(val)

	if err != nil {
		return fallback
	}

	return valAsInt

}

func GetBool(key string, fallback bool) bool {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	val := os.Getenv(key)

	boolVal, err := strconv.ParseBool(val)

	if err != nil {
		return fallback
	}

	return boolVal
}
