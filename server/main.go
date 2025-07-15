package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"websocket/utils"
)

func main() {
	err := utils.LoadEnv("/home/akash/Downloads/wsnew/.env")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	path := os.Getenv("WEBSOCKET_PATH")
	if path == "" {
		path = "/ws"
	}

	file, err := os.OpenFile("/home/akash/Downloads/wsnew/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	config := NewWebSocketConfig(port, path)
	server := NewWebSocketServer(config, logger)

	go server.Start()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	logger.Println("Shutting down the server...")
}
