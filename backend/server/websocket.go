package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/TeamHermes-jhu/clueless/backend/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
	CheckOrigin: func(_ *http.Request) bool {
		return true // accept any origin for now
	},
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	s.logger.InfoContext(r.Context(), "Initiating websocket handshake")

	sessionID, playerID, err := getURLValues(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	sess, ok := s.sessions[sessionID]
	s.mu.Unlock()

	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Error upgrading session", http.StatusInternalServerError)

		return
	}

	client := session.NewClient(playerID, conn, sess)

	if !sess.Register(client) {
		_ = conn.Close()

		return
	}

	go client.Write()

	client.Read()
}

func getURLValues(r *http.Request) (uuid.UUID, string, error) {
	sessionID, err := uuid.Parse(r.URL.Query().Get("session"))
	if err != nil {
		return uuid.Nil, "", err
	}

	playerID := r.URL.Query().Get("player")
	if playerID == "" {
		return uuid.Nil, "", err
	}

	return sessionID, playerID, nil
}
