package main

import (
	"log"
	"os"
	"strconv"
	"websocket/utils"
)

func main() {
	err := utils.LoadEnv("/home/akash/Downloads/wsnew/.env")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	port := getEnv("PORT", "8080")
	path := getEnv("WEBSOCKET_PATH", "/ws")

	file, err := os.OpenFile("/home/akash/Downloads/wsnew/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	config := NewWebSocketConfig(port, path)
	server := NewWebSocketServer(config, logger)

	server.Start()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueInt
}
