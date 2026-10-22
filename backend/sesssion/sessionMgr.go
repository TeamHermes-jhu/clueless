
package session

import (
	"github.com/TeamHermes-jhu/clueless/backend/cmd"
)

func (sl *SessionList) makeSession(id int) Session {
	// takes in a session id
	// add session to the SessionList
	// make sure that the id is unique in the session list
	// returns an empty session with no players, status waiting, constant min/max
}

func (s *Session) addPlayer(p *Player) {
	// takes in a pointer to a session and a pointer to a player
	// make sure the player is not already in the session
	// modifies the session in place to add the user, if there is a spot available
}

func (s *Session) removePlayer(p *Player) {
	// takes in a pointer to a session and a pointer to a player
	// modifies the session in place to remove the user, if present
}

func (s *Session) beginGame() {
	// if there are enough players, switch game status to active
	// pass control to GameLogicManager maybe?
}

func (sl *SessionList) endSession(id int) {
	// end a session, by removing it from the list of active sessions
}