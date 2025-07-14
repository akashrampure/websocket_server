package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv(path string) {
	err := godotenv.Load(path)
	if err != nil {
		log.Println("Error loading .env file")
	}
}
