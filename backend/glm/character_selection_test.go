package glm //nolint:testpackage // tests the unexported selection logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
)

func TestChoose_AssignsFreeCharacter(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))

	got, ok := cs.characterOf("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PLUM, got)
}

func TestChoose_RejectsTakenCharacter(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))

	err := cs.choose("bob", cardsv1.Characters_PLUM)
	require.ErrorIs(t, err, ErrCharacterTaken)

	got, ok := cs.characterOf("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PLUM, got)

	_, ok = cs.characterOf("bob")
	assert.False(t, ok)
}

func TestChoose_SameCharacterAgainIsNoOp(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))
	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))

	assert.Len(t, cs.assignments(), 1)
}

func TestChoose_ChangingReleasesPreviousCharacter(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))
	require.NoError(t, cs.choose("alice", cardsv1.Characters_GREEN))

	require.NoError(t, cs.choose("bob", cardsv1.Characters_PLUM))

	got, ok := cs.characterOf("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_GREEN, got)
}

func TestChoose_RejectsInvalidCharacter(t *testing.T) {
	var cs characterSelection

	require.ErrorIs(t, cs.choose("alice", cardsv1.Characters(0)), ErrInvalidCharacter)
	require.ErrorIs(t, cs.choose("alice", cardsv1.Characters(7)), ErrInvalidCharacter)
	require.ErrorIs(t, cs.choose("alice", cardsv1.Characters(999)), ErrInvalidCharacter)

	_, ok := cs.characterOf("alice")
	assert.False(t, ok)
}

func TestRelease_FreesCharacter(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))

	cs.release("alice")

	_, ok := cs.characterOf("alice")
	assert.False(t, ok)
	require.NoError(t, cs.choose("bob", cardsv1.Characters_PLUM))
}

func TestRelease_UnknownPlayerDoesNothing(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_PLUM))

	cs.release("nobody")

	assert.Len(t, cs.assignments(), 1)
}

func TestAssignments_OrderedByCharacter(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("bob", cardsv1.Characters_PLUM))
	require.NoError(t, cs.choose("alice", cardsv1.Characters_WHITE))

	got := cs.assignments()

	require.Len(t, got, 2)
	assert.Equal(t, "alice", got[0].playerID)
	assert.Equal(t, "bob", got[1].playerID)
}

func TestToProto_ListsAssignments(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("bob", cardsv1.Characters_PLUM))
	require.NoError(t, cs.choose("alice", cardsv1.Characters_WHITE))

	got := cs.toProto().GetAssignments()

	require.Len(t, got, 2)
	assert.Equal(t, "alice", got[0].GetPlayerId())
	assert.Equal(t, cardsv1.Characters_WHITE, got[0].GetCharacter())
	assert.Equal(t, "bob", got[1].GetPlayerId())
	assert.Equal(t, cardsv1.Characters_PLUM, got[1].GetCharacter())
}

func TestAssignRemaining_FillsGapsInOrder(t *testing.T) {
	var cs characterSelection

	require.NoError(t, cs.choose("alice", cardsv1.Characters_WHITE))

	cs.assignRemaining([]string{"alice", "bob", "carol"})

	bob, ok := cs.characterOf("bob")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_GREEN, bob)

	carol, ok := cs.characterOf("carol")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_PEACOCK, carol)

	alice, ok := cs.characterOf("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_WHITE, alice)
}

func TestSelectCharacter_WorksOnZeroValueManager(t *testing.T) {
	g := &GameLogicManager{}

	require.NoError(t, g.SelectCharacter("alice", cardsv1.Characters_SCARLET))
	require.ErrorIs(t, g.SelectCharacter("bob", cardsv1.Characters_SCARLET), ErrCharacterTaken)

	got, ok := g.GetPlayerCharacter("alice")
	require.True(t, ok)
	assert.Equal(t, cardsv1.Characters_SCARLET, got)
}

func TestGetCharacterSelections_EmptyByDefault(t *testing.T) {
	g := &GameLogicManager{}

	assert.Empty(t, g.GetCharacterSelections().GetAssignments())
}

func TestInitializeGameState_GivesEveryPlayerACharacter(t *testing.T) {
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
