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

// SelectCharacter lets a player claim a character. Choosing a different character frees the
// one the player held before. Choosing the character you already hold succeeds.
// It returns ErrInvalidCharacter or ErrCharacterTaken if the choice is rejected.
func (g *GameLogicManager) SelectCharacter(playerID string, character cardsv1.Characters) error {
	return g.characters.choose(playerID, character)
}

// ReleaseCharacter frees the character a player holds. It does nothing if they hold none.
func (g *GameLogicManager) ReleaseCharacter(playerID string) {
	g.characters.release(playerID)
}

// GetPlayerCharacter returns the character a player holds, if any.
func (g *GameLogicManager) GetPlayerCharacter(playerID string) (cardsv1.Characters, bool) {
	return g.characters.characterOf(playerID)
}

// GetCharacterSelections returns every current selection as the message sent to players.
func (g *GameLogicManager) GetCharacterSelections() *glmv1.CharacterSelectionUpdate {
	return g.characters.toProto()
}

// assignRemainingCharacters gives every player without a character the lowest numbered
// free one. It runs when the game starts so that every player has a character.
func (g *GameLogicManager) assignRemainingCharacters(playerIDs []string) {
	g.characters.assignRemaining(playerIDs)
}

// assignment is one player's chosen character.
type assignment struct {
	playerID  string
	character cardsv1.Characters
}

// characterSelection tracks which player has claimed which character.
// The zero value is ready to use. Not safe for concurrent use: only touch it
// from the session's run goroutine.
type characterSelection struct {
	byPlayer    map[string]cardsv1.Characters
	byCharacter map[cardsv1.Characters]string
}

// choose gives character to the player and releases any character they held before.
func (cs *characterSelection) choose(playerID string, character cardsv1.Characters) error {
	if !validCharacter(character) {
		return ErrInvalidCharacter
	}

	owner, taken := cs.byCharacter[character]
	if taken && owner == playerID {
		return nil
	}

	if taken {
		return ErrCharacterTaken
	}

	cs.release(playerID)

	if cs.byPlayer == nil {
		cs.byPlayer = make(map[string]cardsv1.Characters)
		cs.byCharacter = make(map[cardsv1.Characters]string)
	}

	cs.byPlayer[playerID] = character
	cs.byCharacter[character] = playerID

	return nil
}

// release frees the character the player holds. It does nothing if they hold none.
func (cs *characterSelection) release(playerID string) {
	character, ok := cs.byPlayer[playerID]
	if !ok {
		return
	}

	delete(cs.byPlayer, playerID)
	delete(cs.byCharacter, character)
}

// characterOf returns the character the player holds, if any.
func (cs *characterSelection) characterOf(playerID string) (cardsv1.Characters, bool) {
	character, ok := cs.byPlayer[playerID]

	return character, ok
}

// freeCharacters returns the characters nobody holds, lowest numbered first.
func (cs *characterSelection) freeCharacters() []cardsv1.Characters {
	free := make([]cardsv1.Characters, 0, numSuspects)

	for c := cardsv1.Characters(1); c <= numSuspects; c++ {
		if _, taken := cs.byCharacter[c]; !taken {
			free = append(free, c)
		}
	}

	return free
}

// assignRemaining gives each player in playerIDs who holds no character the next free one,
// in the order the players are listed.
func (cs *characterSelection) assignRemaining(playerIDs []string) {
	free := cs.freeCharacters()

	for _, playerID := range playerIDs {
		if _, ok := cs.byPlayer[playerID]; ok {
			continue
		}

		if len(free) == 0 {
			return
		}

		// Cannot fail: free[0] is a valid character that nobody holds.
		_ = cs.choose(playerID, free[0])
		free = free[1:]
	}
}

// assignments returns every current selection, ordered by character so the result is stable.
func (cs *characterSelection) assignments() []assignment {
	out := make([]assignment, 0, len(cs.byPlayer))

	for playerID, character := range cs.byPlayer {
		out = append(out, assignment{playerID: playerID, character: character})
	}

	slices.SortFunc(out, func(a, b assignment) int {
		return cmp.Compare(a.character, b.character)
	})

	return out
}

// toProto builds the message that is broadcast to players.
func (cs *characterSelection) toProto() *glmv1.CharacterSelectionUpdate {
	current := cs.assignments()
	assignments := make([]*glmv1.CharacterAssignment, 0, len(current))

	for _, a := range current {
		assignments = append(assignments, glmv1.CharacterAssignment_builder{
			PlayerId:  new(a.playerID),
			Character: new(a.character),
		}.Build())
	}

	return glmv1.CharacterSelectionUpdate_builder{
		Assignments: assignments,
	}.Build()
}

// validCharacter reports whether c is a real character (not unspecified or unknown).
func validCharacter(c cardsv1.Characters) bool {
	return c >= 1 && c <= numSuspects
}
