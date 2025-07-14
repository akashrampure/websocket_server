package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"websocket/utils"
)

func main() {
	utils.LoadEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	path := os.Getenv("PATH")
	if path == "" {
		path = "/ws"
	}

	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	config := NewWebSocketConfig(port, path)
	server := NewWebSocketServer(config, logger)

	go server.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			server.Broadcast([]byte("Hello, I am server!"))
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	logger.Println("Shutting down the server...")
}
