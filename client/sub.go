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

type MessageStruct struct {
	ClientID string `json:"client_id"`
	Data     []byte `json:"data"`
}

type SubscribeWS struct {
	Scheme  string
	Host    string
	Path    string
	conn    *websocket.Conn
	onClose chan os.Signal
}

func NewSubscribeWS(scheme, host, path string) *SubscribeWS {
	return &SubscribeWS{
		Scheme:  scheme,
		Host:    host,
		Path:    path,
		onClose: make(chan os.Signal),
	}
}

func (s *SubscribeWS) Start() {
	go func() {
		for i := 0; i < maxRetries; i++ {
			err := s.connectAndListen()
			if err != nil {
				log.Printf("Connection lost: %v. Reconnecting (%d/%d)...", err, i+1, maxRetries)
				time.Sleep(reconnectInterval)
			} else {
				return
			}
		}
		log.Println("Max retries exceeded. Exiting...")
		os.Exit(1)
	}()

}

func (s *SubscribeWS) OnReceive() error {
	_, msg, err := s.conn.ReadMessage()
	if err != nil {
		log.Println("Read error:", err)
		return err
	}
	log.Println(string(msg))
	return nil
}

func (s *SubscribeWS) connectAndListen() error {
	url := s.Scheme + "://" + s.Host + s.Path
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("Initial connection failed: %v", err)
		return err
	}
	s.conn = ws
	log.Printf("Connected to %s", url)

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

func (s *SubscribeWS) SendMessage(msg []byte) {
	if s.conn != nil {
		message := MessageStruct{
			ClientID: "123",
			Data:     msg,
		}
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			log.Println("JSON marshal error:", err)
			return
		}
		if err := s.conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
			log.Println("Write error:", err)
		}
	}
}
