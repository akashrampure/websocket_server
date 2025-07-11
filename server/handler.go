package main

import (
	"log"
	"net/http"
	"time"
)

type Message struct {
	ClientID string `json:"client_id"`
	Data     []byte `json:"data"`
}

func (s *WebSocketServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	clientID := r.RemoteAddr

	s.connectionsMu.Lock()
	s.connections[clientID] = conn
	s.connectionsMu.Unlock()

	if s.onConnect != nil {
		s.onConnect(clientID)
	}

	defer func() {
		s.connectionsMu.Lock()
		delete(s.connections, clientID)
		s.connectionsMu.Unlock()

		if s.onDisconnect != nil {
			s.onDisconnect(clientID, err)
		}

		conn.Close()
	}()

	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read error from %s: %v", clientID, err)
			break
		}

		if s.onReceive != nil {
			s.onReceive(clientID, msg)
		}
		log.Printf("Client %s sent message: %s", clientID, string(msg))
	}
}
