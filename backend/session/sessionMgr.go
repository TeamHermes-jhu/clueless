// Package session provides functions for managing game sessions, including creating sessions, adding and removing players,
// starting games, and ending sessions.
package session

import (
	"errors"

	"github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func makeSession(sl *SessionList, id int32) (*SessionList, *Session, error) {
	errorNotUnique := errors.New("session ID is not unique")

	// make sure that the id is unique in the session list
	for _, otherSession := range sl.GetSessions() {
		if otherSession.GetSessionId() == id {
			// return an error if the id is not unique
			return sl, nil, errorNotUnique
		}
	}

	s := new(Session)

	const minPlayers = 3

	const maxPlayers = 6

	// instantiate session with no players, status waiting, constant min/max
	s.SetSessionId(id)
	s.SetStatus(SessionStatus_WAITING)
	s.SetPlayers([]*Player{})
	s.SetMinPlayers(minPlayers)
	s.SetMaxPlayers(maxPlayers)

	// add session to the SessionList
	sl.SetSessions(append(sl.GetSessions(), s))

	// return the new session
	return sl, s, nil
}

func addPlayer(s *Session, p *Player) error {
	// takes in a pointer to a session and a pointer to a player
	errorAlreadyInSession := errors.New("player is already in the session")
	errorSessionFull := errors.New("session is full, cannot add player")

	// make sure the player is not already in the session
	for _, otherPlayer := range s.GetPlayers() {
		if otherPlayer.GetPlayerId() == p.GetPlayerId() {
			// return an error if the player is already in the session
			return errorAlreadyInSession
		}
	}

	// modifies the session in place to add the user, if there is a spot available
	if len(s.GetPlayers()) < int(s.GetMaxPlayers()) {
		s.SetPlayers(append(s.GetPlayers(), p))
		return nil
	}

	return errorSessionFull
}

func removePlayer(s *Session, p *Player) error {
	// takes in a pointer to a session and a pointer to a player
	// modifies the session in place to remove the user, if present
	errorNotInSession := errors.New("player is not in the session")

	for i, otherPlayer := range s.GetPlayers() {
		if otherPlayer.GetPlayerId() == p.GetPlayerId() {
			// remove the player from the session
			s.SetPlayers(append(s.GetPlayers()[:i], s.GetPlayers()[i+1:]...))
			return nil
		}
	}

	return errorNotInSession
}

func beginGame(s *Session) error {
	// if there are enough players, switch game status to active
	errorNotEnoughPlayers := errors.New("not enough players to start the game")

	if len(s.GetPlayers()) >= int(s.GetMinPlayers()) {
		s.SetStatus(SessionStatus_ACTIVE)
		// instantiate the game state here? or does the server do that?
		return nil
	}
	// not enough players to start the game
	s.SetStatus(SessionStatus_WAITING)

	return errorNotEnoughPlayers
}

func endSession(sl *SessionList, id int32) error {
	// end a session, by removing it from the list of active sessions
	errorNotFound := errors.New("session ID not found")

	for i, s := range sl.GetSessions() {
		if s.GetSessionId() == id {
			sl.SetSessions(append(sl.GetSessions()[:i], sl.GetSessions()[i+1:]...))
			return nil
		}
	}

	return errorNotFound
}
