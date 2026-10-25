// Package session contains the session logic that encompasses a single game.
// Each Session maintains a list of Clients it can communicate with.
package session

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

const (
	maxPlayers            = 6
	minimumPlayersToStart = 2

	startGameDelay = time.Second * 20
)

type inboundMessage struct {
	from string
	data []byte
}

type joinRequest struct {
	playerID string
	result   chan bool
}

type playerQuery struct {
	id     string
	result chan bool
}

type registerRequest struct {
	client *Client
	result chan bool
}

// Session is a stub for a session.
type Session struct {
	logger *slog.Logger

	id uuid.UUID

	players map[string]*Client

	join       chan joinRequest
	query      chan playerQuery
	register   chan registerRequest
	unregister chan string
	inbound    chan inboundMessage
	stop       chan struct{}
	joinOrder  []string
	turnOrder  []string
	turnIdx    int
	started    bool
}

// New creates a new session.
func New(l *slog.Logger) *Session {
	s := &Session{
		id:         uuid.New(),
		players:    make(map[string]*Client),
		join:       make(chan joinRequest),
		query:      make(chan playerQuery),
		register:   make(chan registerRequest),
		unregister: make(chan string),
		inbound:    make(chan inboundMessage),
		stop:       make(chan struct{}),
	}

	s.logger = l.With("session_id", s.id)

	go s.run()

	return s
}

// GetID returns the Sessions uuid.
func (s *Session) GetID() uuid.UUID {
	return s.id
}

// HasPlayer reports if a player exists in this session.
func (s *Session) HasPlayer(id string) bool {
	result := make(chan bool)

	s.query <- playerQuery{id: id, result: result}

	return <-result
}

// Add reserves a spot in this Session for a player.
func (s *Session) Add(id string) bool {
	result := make(chan bool)

	s.join <- joinRequest{playerID: id, result: result}

	return <-result
}

// Register registers a client connection to this session.
func (s *Session) Register(c *Client) bool {
	result := make(chan bool, 1)
	s.register <- registerRequest{client: c, result: result}

	return <-result
}

// Shutdown gracefully shuts down the session.
func (s *Session) Shutdown(ctx context.Context) {
	msg := sessionv1.Envelope_builder{
		Msg: sessionv1.Message_builder{
			Msg: new("Server is shuttind down"),
		}.Build(),
	}.Build()

	s.broadcastMessage(msg)

	for _, client := range s.players {
		if client != nil {
			client.Close(ctx)
		}
	}
}

func (s *Session) run() {
	ticker := time.NewTicker(1 * time.Second)
	start := time.Now()

	for {
		select {
		case req := <-s.join:
			s.handleJoin(req)
		case q := <-s.query:
			q.result <- s.hasPlayer(q.id)
		case req := <-s.register:
			s.handleRegister(req)
		case id := <-s.unregister:
			s.handleUnregister(id)
		case msg := <-s.inbound:
			s.handleInbound(msg)
		case <-ticker.C:
			if time.Since(start) >= startGameDelay && len(s.players) >= minimumPlayersToStart && !s.started {
				ticker.Stop()

				s.startGame()
			}
		case <-s.stop:
			return
		}
	}
}

func (s *Session) handleJoin(req joinRequest) {
	s.logger.Info("Received join request", "player_id", req.playerID)

	if len(s.players) >= maxPlayers {
		s.logger.Info("Max players reached for session", "player_id", req.playerID)

		req.result <- false

		return
	}

	if s.started {
		req.result <- false

		return
	}

	s.players[req.playerID] = nil

	s.joinOrder = append(s.joinOrder, req.playerID)
	req.result <- true

	if len(s.players) == maxPlayers {
		s.startGame()
	}
}

func (s *Session) hasPlayer(id string) bool {
	_, ok := s.players[id]

	return ok
}

func (s *Session) handleRegister(req registerRequest) {
	s.logger.Info("Received register request", "player_id", req.client.id)

	if _, ok := s.players[req.client.id]; !ok {
		req.result <- false

		return
	}

	s.players[req.client.id] = req.client
	req.result <- true
}

func (s *Session) startGame() {
	s.logger.Info("Game is starting!", "players", len(s.players))
	s.started = true

	s.turnOrder = s.joinOrder
}

func (s *Session) handleUnregister(id string) {
	s.logger.Info("Received unregister request", "player_id", id)

	if c, ok := s.players[id]; ok && c != nil {
		close(c.send)
	}

	s.players[id] = nil
}

func (s *Session) handleInbound(msg inboundMessage) {
	s.logger.Info("Received inbound message request", "player_id", msg.from)

	var env sessionv1.Envelope
	if err := protojson.Unmarshal(msg.data, &env); err != nil {
		s.sendTo(msg.from, sessionv1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("malformed message"),
			}.Build(),
		}.Build())

		return
	}

	if !s.started {
		s.sendTo(msg.from, sessionv1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("error game not started"),
			}.Build(),
		}.Build())

		return
	}

	if msg.from != s.currentPlayer() {
		s.sendTo(msg.from, sessionv1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("not your turn"),
			}.Build(),
		}.Build())

		return
	}

	switch env.WhichPayload() {
	case sessionv1.Envelope_Move_case:
		to := env.GetMove().GetRoom().String()

		if env.GetMove().WhichMoveTo() == sessionv1.Move_Hall_case {
			to = env.GetMove().GetRoom().String()
		}

		s.broadcastMessage(sessionv1.Envelope_builder{
			Msg: sessionv1.Message_builder{
				Msg: new("player " + env.GetMove().GetPlayerId() + "has moved to " + to),
			}.Build(),
		}.Build())

		return

	default:
	}
}

func (s *Session) sendTo(id string, env *sessionv1.Envelope) {
	s.logger.Info("Sending message", "player_id", id, "message_type", env.WhichPayload().String())

	c, ok := s.players[id]
	if !ok || c == nil {
		return
	}

	data, _ := protojson.Marshal(env)

	select {
	case c.send <- data:
	default:
	}
}

func (s *Session) currentPlayer() string {
	if s.turnIdx < 0 || s.turnIdx >= len(s.turnOrder) {
		return ""
	}

	return s.turnOrder[s.turnIdx]
}

func (s *Session) broadcastMessage(env *sessionv1.Envelope) {
	s.logger.Info("Broadcasting message", "message_type", env.WhichPayload().String())

	for id := range s.players {
		s.sendTo(id, env)
	}
}
