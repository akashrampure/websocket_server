package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Data     []byte `json:"data"`
}

func (s *WebSocketServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Println("upgrade error:", err)
		return
	}

	clientID := r.Header.Get("Client-ID")

	s.connectionsMu.Lock()
	s.connections[clientID] = conn
	s.connectionsMu.Unlock()

	if s.onConnect != nil {
		s.onConnect(clientID)
	}

	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					s.logger.Printf("ping error to %s: %v", clientID, err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	defer func() {
		close(done)

		s.connectionsMu.Lock()
		delete(s.connections, clientID)
		s.connectionsMu.Unlock()

		if s.onDisconnect != nil {
			s.onDisconnect(clientID, err)
		}

		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.logger.Printf("read error from %s: %v", clientID, err)
			break
		}

		if s.onReceive != nil {
			s.onReceive(clientID, msg)
		}
	}
}
