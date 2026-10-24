package session

import (
	"fmt"
	. "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func makeSession(sl *SessionList, id int32) (*SessionList, *Session, error) {

	// make sure that the id is unique in the session list
	for _, otherSession := range sl.GetSessions() {
		if otherSession.GetSessionId() == id {
			// return an error if the id is not unique
			return sl, nil, fmt.Errorf("session ID %d is not unique", id)

		}
	}

	s := new(Session)

	// instantiate session with no players, status waiting, constant min/max
	s.SetSessionId(id)
	s.SetStatus(SessionStatus_WAITING)
	s.SetPlayers([]*Player{})
	s.SetMinPlayers(3)
	s.SetMaxPlayers(6)

	// add session to the SessionList
	sl.SetSessions(append(sl.GetSessions(), s))

	// return the new session
	return sl, s, nil
}

func addPlayer(s *Session, p *Player) {
	// takes in a pointer to a session and a pointer to a player
	// make sure the player is not already in the session
	// modifies the session in place to add the user, if there is a spot available
}

func removePlayer(s *Session, p *Player) {
	// takes in a pointer to a session and a pointer to a player
	// modifies the session in place to remove the user, if present
}

func beginGame(s *Session) {
	// if there are enough players, switch game status to active
	// pass control to GameLogicManager maybe?
}

func endSession(sl *SessionList, id int) {
	// end a session, by removing it from the list of active sessions
}
