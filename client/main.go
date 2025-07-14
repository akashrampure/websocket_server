package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	file, err := os.OpenFile("/home/akash/Downloads/wsnew/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	clientID := os.Args[1]
	receiver := os.Args[2]

	sub := NewSubscribeWS("ws", "localhost:8080", "/ws", clientID, logger)
	sub.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			sub.SendMessage(
				Message{
					Sender:   clientID,
					Receiver: receiver,
					Data:     []byte("Hello, I am client " + clientID + "!"),
				},
			)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	logger.Println("Shutting down the client...")
	sub.Close()
}
