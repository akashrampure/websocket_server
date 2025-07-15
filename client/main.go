package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"websocket/utils"
)

func main() {
	err := utils.LoadEnv("/home/akash/Downloads/wsnew/.env")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	scheme := getEnv("SCHEME", "ws")
	host := getEnv("HOST", "localhost")
	port := getEnv("SERVER_PORT", "8080")
	path := getEnv("SERVER_PATH", "/ws")
	maxRetries := getEnvInt("MAX_RETRIES", 5)
	reconnectInterval := getEnvInt("RECONNECT_INTERVAL", 2)

	config := NewSubscribeConfig(scheme, host, port, path, maxRetries, reconnectInterval)

	file, err := os.OpenFile("/home/akash/Downloads/wsnew/client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	if len(os.Args) < 3 {
		log.Fatal("Usage: ./client <clientID> <receiver>")
	}

	clientID := os.Args[1]
	receiver := os.Args[2]

	sub := NewSubscribeWS(config, clientID, logger)

	sub.SetOnReceive(func(msg Message) {
		fmt.Printf("Received message from %s to %s: %s\n", msg.Sender, msg.Receiver, string(msg.Data))
	})

	sub.Start()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			sub.SendMessage(receiver, "Hello, I am "+clientID+"!")
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	logger.Println("Shutting down the client...")
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
