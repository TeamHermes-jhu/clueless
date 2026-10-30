package session

import (
	"fmt"
	"math/rand/v2"

	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

const (
	suggestFmt = "player %s has suggested: %s with the %s in the %s"
	accuseFmt  = "player %s has made an accusation: %s with the %s in the %s"
	moveFmt    = "player %s has moved to %s"
)

func (s *Session) startGame() {
	s.logger.Info("Starting game", "players", len(s.players))
	s.started = true

	s.broadcastMessage("game is starting")

	r := rand.Perm(len(s.players)) //nolint:gosec

	s.turnOrder = make([]string, len(s.joinOrder))
	for i, v := range r {
		s.turnOrder[v] = s.joinOrder[i]
	}
}

func (s *Session) move(move *sessionv1.Move) {
	// TODO: validate its a good move
	to := move.GetRoom().String()

	if move.WhichMoveTo() == sessionv1.Move_Hall_case {
		to = move.GetRoom().String()
	}

	s.broadcastMessage(fmt.Sprintf(moveFmt, move.GetPlayerId(), to))
}

func (s *Session) suggest(suggestion *suggestionv1.SuggestionRequest) {
	// TODO: figure out how to handle suggestions
	s.broadcastMessage(fmt.Sprintf(
		suggestFmt,
		suggestion.GetPlayerId(),
		suggestion.GetCharacter().String(),
		suggestion.GetWeapon().String(),
		suggestion.GetRoom().String(),
	))
}

func (s *Session) accuse(accusation *suggestionv1.AccusationRequest) {
	// TODO: validation + logic
	s.broadcastMessage(fmt.Sprintf(
		accuseFmt,
		accusation.GetSuggestion().GetPlayerId(),
		accusation.GetSuggestion().GetCharacter().String(),
		accusation.GetSuggestion().GetWeapon().String(),
		accusation.GetSuggestion().GetRoom().String(),
	))
}
