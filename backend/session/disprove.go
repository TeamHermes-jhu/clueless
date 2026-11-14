package session

import (
	"fmt"

	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

// disprove handles a player's answer to a disprove request.
func (s *Session) disprove(playerID string, resp *suggestionv1.SuggestionResponse) {
	result, err := s.gameLogicManager.Disprove(playerID, resp)
	if err != nil {
		s.sendDisproveError(playerID, err.Error())

		return
	}

	// Only the suggester sees the card. Everyone else only sees who disproved it.
	s.sendTo(result.SuggesterID, envelopev1.Envelope_builder{
		SuggestionResponse: result.Reveal,
	}.Build())

	s.broadcastMessage(result.Message)
}

// rejectWhileDisproving refuses a turn action while a player still has to show a card.
// It reports whether it refused.
func (s *Session) rejectWhileDisproving(playerID string) bool {
	disprover, waiting := s.gameLogicManager.GetPendingDisprover()
	if !waiting {
		return false
	}

	s.sendDisproveError(playerID, fmt.Sprintf("waiting for %s to show a card", disprover))

	return true
}

// sendDisproveError tells one player why their message was refused.
func (s *Session) sendDisproveError(playerID, reason string) {
	s.sendTo(playerID, envelopev1.Envelope_builder{
		Error: sessionv1.Error_builder{
			Error: new(reason),
		}.Build(),
	}.Build())
}
