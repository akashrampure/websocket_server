package main

import (
	"encoding/json"
	"log"
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
	Scheme string
	Host   string
	Path   string
	conn   *websocket.Conn
	logger *log.Logger
}

func NewSubscribeWS(scheme, host, path string, logger *log.Logger) *SubscribeWS {
	return &SubscribeWS{
		Scheme: scheme,
		Host:   host,
		Path:   path,
		logger: logger,
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

func (s *SubscribeWS) OnReceive() error {
	_, msg, err := s.conn.ReadMessage()
	if err != nil {
		s.logger.Println("Read error:", err)
		return err
	}
	s.logger.Println(string(msg))
	return nil
}

func (s *SubscribeWS) connectAndListen() error {
	url := s.Scheme + "://" + s.Host + s.Path
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
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
		if err := s.OnReceive(); err != nil {
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
