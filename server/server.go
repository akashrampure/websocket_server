package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var serverOnce sync.Once

type MessageHandler func(message Message)

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
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

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
		onReceive:    func(message Message) {},
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
			s.logger.Printf("Connected: %s\n", clientID)
		})

		s.OnDisconnect(func(clientID string, err error) {
			if err != nil {
				s.logger.Printf("Error disconnecting: %s: %v\n", clientID, err)
			} else {
				s.logger.Printf("Disconnected: %s\n", clientID)
			}

			s.connectionsMu.RLock()
			noClients := len(s.connections) == 0
			s.connectionsMu.RUnlock()

			if noClients {
				s.logger.Println("No clients connected, shutting down server in 5 seconds...")
				time.Sleep(5 * time.Second)
				os.Exit(0)
			}
		})

		s.OnReceive(func(message Message) {
			fmt.Printf("Received message from %s to %s: %s\n", message.Sender, message.Receiver, string(message.Data))
			if err := s.RelayMessage(message); err != nil {
				s.logger.Printf("Error relaying message: %v\n", err)
			}
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

func (s *WebSocketServer) RelayMessage(message Message) error {
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()

	receiver := message.Receiver
	conn, exists := s.connections[receiver]
	if !exists {
		return errors.New("receiver not found")
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Error marshalling message: %v\n", err)
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		s.logger.Printf("Error writing message: %v\n", err)
		return err
	}
	s.logger.Printf("Message relayed to %s\n", receiver)
	return nil
}
