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
}

func NewWebSocketServer(config WebSocketConfig) *WebSocketServer {
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
			log.Printf("Client connected: %v", clientID)
		})

		s.OnDisconnect(func(clientID string, err error) {
			log.Printf("Client disconnected: %s", clientID)
			if len(s.connections) == 0 {
				log.Println("No clients connected, shutting down server in 3 seconds...")
				time.Sleep(3 * time.Second)
				os.Exit(0)
			}
		})

		s.OnReceive(func(clientID string, msg []byte) {
			var message Message
			err := json.Unmarshal(msg, &message)
			if err != nil {
				log.Printf("Error marshalling message: %v", err)
				return
			}
			log.Printf("Client %s sent message: %s", message.ClientID, string(message.Data))
		})

		log.Printf("WebSocket server started on %s", s.Config.Port)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Println("server closed:", err)
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

func (s *WebSocketServer) Broadcast(msg []byte) error {
	s.connectionsMu.RLock()
	defer s.connectionsMu.RUnlock()
	for id, conn := range s.connections {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("broadcast to %s failed: %v", id, err)
			conn.Close()
			delete(s.connections, id)
		}
	}
	return nil
}
