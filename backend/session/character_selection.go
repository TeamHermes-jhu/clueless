package session

import (
	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

// selectCharacter handles a player's request to claim a character.
// It is only allowed before the game starts and is not turn based.
func (s *Session) selectCharacter(playerID string, req *glmv1.CharacterSelectRequest) {
	if s.started {
		s.rejectSelection(playerID, "game already started")

		return
	}

	if err := s.gameLogicManager.SelectCharacter(playerID, req.GetCharacter()); err != nil {
		s.rejectSelection(playerID, err.Error())

		return
	}

	s.broadcastCharacterSelections()
}

// rejectSelection tells one player that their request was refused.
func (s *Session) rejectSelection(playerID, reason string) {
	s.sendTo(playerID, envelopev1.Envelope_builder{
		Error: sessionv1.Error_builder{
			Error: new(reason),
		}.Build(),
	}.Build())
}

// broadcastCharacterSelections sends every player the current selections.
func (s *Session) broadcastCharacterSelections() {
	env := envelopev1.Envelope_builder{
		CharacterSelection: s.gameLogicManager.GetCharacterSelections(),
	}.Build()

	for id := range s.players {
		s.sendTo(id, env)
	}
}

// sendCharacterSelections sends one player the current selections, for example when they
// connect. Nothing is sent while no character has been chosen yet.
func (s *Session) sendCharacterSelections(playerID string) {
	update := s.gameLogicManager.GetCharacterSelections()
	if len(update.GetAssignments()) == 0 {
		return
	}

	s.sendTo(playerID, envelopev1.Envelope_builder{
		CharacterSelection: update,
	}.Build())
}
