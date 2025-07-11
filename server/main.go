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

	config := NewWebSocketConfig(port, path)
	server := NewWebSocketServer(config)

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

	log.Println("Shutting down the server...")
}
