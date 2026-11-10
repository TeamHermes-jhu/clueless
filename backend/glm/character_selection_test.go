package glm //nolint:testpackage // tests the unexported selection helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
)

func TestSelectCharacter_AssignsFreeCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))

	got, ok := g.GetPlayerCharacter("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PLUM, got)
}

func TestSelectCharacter_RejectsTakenCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))
	require.ErrorIs(t, g.SelectCharacter("bob", cardsv1.Characters_PLUM), ErrCharacterTaken)

	got, ok := g.GetPlayerCharacter("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PLUM, got)

	_, ok = g.GetPlayerCharacter("bob")
	assert.False(t, ok)
}

func TestSelectCharacter_SameCharacterAgainIsNoOp(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))
	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))

	assert.Len(t, g.GetCharacterSelections().GetAssignments(), 1)
}

func TestSelectCharacter_ChangingFreesPreviousCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))
	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_GREEN))
	require.NoError(t, g.SelectCharacter("bob", cardsv1.Characters_PLUM))

	got, ok := g.GetPlayerCharacter("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_GREEN, got)
}

func TestSelectCharacter_RejectsInvalidCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.ErrorIs(t, g.SelectCharacter("alice", cardsv1.Characters(0)), ErrInvalidCharacter)
	require.ErrorIs(t, g.SelectCharacter("alice", cardsv1.Characters(7)), ErrInvalidCharacter)
	require.ErrorIs(t, g.SelectCharacter("alice", cardsv1.Characters(999)), ErrInvalidCharacter)

	_, ok := g.GetPlayerCharacter("alice")
	assert.False(t, ok)
}

func TestReleaseCharacter_FreesCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))

	g.ReleaseCharacter("alice")

	_, ok := g.GetPlayerCharacter("alice")
	assert.False(t, ok)
	require.NoError(t, g.SelectCharacter("bob", cardsv1.Characters_PLUM))
}

func TestReleaseCharacter_UnknownPlayerDoesNothing(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_PLUM))

	g.ReleaseCharacter("nobody")

	assert.Len(t, g.GetCharacterSelections().GetAssignments(), 1)
}

func TestGetCharacterSelections_EmptyByDefault(t *testing.T) {
	g := &GameLogicManager{}

	assert.Empty(t, g.GetCharacterSelections().GetAssignments())
}

func TestGetCharacterSelections_OrderedByCharacter(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("bob", cardsv1.Characters_PLUM))
	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_WHITE))

	got := g.GetCharacterSelections().GetAssignments()

	require.Len(t, got, 2)
	assert.Equal(t, "alice", got[0].GetPlayerId())
	assert.Equal(t, cardsv1.Characters_WHITE, got[0].GetCharacter())
	assert.Equal(t, "bob", got[1].GetPlayerId())
	assert.Equal(t, cardsv1.Characters_PLUM, got[1].GetCharacter())
}

func TestAssignRemainingCharacters_FillsGapsInOrder(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_WHITE))

	g.assignRemainingCharacters([]string{"alice", "bob", "carol"})

	bob, ok := g.GetPlayerCharacter("bob")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_GREEN, bob)

	carol, ok := g.GetPlayerCharacter("carol")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PEACOCK, carol)
}

func TestInitializeGameState_KeepsChoicesAndAssignsTheRest(t *testing.T) {
	g := &GameLogicManager{}
	players := []string{"alice", "bob", "carol"}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_MUSTARD))

	g.InitializeGameState(players)

	seen := make(map[cardsv1.Characters]bool)

	for _, playerID := range players {
		got, ok := g.GetPlayerCharacter(playerID)
		require.True(t, ok, "player %s has no character", playerID)

		seen[got] = true
	}

	assert.Len(t, seen, len(players))

	alice, _ := g.GetPlayerCharacter("alice")
	assert.Equal(t, cardsv1.Characters_MUSTARD, alice)
}
