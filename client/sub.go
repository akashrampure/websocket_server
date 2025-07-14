package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var (
	reconnectInterval = 2 * time.Second
	maxRetries        = 5
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Data     []byte `json:"data"`
}

type SubscribeWS struct {
	Scheme   string
	Host     string
	Path     string
	ClientID string
	conn     *websocket.Conn
	logger   *log.Logger
}

func NewSubscribeWS(scheme, host, path, clientID string, logger *log.Logger) *SubscribeWS {
	return &SubscribeWS{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		ClientID: clientID,
		logger:   logger,
	}
}

func (s *SubscribeWS) Start() {
	go func() {
		for i := 0; i < maxRetries; i++ {
			err := s.connectAndListen()
			if err != nil {
				s.logger.Printf("Connection lost: %v. Reconnecting (%d/%d)...", err, i+1, maxRetries)
				time.Sleep(reconnectInterval)
			} else {
				return
			}
		}
		s.logger.Println("Max retries exceeded. Exiting...")
		os.Exit(1)
	}()

}

func (s *SubscribeWS) OnReceive(clientID string) error {
	_, msg, err := s.conn.ReadMessage()
	if err != nil {
		s.logger.Println("Read error:", err)
		return err
	}
	var message Message
	err = json.Unmarshal(msg, &message)
	if err != nil {
		s.logger.Println("JSON unmarshal error:", err)
		return err
	}
	if message.Receiver == clientID {
		fmt.Printf("Received message from %s to %s: %s\n", message.Sender, message.Receiver, string(message.Data))
		return nil
	}
	return nil
}

func (s *SubscribeWS) connectAndListen() error {
	url := s.Scheme + "://" + s.Host + s.Path

	headers := http.Header{}
	headers.Add("Client-ID", s.ClientID)

	ws, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		s.logger.Printf("Initial connection failed: %v", err)
		return err
	}
	s.conn = ws
	s.logger.Printf("Connected to %s", url)

	s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.conn.SetPongHandler(func(appData string) error {
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	defer s.conn.Close()

	for {
		if err := s.OnReceive(s.ClientID); err != nil {
			return err
		}
	}
}

func (s *SubscribeWS) Close() {
	if s.conn != nil {
		_ = s.conn.Close()
	}
}

func (s *SubscribeWS) SendMessage(message Message) {
	if s.conn != nil {
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			s.logger.Println("JSON marshal error:", err)
			return
		}
		if err := s.conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
			s.logger.Println("Write error:", err)
		}
	}
}
