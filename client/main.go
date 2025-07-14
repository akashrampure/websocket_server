package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	file, err := os.OpenFile("client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	sub := NewSubscribeWS("ws", "localhost:8080", "/ws", logger)
	sub.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			sub.SendMessage("123", []byte("ping from client"))
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	logger.Println("Shutting down the client...")
	sub.Close()
}
