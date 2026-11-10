package glm

import (
	"cmp"
	"errors"
	"slices"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
)

var (
	// ErrInvalidCharacter is returned when the requested character does not exist.
	ErrInvalidCharacter = errors.New("invalid character")
	// ErrCharacterTaken is returned when another player already holds the requested character.
	ErrCharacterTaken = errors.New("character already taken")
)

// SelectCharacter lets a player claim a character.
// Returns ErrInvalidCharacter or ErrCharacterTaken if the choice is rejected.
func (g *GameLogicManager) SelectCharacter(playerID string, character cardsv1.Characters) error {
	if !validCharacter(character) {
		return ErrInvalidCharacter
	}

	owner, taken := g.characterOwner(character)
	if taken && owner == playerID {
		return nil
	}

	if taken {
		return ErrCharacterTaken
	}

	g.ensurePlayerState(playerID).SetCharacter(character)

	return nil
}

// ReleaseCharacter frees the character a player holds. It does nothing if they hold none.
func (g *GameLogicManager) ReleaseCharacter(playerID string) {
	state := g.getPlayerState(playerID)
	if state == nil {
		return
	}

	state.ClearCharacter()
}

// GetPlayerCharacter returns the character a player holds, if any.
func (g *GameLogicManager) GetPlayerCharacter(playerID string) (cardsv1.Characters, bool) {
	state := g.getPlayerState(playerID)
	if state == nil || !state.HasCharacter() {
		return 0, false
	}

	return state.GetCharacter(), true
}

// GetCharacterSelections returns every current selection, ordered by character.
func (g *GameLogicManager) GetCharacterSelections() *glmv1.CharacterSelectChoices {
	choices := make([]*glmv1.CharacterSelectResponse, 0, len(g.playerStates))

	for playerID, state := range g.playerStates {
		if !state.HasCharacter() {
			continue
		}

		choices = append(choices, glmv1.CharacterSelectResponse_builder{
			PlayerId:  new(playerID),
			Character: new(state.GetCharacter()),
		}.Build())
	}

	slices.SortFunc(choices, func(a, b *glmv1.CharacterSelectResponse) int {
		return cmp.Compare(a.GetCharacter(), b.GetCharacter())
	})

	return glmv1.CharacterSelectChoices_builder{
		Assignments: choices,
	}.Build()
}

// assignRemainingCharacters gives every player without a character the lowest numbered
// free one, in the order the players are listed. It runs when the game starts.
func (g *GameLogicManager) assignRemainingCharacters(playerIDs []string) {
	for _, playerID := range playerIDs {
		state := g.ensurePlayerState(playerID)
		if state.HasCharacter() {
			continue
		}

		for c := cardsv1.Characters(1); c <= numSuspects; c++ {
			if _, taken := g.characterOwner(c); !taken {
				state.SetCharacter(c)

				break
			}
		}
	}
}

// characterOwner returns the player who holds character, if any.
func (g *GameLogicManager) characterOwner(character cardsv1.Characters) (string, bool) {
	for playerID, state := range g.playerStates {
		if state.HasCharacter() && state.GetCharacter() == character {
			return playerID, true
		}
	}

	return "", false
}

// ensurePlayerState returns the player's state, creating it if it does not exist yet.
func (g *GameLogicManager) ensurePlayerState(playerID string) *glmv1.PlayerGameState {
	if g.playerStates == nil {
		g.playerStates = make(map[string]*glmv1.PlayerGameState)
	}

	state, ok := g.playerStates[playerID]
	if !ok {
		state = glmv1.PlayerGameState_builder{}.Build()
		g.playerStates[playerID] = state
	}

	return state
}

// validCharacter reports whether c is a real character (not unspecified or unknown).
func validCharacter(c cardsv1.Characters) bool {
	return c >= 1 && c <= numSuspects
}

// getPlayerState returns the player's state, or nil if they have none.
func (g *GameLogicManager) getPlayerState(playerID string) *glmv1.PlayerGameState {
	return g.playerStates[playerID]
}
