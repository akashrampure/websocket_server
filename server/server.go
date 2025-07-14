package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var serverOnce sync.Once

type MessageHandler func(clientID string, msg []byte)

type WebSocketConfig struct {
	Port            string
	Path            string
	ReadBufferSize  int
	WriteBufferSize int
	AllowedOrigins  []string
}

func NewWebSocketConfig(port, path string) WebSocketConfig {
	return WebSocketConfig{
		Port:            port,
		Path:            path,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		AllowedOrigins:  []string{"*"},
	}
}

type WebSocketServer struct {
	Config        WebSocketConfig
	Upgrader      *websocket.Upgrader
	connections   map[string]*websocket.Conn
	connectionsMu sync.RWMutex
	onReceive     MessageHandler
	onConnect     func(clientID string)
	onDisconnect  func(clientID string, err error)
	logger        *log.Logger
}

func NewWebSocketServer(config WebSocketConfig, logger *log.Logger) *WebSocketServer {

	return &WebSocketServer{
		Config: config,
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections:  make(map[string]*websocket.Conn),
		onReceive:    func(clientID string, msg []byte) {},
		onConnect:    func(clientID string) {},
		onDisconnect: func(clientID string, err error) {},
		logger:       logger,
	}
}

func (s *WebSocketServer) Start() {
	serverOnce.Do(func() {
		mux := http.NewServeMux()
		mux.HandleFunc(s.Config.Path, s.wsHandler)
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", s.Config.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		s.OnConnect(func(clientID string) {
			s.logger.Printf("Client connected: %v", clientID)
		})

		s.OnDisconnect(func(clientID string, err error) {
			s.logger.Printf("Client disconnected: %s", clientID)
			if len(s.connections) == 0 {
				s.logger.Println("No clients connected, shutting down server in 3 seconds...")
				time.Sleep(3 * time.Second)
				os.Exit(0)
			}
		})

		s.OnReceive(func(clientID string, msg []byte) {
			var message Message
			err := json.Unmarshal(msg, &message)
			if err != nil {
				s.logger.Printf("Error marshalling message: %v", err)
				return
			}
			fmt.Printf("Received message from %s to %s: %s\n", message.Sender, message.Receiver, string(message.Data))
			s.RelayMessage(message)
		})

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				s.logger.Println("server closed:", err)
			}
		}()
	})
}

func (s *WebSocketServer) OnReceive(handler MessageHandler) {
	s.onReceive = handler

}

func (s *WebSocketServer) OnConnect(fn func(clientID string)) {
	s.onConnect = fn
}

func (s *WebSocketServer) OnDisconnect(fn func(clientID string, err error)) {
	s.onDisconnect = fn
}

func (s *WebSocketServer) Broadcast(message Message) error {
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()

	receiver := message.Receiver

	conn, exists := s.connections[receiver]
	if !exists {
		return fmt.Errorf("client %s not found", receiver)
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Error marshalling message: %v", err)
		return err
	}
	conn.WriteMessage(websocket.TextMessage, msgBytes)

	return nil
}

func (s *WebSocketServer) RelayMessage(message Message) error {
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()

	receiver := message.Receiver
	conn, exists := s.connections[receiver]
	if !exists {
		return fmt.Errorf("client %s not found", receiver)
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Error marshalling message: %v", err)
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		s.logger.Printf("Error writing message: %v", err)
		return err
	}
	s.logger.Printf("Message relayed to %s", receiver)
	return nil
}
