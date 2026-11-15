// Package session contains the session logic that encompasses a single game.
// Each Session maintains a list of Clients it can communicate with.
package session

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/TeamHermes-jhu/clueless/backend/glm"
	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

const (
	maxPlayers            = 6
	minimumPlayersToStart = 3

	startGameDelay = time.Second * 120

	gameFullAndAllSelectedMessage = "Game is full and all characters are selected. Starting game."
	gameTimerFmt                  = "%d/6 players. Game starts in %ds"
	playerLocationFmt             = "player %s location: %s"
	playerHandFmt                 = "Your hand: suspects %v, rooms %v, weapons %v"

	suggestPendingFmt = "waiting for %s to disprove your suggestion"
	suggestNoneFmt    = "no one could disprove your suggestion"
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
	started    bool
	logic      *glm.GameLogicManager
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
		logic:      glm.New(),
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
	s.broadcastMessage("Server is shutting down")

	for _, client := range s.players {
		if client != nil {
			client.Close(ctx)
		}
	}
}

func (s *Session) run() {
	var (
		ticker  *time.Ticker
		tickerC <-chan time.Time
		start   time.Time
		end     time.Time
	)

	for {
		select {
		case req := <-s.join:
			s.handleJoin(req)

			if ticker == nil && len(s.players) >= minimumPlayersToStart {
				start = time.Now()
				end = start.Add(startGameDelay)
				ticker = time.NewTicker(time.Second)
				tickerC = ticker.C
			}
		case q := <-s.query:
			q.result <- s.hasPlayer(q.id)
		case req := <-s.register:
			s.handleRegister(req)
		case id := <-s.unregister:
			s.handleUnregister(id)
		case msg := <-s.inbound:
			s.handleInbound(msg)
		case tick := <-tickerC:
			if len(s.players) >= minimumPlayersToStart && !s.started {
				s.broadcastMessage(fmt.Sprintf(gameTimerFmt, len(s.players), int(end.Sub(tick).Seconds())))

				if time.Since(start) >= startGameDelay {
					ticker.Stop()
					s.startGame()
				}
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

	if !s.hasPlayer(req.playerID) {
		s.joinOrder = append(s.joinOrder, req.playerID)
	}

	s.players[req.playerID] = nil
	req.result <- true
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

	s.sendCharacterSelections(req.client.id)
}

func (s *Session) handleUnregister(id string) {
	s.logger.Info("Received unregister request", "player_id", id)

	if c, ok := s.players[id]; ok && c != nil {
		close(c.send)
	}

	s.players[id] = nil

	if disprover, waiting := s.logic.GetPendingDisprover(); waiting && disprover == id {
		s.cancelDisprove(id)
	}
}

func (s *Session) handleInbound(msg inboundMessage) {
	s.logger.Info("Received inbound message request", "player_id", msg.from)

	var env envelopev1.Envelope
	if err := protojson.Unmarshal(msg.data, &env); err != nil {
		s.sendTo(msg.from, envelopev1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("malformed message"),
			}.Build(),
		}.Build())

		return
	}

	if env.WhichPayload() == envelopev1.Envelope_SelectCharacter_case {
		s.selectCharacter(msg.from, env.GetSelectCharacter())

		return
	}

	if !s.started {
		s.sendTo(msg.from, envelopev1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("error game not started"),
			}.Build(),
		}.Build())

		return
	}

	switch env.WhichPayload() {
	case envelopev1.Envelope_ViewHand_case:
		s.viewHand(msg.from)

		return

	case envelopev1.Envelope_ViewLocations_case:
		s.viewLocations(msg.from)

		return
	case envelopev1.Envelope_SuggestionResponse_case:
		s.disprove(msg.from, env.GetSuggestionResponse())

		return
	default:
	}

	if msg.from != s.logic.GetCurrentPlayer() {
		s.sendTo(msg.from, envelopev1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new("not your turn"),
			}.Build(),
		}.Build())

		return
	}

	if s.rejectWhileDisproving(msg.from) {
		return
	}

	s.handlePlayerTurnAction(&env)
}

func (s *Session) handlePlayerTurnAction(env *envelopev1.Envelope) {
	if env == nil {
		return
	}

	switch env.WhichPayload() {
	case envelopev1.Envelope_Move_case:
		s.move(env.GetMove())

		return

	case envelopev1.Envelope_Suggest_case:
		s.suggest(env.GetSuggest())

		return

	case envelopev1.Envelope_Accuse_case:
		s.accuse(env.GetAccuse())

		return

	case envelopev1.Envelope_EndTurn_case:
		s.endTurn()

		return
	}
}

func (s *Session) viewHand(playerID string) {
	playerHand := s.logic.GetPlayerHand(playerID)
	s.sendMessageTo(playerID, fmt.Sprintf(playerHandFmt, playerHand.GetCharacters(), playerHand.GetRooms(), playerHand.GetWeapons()))
}

func (s *Session) viewLocations(playerID string) {
	for _, id := range s.joinOrder {
		s.sendMessageTo(playerID, s.getPlayerLocationMessage(id))
	}
}

func (s *Session) sendMessageTo(id, msg string) {
	s.sendTo(id, envelopev1.Envelope_builder{
		Msg: sessionv1.Message_builder{
			Msg: new(msg),
		}.Build(),
	}.Build())
}

func (s *Session) hasLiveClient(id string) bool {
	c, ok := s.players[id]

	return ok && c != nil
}

func (s *Session) sendTo(id string, env *envelopev1.Envelope) {
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

func (s *Session) startGame() {
	s.logger.Info("Starting game", "players", len(s.players))
	s.started = true
	s.sendGameStartMessage()
	s.broadcastMessage("game is starting")
	s.logic.InitializeGameState(s.joinOrder)
	s.broadcastCharacterSelections()

	for _, playerID := range s.joinOrder {
		playerHand := s.logic.GetPlayerHand(playerID)
		playerHandMessage := fmt.Sprintf(playerHandFmt, playerHand.GetCharacters(), playerHand.GetRooms(), playerHand.GetWeapons())
		s.sendTo(playerID, envelopev1.Envelope_builder{
			Msg: sessionv1.Message_builder{
				Msg: new(playerHandMessage),
			}.Build(),
		}.Build())

		playerLocationMessage := s.getPlayerLocationMessage(playerID)
		s.sendTo(playerID, envelopev1.Envelope_builder{
			Msg: sessionv1.Message_builder{
				Msg: new(playerLocationMessage),
			}.Build(),
		}.Build())
	}

	turnMessage := s.logic.GetTurnMessage()
	s.broadcastMessage(turnMessage)
}

func (s *Session) sendGameStartMessage() {
	sessionID := s.id.String()
	startGame := envelopev1.Envelope_builder{
		StartGame: sessionv1.StartGame_builder{
			SessionId: &sessionID,
		}.Build(),
	}.Build()
	s.broadcast(startGame)
}

func (s *Session) move(move *sessionv1.Move) {
	moveMessage, err := s.logic.Move(move)
	if err != nil {
		s.sendError(move.GetPlayerId(), err.Error())
		return
	}

	s.broadcastMessage(moveMessage)
}

// suggest handles suggestion requests.
func (s *Session) suggest(suggestion *suggestionv1.SuggestionRequest) {
	result, err := s.logic.Suggest(suggestion)
	if err != nil {
		s.logger.Info("Rejected suggestion", "player_id", suggestion.GetPlayerId(), "reason", err.Error())
		s.sendError(suggestion.GetPlayerId(), err.Error())

		return
	}

	s.broadcastMessage(result.Announcement)

	if result.Moved != "" {
		s.broadcastMessage(result.Moved)
	}

	s.broadcastMessage(result.Resolution)

	// Nobody holds a suggested card, so the suggester hears about it right away.
	if result.DisproverID == "" {
		s.sendMessageTo(suggestion.GetPlayerId(), suggestNoneFmt)

		return
	}

	s.requestDisprove(result.DisproverID, suggestion)
}

func (s *Session) sendError(playerID, message string) {
	s.sendTo(playerID, envelopev1.Envelope_builder{
		Error: sessionv1.Error_builder{
			Error: new(message),
		}.Build(),
	}.Build())
}

// processes player's accusation and broadcasts the result.
func (s *Session) accuse(accusation *suggestionv1.AccusationRequest) {
	// Get playerID of the player who made the accusation.
	playerID := accusation.GetSuggestion().GetPlayerId()

	result, err := s.logic.Accuse(accusation)
	if err != nil {
		s.sendTo(playerID, envelopev1.Envelope_builder{
			Error: sessionv1.Error_builder{
				Error: new(err.Error()),
			}.Build(),
		}.Build())

		return
	}

	if result.Announcement != "" {
		s.broadcastMessage(result.Announcement)
	}

	if result.IsCorrect {
		s.broadcastMessage(fmt.Sprintf(
			"Player %s Wins! %s",
			playerID,
			s.logic.GetCaseFile(),
		))

		s.logic.EndGame()

		return
	}

	s.broadcastMessage(fmt.Sprintf(
		"Player %s has made an incorrect accusation and has been eliminated.",
		playerID,
	))

	if s.logic.PlayersRemaining() == 0 {
		s.broadcastMessage("All players have been eliminated. Game Over!")

		s.logic.EndGame()

		return
	}

	s.endTurn()
}

func (s *Session) endTurn() {
	s.logic.AdvanceTurn()

	turnMessage := s.logic.GetTurnMessage()
	s.broadcastMessage(turnMessage)
}

func (s *Session) getPlayerLocationMessage(playerID string) string {
	playerLocation := s.logic.GetPlayerLocation(playerID)

	if playerLocation == nil {
		return ""
	}

	if playerLocation.HasRoom() {
		return fmt.Sprintf(playerLocationFmt, playerID, playerLocation.GetRoom().String())
	}

	return fmt.Sprintf(playerLocationFmt, playerID, playerLocation.GetHall().String())
}

func (s *Session) broadcastMessage(msg string) {
	s.logger.Info("Broadcasting message")

	s.broadcast(envelopev1.Envelope_builder{
		Msg: sessionv1.Message_builder{
			Msg: new(msg),
		}.Build(),
	}.Build())
}

func (s *Session) broadcast(env *envelopev1.Envelope) {
	for id := range s.players {
		s.sendTo(id, env)
	}
}
