package session

import (
	"fmt"

	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

const (
	// disproveRequestFmt asks a player to show a card that disproves a suggestion.
	disproveRequestFmt = "player %s suggested %s with the %s in the %s. Show a card to disprove it"
	// disproveGoneFmt tells the suggester that the asked player cannot answer.
	disproveGoneFmt = "%s is not connected and cannot disprove your suggestion"
)

// requestDisprove asks one player to show a card that disproves a suggestion, and tells the
// suggester who was asked. The suggestion itself is sent so the asked player knows what to
// answer about, and their SuggestionResponse comes back to disprove.
//
// A disconnected player can never answer, so the request is cancelled instead of leaving the
// game waiting on them forever.
func (s *Session) requestDisprove(disproverID string, suggestion *suggestionv1.SuggestionRequest) {
	if !s.hasLiveClient(disproverID) {
		s.cancelDisprove(disproverID)

		return
	}

	s.sendMessageTo(suggestion.GetPlayerId(), fmt.Sprintf(suggestPendingFmt, disproverID))

	s.sendMessageTo(disproverID, fmt.Sprintf(
		disproveRequestFmt,
		suggestion.GetPlayerId(),
		suggestion.GetCharacter().String(),
		suggestion.GetWeapon().String(),
		suggestion.GetRoom().String(),
	))

	s.sendTo(disproverID, envelopev1.Envelope_builder{
		Suggest: suggestion,
	}.Build())
}

// cancelDisprove drops the pending disprove request and tells the suggester that the asked
// player is gone. It does nothing when no request is pending.
func (s *Session) cancelDisprove(disproverID string) {
	suggesterID, cancelled := s.logic.CancelDisprove()
	if !cancelled {
		return
	}

	s.logger.Info("Cancelled disprove request", "disprover_id", disproverID)
	s.sendMessageTo(suggesterID, fmt.Sprintf(disproveGoneFmt, disproverID))
}

// disprove handles a player's answer to a disprove request.
func (s *Session) disprove(playerID string, resp *suggestionv1.SuggestionResponse) {
	result, err := s.logic.Disprove(playerID, resp)
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
	disprover, waiting := s.logic.GetPendingDisprover()
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
