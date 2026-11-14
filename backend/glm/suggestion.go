package glm

import (
	"errors"
	"fmt"
	"slices"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

// Rule violations reported back to the player that made the suggestion.
var (
	ErrMalformedSuggestion = errors.New("a suggestion needs a character, a weapon and a room")
	ErrUnknownPlayer       = errors.New("you are not part of this game")
	ErrNotYourTurn         = errors.New("it is not your turn")
	ErrAlreadySuggested    = errors.New("you have already made a suggestion this turn")
	ErrWrongRoom           = errors.New("you can only suggest the room you are currently in")
)

const (
	suggestFmt           = "player %s has suggested: %s with the %s in the %s"
	suggestMovedFmt      = "%s was moved to the %s by the suggestion"
	notDisprovedFmt      = "Not disproved by anyone"
	disproveRequestedFmt = "%s is being asked to disprove %s's suggestion"
)

// SuggestionResult holds everything the session needs to report a suggestion back to the players.
type SuggestionResult struct {
	Announcement string
	Moved        string
	Resolution   string
	DisproverID  string
}

// Suggest validates a suggestion, moves the suggested character into the room and looks for a
// player able to disprove it.
func (g *GameLogicManager) Suggest(suggestion *suggestionv1.SuggestionRequest) (*SuggestionResult, error) {
	if err := g.validateSuggestion(suggestion); err != nil {
		return nil, err
	}

	playerID := suggestion.GetPlayerId()
	g.hasSuggested = true

	result := &SuggestionResult{
		Announcement: fmt.Sprintf(
			suggestFmt,
			playerID,
			suggestion.GetCharacter().String(),
			suggestion.GetWeapon().String(),
			suggestion.GetRoom().String(),
		),
		Moved: g.moveFromSuggestion(suggestion.GetCharacter(), suggestion.GetRoom()),
	}

	disproverID := g.findDisprover(suggestion)
	if disproverID == "" {
		result.Resolution = notDisprovedFmt

		return result, nil
	}

	result.DisproverID = disproverID
	result.Resolution = fmt.Sprintf(disproveRequestedFmt, disproverID, playerID)

	// TODO: replace the stub with the real disprove request/response exchange.
	g.disproveRequest(disproverID, suggestion)

	return result, nil
}

// validateSuggestion reports whether the suggestion follows the rules of the game.
func (g *GameLogicManager) validateSuggestion(suggestion *suggestionv1.SuggestionRequest) error {
	if suggestion.GetCharacter() == cardsv1.Characters_UNSPECIFIED ||
		suggestion.GetWeapon() == cardsv1.Weapons_UNSPECIFIED ||
		suggestion.GetRoom() == cardsv1.Rooms_UNSPECIFIED {
		return ErrMalformedSuggestion
	}

	if err := g.checkIsPlayerAllowed(suggestion.GetPlayerId()); err != nil {
		return err
	}

	location := g.GetPlayerLocation(suggestion.GetPlayerId())
	if !location.HasRoom() || location.GetRoom() != suggestion.GetRoom() {
		return ErrWrongRoom
	}

	return nil
}

// checkIsPlayerAllowed reports whether the player may act right now.
func (g *GameLogicManager) checkIsPlayerAllowed(playerID string) error {
	if _, ok := g.playerStates[playerID]; !ok {
		return ErrUnknownPlayer
	}

	if playerID != g.GetCurrentPlayer() {
		return ErrNotYourTurn
	}

	if g.hasSuggested {
		return ErrAlreadySuggested
	}

	return nil
}

// moveFromSuggestion pulls the suggested character into the suggested room and returns a message
// describing the move. It returns an empty string when no player controls that character.
func (g *GameLogicManager) moveFromSuggestion(character cardsv1.Characters, room cardsv1.Rooms) string {
	state := g.stateOfCharacter(character)
	if state == nil {
		return ""
	}

	state.SetLocation(glmv1.PlayerLocation_builder{
		Room: &room,
	}.Build())

	return fmt.Sprintf(suggestMovedFmt, character.String(), room.String())
}

// stateOfCharacter returns the game state of the player controlling a character,
// or nil when no player controls it.
func (g *GameLogicManager) stateOfCharacter(character cardsv1.Characters) *glmv1.PlayerGameState {
	for _, state := range g.playerStates {
		if state.GetCharacter() == character {
			return state
		}
	}

	return nil
}

// findDisprover returns the first player after the suggesting player in the turn order that holds
// one of the suggested cards. It returns an empty string when nobody can disprove the suggestion.
func (g *GameLogicManager) findDisprover(suggestion *suggestionv1.SuggestionRequest) string {
	suggesterIdx := g.turnOrderIndex(suggestion.GetPlayerId())
	if suggesterIdx < 0 {
		return ""
	}

	for offset := 1; offset < len(g.turnOrder); offset++ {
		playerID := g.turnOrder[(suggesterIdx+offset)%len(g.turnOrder)]
		if handDisproves(g.GetPlayerHand(playerID), suggestion) {
			return playerID
		}
	}

	// call g.disprove here instead of calling handDisproves. Then let disprove reach out to the disprover

	return ""
}

// turnOrderIndex returns the position of a player in the turn order, or -1 when it is not found.
func (g *GameLogicManager) turnOrderIndex(playerID string) int {
	for i, id := range g.turnOrder {
		if id == playerID {
			return i
		}
	}

	return -1
}

// handDisproves reports whether a hand holds any of the suggested cards.
func handDisproves(hand *glmv1.PlayerHand, suggestion *suggestionv1.SuggestionRequest) bool {
	return slices.Contains(hand.GetCharacters(), suggestion.GetCharacter()) ||
		slices.Contains(hand.GetWeapons(), suggestion.GetWeapon()) ||
		slices.Contains(hand.GetRooms(), suggestion.GetRoom())
}
