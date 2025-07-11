package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sub := NewSubscribeWS("ws", "localhost:8080", "/ws")
	sub.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			sub.SendMessage([]byte("ping from client"))
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	log.Println("Shutting down the client...")
	sub.Close()
}
